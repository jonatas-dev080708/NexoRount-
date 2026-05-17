package provider

import (
	"context"
	"time"
)

type MockProvider struct{}

func (p *MockProvider) Name() string { return "MockProvider" }

func (p *MockProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	return &Result{
		Content: "SIM", // Resposta padrão para passar nos testes de lógica
		Tokens:  0,
	}, nil
}

func (p *MockProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	ch := make(chan string)
	words := []string{"Esta ", "é ", "uma ", "resposta ", "viva ", "em ", "streaming ", "do ", "seu ", "framework!"}
	
	go func() {
		defer close(ch)
		for _, w := range words {
			ch <- w
			time.Sleep(150 * time.Millisecond) // Simula o tempo de geração
		}
	}()
	return ch, nil
}

func (p *MockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return make([]float32, 1536), nil
}
