package provider

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(ctx context.Context, model string) (*GeminiProvider, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY não encontrada no ambiente")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente Gemini: %v", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "Gemini:" + p.model
}

func (p *GeminiProvider) Predict(ctx context.Context, messages []Message) (*Result, error) {
	model := p.client.GenerativeModel(p.model)
	
	// Convertemos nosso formato universal para o formato do Gemini
	// Para simplicidade inicial, enviamos a última mensagem como prompt
	// e as anteriores como contexto se necessário (evoluiremos isso)
	prompt := ""
	if len(messages) > 0 {
		prompt = messages[len(messages)-1].Content
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, errors.New("nenhuma resposta gerada pelo Gemini")
	}

	content := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		content += fmt.Sprintf("%v", part)
	}

	return &Result{
		Content: content,
		Tokens:  0, // Gemini SDK Go não expõe tokens de forma trivial no GenerateContent simples
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, messages []Message) (chan string, error) {
	model := p.client.GenerativeModel(p.model)
	
	prompt := ""
	if len(messages) > 0 {
		prompt = messages[len(messages)-1].Content
	}

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	ch := make(chan string)

	go func() {
		defer close(ch)
		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				return
			}
			if err != nil {
				return
			}
			
			for _, part := range resp.Candidates[0].Content.Parts {
				ch <- fmt.Sprintf("%v", part)
			}
		}
	}()

	return ch, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// O Gemini exige um modelo específico para embeddings
	model := p.client.EmbeddingModel("text-embedding-004")
	res, err := model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}
	return res.Embedding.Values, nil
}

func (p *GeminiProvider) Close() error {
	return p.client.Close()
}
