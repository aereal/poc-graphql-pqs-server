package apollo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/99designs/gqlgen/graphql"
)

type queryList map[string]string

type ManifestFilePath string

func ProvideManifestFromPath(file ManifestFilePath) (*Manifest, error) {
	f, err := os.Open(string(file))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	manifest := new(Manifest)
	if err := json.NewDecoder(f).Decode(manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}
	return manifest, nil
}

func ProvideQueryCacheFromManifest(manifest *Manifest) (graphql.Cache, error) {
	list := make(queryList)
	for _, op := range manifest.Operations {
		list[op.ID] = op.Body
	}
	return list, nil
}

var _ graphql.Cache = (queryList)(nil)

func (l queryList) Get(_ context.Context, hash string) (any, bool) {
	query, ok := l[hash]
	return query, ok
}

func (queryList) Add(context.Context, string, any) {}

type Manifest struct {
	Format     string      `json:"format"`
	Version    int         `json:"version"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Body string `json:"body"`
}
