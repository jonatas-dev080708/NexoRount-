package provider

import (
	"context"
	"fmt"
)

// MultiProvider implementa o padrão Fallback
type MultiProvider struct {
	providers []LLMProvider
}

func NewMultiProvider(providers ...LLMProvider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

func (p *MultiProvider) Name() string {
	return fmt.Sprintf("MultiProvider(%d)", len(p.providers))
}

func (p *MultiProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	var lastErr error
	for _, prov := range p.providers {
		res, err := prov.Predict(ctx, messages, config)
		if err == nil {
			return res, nil
		}
		lastErr = err
		fmt.Printf("⚠️  Provider [%s] falhou: %v. Tentando próximo...\n", prov.Name(), err)
	}
	return nil, fmt.Errorf("todos os providers falharam. Último erro: %v", lastErr)
}

func (p *MultiProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	// Fallback no stream é mais complexo, aqui tentamos o primeiro que não der erro ao abrir
	for _, prov := range p.providers {
		ch, err := prov.Stream(ctx, messages, config)
		if err == nil {
			return ch, nil
		}
		fmt.Printf("⚠️  Provider Stream [%s] falhou: %v. Tentando próximo...\n", prov.Name(), err)
	}
	return nil, fmt.Errorf("todos os providers de stream falharam")
}

func (p *MultiProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	for _, prov := range p.providers {
		res, err := prov.Embed(ctx, text)
		if err == nil {
			return res, nil
		}
	}
	return nil, fmt.Errorf("todos os providers de embedding falharam")
}
