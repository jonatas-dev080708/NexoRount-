package provider

import (
	"context"
	"fmt"
)

// Strategy define como o router deve escolher o provedor
type Strategy string

const (
	StrategyCheap Strategy = "cheap" // Prioriza modelos rápidos e baratos (ex: Gemini Flash, Haiku)
	StrategySmart Strategy = "smart" // Prioriza modelos inteligentes (ex: GPT-4, Claude Sonnet)
	StrategyFast  Strategy = "fast"  // Prioriza latência
)

// RouterProvider decide qual sub-provedor usar baseado na estratégia ou contexto
type RouterProvider struct {
	providers map[Strategy]LLMProvider
	current   Strategy
}

func NewRouterProvider(cheap, smart LLMProvider) *RouterProvider {
	return &RouterProvider{
		providers: map[Strategy]LLMProvider{
			StrategyCheap: cheap,
			StrategySmart: smart,
			StrategyFast:  cheap, // Geralmente o cheap é o mais fast
		},
		current: StrategySmart,
	}
}

func (r *RouterProvider) WithStrategy(s Strategy) *RouterProvider {
	r.current = s
	return r
}

func (r *RouterProvider) Name() string {
	p := r.providers[r.current]
	return fmt.Sprintf("Router(%s -> %s)", r.current, p.Name())
}

func (r *RouterProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	p, ok := r.providers[r.current]
	if !ok {
		return nil, fmt.Errorf("estratégia %s não configurada no router", r.current)
	}
	return p.Predict(ctx, messages, config)
}

func (r *RouterProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	p, ok := r.providers[r.current]
	if !ok {
		return nil, fmt.Errorf("estratégia %s não configurada no router", r.current)
	}
	return p.Stream(ctx, messages, config)
}

func (r *RouterProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Router geralmente usa o provedor cheap para embeddings por custo
	return r.providers[StrategyCheap].Embed(ctx, text)
}
