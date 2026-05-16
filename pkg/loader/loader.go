package loader

import (
	"context"
)

// Document representa um pedaço de informação carregada
type Document struct {
	Content  string
	Metadata map[string]interface{}
}

// Loader define a interface para carregar dados de fontes externas
type Loader interface {
	Load(ctx context.Context) ([]Document, error)
}
