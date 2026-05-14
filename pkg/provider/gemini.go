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

func (p *GeminiProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	model := p.client.GenerativeModel(p.model)

	if config != nil {
		model.SetTemperature(config.Temperature)
		model.SetMaxOutputTokens(int32(config.MaxTokens))
		model.SetTopP(config.TopP)
		model.StopSequences = config.StopSequences
		if config.ResponseFormat == "json_object" {
			model.ResponseMIMEType = "application/json"
		}
	}
	
	// Prepara o chat/contexto
	var systemInstruction *genai.Content
	var history []*genai.Content

	for _, m := range messages {
		if m.Role == "system" {
			systemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(m.Content)},
			}
		} else {
			role := "user"
			if m.Role == "assistant" || m.Role == "model" {
				role = "model"
			}
			history = append(history, &genai.Content{
				Role:  role,
				Parts: []genai.Part{genai.Text(m.Content)},
			})
		}
	}

	if systemInstruction != nil {
		model.SystemInstruction = systemInstruction
	}

	// Pega a última mensagem como o prompt atual e remove do histórico
	if len(history) == 0 {
		return nil, errors.New("nenhuma mensagem enviada")
	}
	
	lastMsg := history[len(history)-1]
	history = history[:len(history)-1]

	chat := model.StartChat()
	chat.History = history

	resp, err := chat.SendMessage(ctx, lastMsg.Parts...)
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
		Tokens:  0,
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	model := p.client.GenerativeModel(p.model)

	if config != nil {
		model.SetTemperature(config.Temperature)
		model.SetMaxOutputTokens(int32(config.MaxTokens))
		model.SetTopP(config.TopP)
		model.StopSequences = config.StopSequences
	}
	
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
