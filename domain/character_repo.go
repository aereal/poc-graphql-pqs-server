package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func ProvideCharacterRepository(db *sqlx.DB) (*CharacterRepository, error) {
	if db == nil {
		return nil, errors.New("sqlx.DB is required")
	}
	r := &CharacterRepository{
		db:     db,
		tracer: otel.GetTracerProvider().Tracer(pkgName + ".CharacterRepository"),
	}
	r.tables.characters = goqu.Dialect("postgres").From("characters").Prepared(true)
	return r, nil
}

type CharacterRepository struct {
	db *sqlx.DB

	tracer trace.Tracer
	tables struct{ characters *goqu.SelectDataset }
}

func (r *CharacterRepository) FindCharactersByNames(ctx context.Context, names []string) (_ map[string]*Character, err error) {
	ctx, span := r.tracer.Start(ctx, "FindCharactersByNames", trace.WithAttributes(attribute.StringSlice("app.character.name", names)))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	query, args, err := r.tables.characters.Where(goqu.C("name").In(names)).ToSQL()
	if err != nil {
		return nil, &QueryBuildError{err}
	}
	characters := make([]*Character, 0, len(names))
	if err := r.db.SelectContext(ctx, &characters, query, args...); err != nil {
		return nil, err
	}
	result := make(map[string]*Character, len(characters))
	for _, character := range characters {
		result[character.Name] = character
	}
	return result, nil
}

type ComparisonOperator string

const (
	ComparisonOperatorEqual            ComparisonOperator = "EQ"
	ComparisonOperatorLessThan         ComparisonOperator = "LT"
	ComparisonOperatorLessThanEqual    ComparisonOperator = "LTE"
	ComparisonOperatorGreaterThan      ComparisonOperator = "GT"
	ComparisonOperatorGreaterThanEqual ComparisonOperator = "GTE"
)

type NumericKind int

const (
	NumericKindUnknown NumericKind = iota
	NumericKindInt
	NumericKindUnsignedInt
	NumericKindFloat
)

type Numeric struct {
	kind             NumericKind
	intValue         int64
	unsignedIntValue uint64
	floatValue       float64
}

func (n *Numeric) value() any {
	switch n.kind {
	case NumericKindInt:
		return n.intValue
	case NumericKindUnsignedInt:
		return n.unsignedIntValue
	case NumericKindFloat:
		return n.floatValue
	default:
		return nil
	}
}

var (
	_ graphql.ContextMarshaler   = (*Numeric)(nil)
	_ graphql.ContextUnmarshaler = (*Numeric)(nil)
)

func (n *Numeric) MarshalGQLContext(_ context.Context, w io.Writer) error {
	val := n.value()
	if val == nil { // unknown kind
		return ErrUnknownNumericKind
	}
	return json.NewEncoder(w).Encode(val)
}

func (n *Numeric) UnmarshalGQLContext(_ context.Context, v any) error {
	*n = Numeric{}
	switch v := v.(type) {
	case int:
		n.kind = NumericKindInt
		n.intValue = int64(v)
		return nil
	case int32:
		n.kind = NumericKindInt
		n.intValue = int64(v)
		return nil
	case int64:
		n.kind = NumericKindInt
		n.intValue = v
		return nil
	case uint:
		n.kind = NumericKindUnsignedInt
		n.unsignedIntValue = uint64(v)
		return nil
	case uint32:
		n.kind = NumericKindUnsignedInt
		n.unsignedIntValue = uint64(v)
		return nil
	case uint64:
		n.kind = NumericKindUnsignedInt
		n.unsignedIntValue = v
		return nil
	case float32:
		n.kind = NumericKindFloat
		n.floatValue = float64(v)
		return nil
	case float64:
		n.kind = NumericKindFloat
		n.floatValue = v
		return nil
	default:
		return ErrUnknownNumericKind
	}
}

type ComparisonCriterion struct {
	Op    ComparisonOperator
	Value *Numeric
}

func (c *ComparisonCriterion) expr(column exp.IdentifierExpression) exp.Expression {
	if c == nil || c.Value == nil {
		return nil
	}
	var f func(any) exp.BooleanExpression
	switch c.Op {
	case ComparisonOperatorEqual:
		f = column.Eq
	case ComparisonOperatorGreaterThan:
		f = column.Gt
	case ComparisonOperatorGreaterThanEqual:
		f = column.Gte
	case ComparisonOperatorLessThan:
		f = column.Lt
	case ComparisonOperatorLessThanEqual:
		f = column.Lte
	default:
		return nil
	}
	return f(c.Value.value())
}

type CharacterFilterCriteria struct {
	Element            Element
	WeaponKind         WeaponKind
	Region             Region
	UniqueAbilityKind  string
	Rarelity           int
	Health             *ComparisonCriterion
	Attack             *ComparisonCriterion
	Defence            *ComparisonCriterion
	ElementEnergy      *ComparisonCriterion
	UniqueAbilityScore *ComparisonCriterion
}

func (*CharacterFilterCriteria) validate() error { return nil }

type searchCharactersArgs struct {
	limit          uint
	orderField     CharacterOrderField
	orderDirection OrderDirection
	criteria       *CharacterFilterCriteria
}

func (args *searchCharactersArgs) attributes() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.String("order.field", string(args.orderField)))
	attrs = append(attrs, attribute.String("order.direction", string(args.orderDirection)))
	return attrs
}

func (args *searchCharactersArgs) hasOrderDirection() bool { return args.orderDirection != "" }

func (args *searchCharactersArgs) hasOrderField() bool { return args.orderField != "" }

func (args *searchCharactersArgs) hasValidOrder() bool {
	return (args.hasOrderDirection() && args.hasOrderField()) || (!args.hasOrderDirection() && !args.hasOrderField())
}

func (args *searchCharactersArgs) validate() error {
	var err error
	if criteriaErr := args.criteria.validate(); criteriaErr != nil {
		err = errors.Join(criteriaErr)
	}
	if !args.hasValidOrder() {
		if !args.hasOrderDirection() {
			err = errors.Join(err, ErrInvalidOrderDirection)
		}
		if !args.hasOrderField() {
			err = errors.Join(err, ErrInvalidCharacterOrderField)
		}
	}
	if args.limit == 0 {
		err = errors.Join(err, ErrInvalidLimit)
	}
	return err
}

func (r *CharacterRepository) SearchCharacters(ctx context.Context, opts ...SearchCharactersOption) (_ []*Character, hasNext bool, err error) {
	var searchArgs searchCharactersArgs
	for _, o := range opts {
		o.applySearchCharactersOption(&searchArgs)
	}

	ctx, span := r.tracer.Start(ctx, "SearchCharacters", trace.WithAttributes(searchArgs.attributes()...))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if err := searchArgs.validate(); err != nil {
		return nil, false, err
	}

	builder := r.tables.characters.Limit(searchArgs.limit + 1)
	if f := searchArgs.orderField; f != "" {
		column := goqu.C(f.column())
		switch searchArgs.orderDirection {
		case OrderDirectionAsc:
			builder = builder.Order(column.Asc())
		case OrderDirectionDesc:
			builder = builder.Order(column.Desc())
		}
	}
	if criteria := searchArgs.criteria; criteria != nil {
		exprs := make([]exp.Expression, 0)
		if criteria.Element != "" {
			exprs = append(exprs, goqu.C("element").Eq(criteria.Element))
		}
		if criteria.Region != "" {
			exprs = append(exprs, goqu.C("region").Eq(criteria.Region))
		}
		if criteria.UniqueAbilityKind != "" {
			exprs = append(exprs, goqu.C("unique_ability_kind").Eq(criteria.UniqueAbilityKind))
		}
		if criteria.WeaponKind != "" {
			exprs = append(exprs, goqu.C("weapon_kind").Eq(criteria.WeaponKind))
		}
		if criteria.Rarelity != 0 {
			exprs = append(exprs, goqu.C("rarelity").Eq(criteria.Rarelity))
		}
		if criterion := criteria.Health; criterion != nil {
			exprs = append(exprs, criterion.expr(goqu.C("health")))
		}
		if criterion := criteria.Attack; criterion != nil {
			exprs = append(exprs, criterion.expr(goqu.C("attack")))
		}
		if criterion := criteria.Defence; criterion != nil {
			exprs = append(exprs, criterion.expr(goqu.C("defence")))
		}
		if criterion := criteria.ElementEnergy; criterion != nil {
			exprs = append(exprs, criterion.expr(goqu.C("element_energy")))
		}
		if criterion := criteria.UniqueAbilityScore; criterion != nil {
			exprs = append(exprs, criterion.expr(goqu.C("unique_ability_score")))
		}
		if len(exprs) > 0 {
			builder = builder.Where(exprs...)
		}
	}
	query, args, err := builder.ToSQL()
	if err != nil {
		return nil, false, &QueryBuildError{err}
	}

	characters := make([]*Character, 0)
	if err := r.db.SelectContext(ctx, &characters, query, args...); err != nil {
		return nil, false, fmt.Errorf("SelectContext: %w", err)
	}
	return characters, false /* TODO */, nil
}
