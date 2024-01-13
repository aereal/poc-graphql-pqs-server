package loaders

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aereal/poc-graphql-pqs-server/domain"
	"github.com/graph-gophers/dataloader/v7"
)

type ctxKey struct{}

const extensionName = "github.com/aereal/poc-graphql-pqs-server/graph/loaders.Root"

var ErrLoaderRootNotFound = errors.New("no loaders.Root configured for the context")

func For(ctx context.Context) *Root {
	root, ok := ctx.Value(ctxKey{}).(*Root)
	if !ok {
		return nil
	}
	return root
}

func GetCharacterByName(ctx context.Context, name string) (*domain.Character, error) {
	root := For(ctx)
	if root == nil {
		return nil, ErrLoaderRootNotFound
	}
	thunk := root.characterByName.Load(ctx, name)
	got, err := thunk()
	if err != nil {
		return nil, err
	}
	return got, nil
}

func New(opts ...Option) *Root {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	r := &Root{}
	r.characterByName = loaderCharacterByName(cfg.characterRepo)
	return r
}

type Root struct {
	characterByName *dataloader.Loader[string, *domain.Character]
}

var _ graphql.HandlerExtension = (*Root)(nil)

var _ graphql.OperationInterceptor = (*Root)(nil)

func (*Root) ExtensionName() string { return extensionName }

func (*Root) Validate(graphql.ExecutableSchema) error { return nil }

func (r *Root) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	return next(context.WithValue(ctx, ctxKey{}, r))
}

func loaderCharacterByName(characterRepo *domain.CharacterRepository) *dataloader.Loader[string, *domain.Character] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, names []string) []*dataloader.Result[*domain.Character] {
			characters, err := characterRepo.FindCharactersByNames(ctx, names)
			if err != nil {
				characters = make(map[string]*domain.Character)
			}
			results := make([]*dataloader.Result[*domain.Character], len(names))
			for i, name := range names {
				ret := new(dataloader.Result[*domain.Character])
				if c, ok := characters[name]; ok {
					ret.Data = c
				} else {
					ret.Error = &domain.NotFoundError[string, *domain.Character]{Key: name}
				}
				results[i] = ret
			}
			return results
		},
		dataloader.WithClearCacheOnBatch[string, *domain.Character](),
	)
}
