package domain

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewCharacterRepository(opts ...CharacterRepositoryOption) *CharacterRepository {
	r := &CharacterRepository{
		tracer: otel.GetTracerProvider().Tracer(pkgName + ".CharacterRepository"),
	}
	for _, o := range opts {
		o.applyCharacterRepositoryOption(r)
	}
	r.tables.characters = goqu.Dialect("postgres").From("characters").Prepared(true)
	return r
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
