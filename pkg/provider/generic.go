package provider

import (
	"context"
	"io"
	"errors"

	"github.com/sashabaranov/go-openai"
)

type GenericOpenAIProvider struct {
	client *openai.Client
	model  string
	name   string
}

// NewGenericProvider cria um provedor para qualquer API compatível com OpenAI (DeepSeek, Groq, Ollama, Maritaca)
func NewGenericProvider(name, baseURL, apiKey, model string) *GenericOpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	
	return &GenericOpenAIProvider{
		client: openai.NewClientWithConfig(config),
		model:  model,
		name:   name,
	}
}

func (p *GenericOpenAIProvider) Name() string {
	return p.name
}

func (p *GenericOpenAIProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	oaMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		oaMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: oaMessages,
	}

	if config != nil {
		req.Temperature = config.Temperature
		req.MaxTokens = config.MaxTokens
		req.TopP = config.TopP
		req.Stop = config.StopSequences
		if config.ResponseFormat == "json_object" {
			req.ResponseFormat = &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			}
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return nil, err
	}

	return &Result{
		Content: resp.Choices[0].Message.Content,
		Tokens:  resp.Usage.TotalTokens,
	}, nil
}

func (p *GenericOpenAIProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	oaMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		oaMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: oaMessages,
		Stream:   true,
	}

	if config != nil {
		req.Temperature = config.Temperature
		req.MaxTokens = config.MaxTokens
		req.TopP = config.TopP
		req.Stop = config.StopSequences
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return nil, err
	}

	ch := make(chan string, 100)
	go func() {
		defer close(ch)
		defer stream.Close()
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				return
			}
			if len(response.Choices) > 0 {
				ch <- response.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

func (p *GenericOpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: "text-embedding-3-small", // Ou o modelo que o provedor suportar
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
