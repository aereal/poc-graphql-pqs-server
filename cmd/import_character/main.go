package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aereal/poc-graphql-pqs-server/domain"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/logging"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	os.Exit(run())
}

func run() int {
	logging.Init(logging.WithOutput(os.Stdout), logging.WithDebug(os.Getenv("DEBUG") != ""))
	ctx := context.Background()
	shutdown, err := otelinstrument.Instrument(ctx, otelinstrument.WithShutdownGrace(time.Second*5), otelinstrument.WithSetGlobalTracerProvider(true))
	if err != nil {
		slog.Error("failed to instrument OpenTelemetry", slog.String("error", err.Error()))
		return 1
	}
	defer shutdown()

	db, err := infra.OpenDB(infra.WithAddr(os.Getenv("DB_ADDR")), infra.WithDBName(os.Getenv("DB_NAME")), infra.WithUser(os.Getenv("DB_USER")), infra.WithPassword(os.Getenv("DB_PASSWORD")), infra.WithSSLMode("disable"))
	if err != nil {
		slog.Error("failed to open DB", slog.String("error", err.Error()))
		return 1
	}

	if err := (&app{tracer: otel.GetTracerProvider().Tracer("import_character"), db: db}).do(ctx); err != nil {
		slog.Error("import failure", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

type app struct {
	tracer trace.Tracer
	db     *sqlx.DB
}

func (a *app) do(ctx context.Context) (err error) {
	ctx, span := a.tracer.Start(ctx, "do")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	characters, err := a.parseFile(ctx)
	if err != nil {
		return fmt.Errorf("parseFile: %w", err)
	}
	if err := a.insert(ctx, characters); err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	return nil
}

func (a *app) insert(ctx context.Context, characters []*domain.Character) (err error) {
	ctx, span := a.tracer.Start(ctx, "insert", trace.WithAttributes(attribute.Int("character_count", len(characters))))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if err := a.db.PingContext(ctx); err != nil {
		return fmt.Errorf("PingContext: %w", err)
	}

	type dto struct {
		*domain.UniqueAbility
		Name          string            `db:"name"`
		Rarelity      int               `db:"rarelity"`
		Element       domain.Element    `db:"element"`
		Health        int               `db:"health"`
		Attack        int               `db:"attack"`
		Defence       int               `db:"defence"`
		ElementEnergy int               `db:"element_energy"`
		Region        domain.Region     `db:"region"`
		WeaponKind    domain.WeaponKind `db:"weapon_kind"`
	}
	dtos := make([]dto, len(characters))
	for i, c := range characters {
		dtos[i] = dto{
			UniqueAbility: c.UniqueAbility,
			Name:          c.Name,
			Rarelity:      c.Rarelity,
			Element:       c.Element,
			Health:        c.Health,
			Attack:        c.Attack,
			Defence:       c.Defence,
			ElementEnergy: c.ElementEnergy,
			Region:        c.Region,
			WeaponKind:    c.WeaponKind,
		}
	}

	query, args, err :=
		goqu.Dialect("postgres").
			From("characters").
			Prepared(true).
			Insert().
			Rows(dtos).
			ToSQL()
	if err != nil {
		return err
	}
	if _, err := a.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ExecContext: %w", err)
	}
	return nil
}

func (a *app) parseFile(ctx context.Context) (_ []*domain.Character, err error) {
	ctx, span := a.tracer.Start(ctx, "parseFile")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	f, err := os.Open(os.Getenv("INPUT_FILE"))
	if err != nil {
		return nil, fmt.Errorf("os.Open(): %w", err)
	}
	defer func() { _ = f.Close() }()
	reader := csv.NewReader(f)
	reader.Comma = '\t'
	var characters []*domain.Character
	var idx int
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("#%d: csv.Reader.Read(): %w", idx, err)
		}
		character, err := a.parseRow(ctx, record)
		if err != nil {
			return nil, fmt.Errorf("#%d: parse(): %w", idx, err)
		}
		slog.InfoContext(ctx, "parsed character")
		characters = append(characters, character)
		idx++
	}
	span.SetAttributes(attribute.Int("character_num", len(characters)))
	return characters, nil
}

func (a *app) parseRow(ctx context.Context, record []string) (_ *domain.Character, err error) {
	ctx, span := a.tracer.Start(ctx, "parseRow")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	character := new(domain.Character)
	character.UniqueAbility = new(domain.UniqueAbility)
	if rarelity := strings.TrimFunc(record[0], func(r rune) bool { return !unicode.IsNumber(r) }); rarelity != "" {
		var err error
		character.Rarelity, err = strconv.Atoi(rarelity)
		if err != nil {
			return nil, fmt.Errorf("Rarelity: %w", err)
		}
	}
	character.Name = strings.TrimSpace(record[3])
	if el, ok := elementsMap[record[4]]; ok {
		character.Element = el
	}
	if kind, ok := weaponKind[record[5]]; ok {
		character.WeaponKind = kind
	}
	if s := record[8]; s != "" {
		switch {
		case strings.HasPrefix(s, domain.RegionMondstadt.Name()):
			character.Region = domain.RegionMondstadt
		case strings.HasPrefix(s, (domain.RegionLiyue.Name())):
			character.Region = domain.RegionLiyue
		case strings.HasPrefix(s, (domain.RegionInazuma.Name())):
			character.Region = domain.RegionInazuma
		case strings.HasPrefix(s, (domain.RegionSumeru.Name())):
			character.Region = domain.RegionSumeru
		case strings.HasPrefix(s, (domain.RegionFontaine.Name())):
			character.Region = domain.RegionFontaine
		case strings.HasPrefix(s, (domain.RegionNatlan.Name())):
			character.Region = domain.RegionNatlan
		case strings.HasPrefix(s, (domain.RegionSnezhnaya.Name())):
			character.Region = domain.RegionSnezhnaya
		default:
			slog.WarnContext(ctx, "unknown region", slog.String("region", s))
		}
	}
	if hp, err := parseNumeric(record[10]); err != nil {
		slog.WarnContext(ctx, "failed to parse HP", slog.String("error", err.Error()))
	} else {
		character.Health = hp
	}
	if attack, err := parseNumeric(record[11]); err != nil {
		slog.WarnContext(ctx, "failed to parse attack", slog.String("error", err.Error()))
	} else {
		character.Attack = attack
	}
	if defence, err := parseNumeric(record[12]); err != nil {
		slog.WarnContext(ctx, "failed to parse defence", slog.String("error", err.Error()))
	} else {
		character.Defence = defence
	}
	if v, err := parseNumeric(record[15]); err != nil {
		slog.WarnContext(ctx, "failed to parse element energy", slog.String("error", err.Error()))
	} else {
		character.ElementEnergy = v
	}
	{
		var buf bytes.Buffer
		for i, r := range record[13] {
			if i == 0 {
				continue
			}
			buf.WriteRune(r)
		}
		character.UniqueAbility.Kind = buf.String()
	}
	{
		s := record[14]
		percentageIndex := strings.IndexFunc(s, func(r rune) bool { return r == '%' || r == '％' })
		isPercentage := percentageIndex != -1
		if isPercentage {
			s = s[:percentageIndex-1]
		}
		score, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return nil, fmt.Errorf("UniqueAbilityScore: %w", err)
		}
		if isPercentage {
			score /= 100
		}
		character.UniqueAbility.Score = score
	}
	return character, nil
}

func parseNumeric(v string) (int, error) {
	var buf bytes.Buffer
	for _, r := range v {
		if !unicode.IsNumber(r) {
			continue
		}
		buf.WriteRune(r)
	}
	return strconv.Atoi(buf.String())
}

var (
	elementsMap = map[string]domain.Element{
		"炎": domain.ElementPyro,
		"水": domain.ElementHydro,
		"氷": domain.ElementCryo,
		"雷": domain.ElementElectro,
		"風": domain.ElementAnemo,
		"岩": domain.ElementGeo,
		"草": domain.ElementDendro,
	}
	weaponKind = map[string]domain.WeaponKind{
		"片手剣":  domain.WeaponKindSword,
		"両手剣":  domain.WeaponKindClaymore,
		"弓":    domain.WeaponKindBow,
		"法器":   domain.WeaponKindCatalyst,
		"長柄武器": domain.WeaponKindPolearm,
	}
)
