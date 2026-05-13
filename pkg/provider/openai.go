package provider

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(model string) (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY não encontrada no ambiente")
	}

	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model,
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI:" + p.model
}

func (p *OpenAIProvider) Predict(ctx context.Context, messages []Message) (*Result, error) {
	apiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		apiMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: apiMessages,
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		Content: resp.Choices[0].Message.Content,
		Tokens:  resp.Usage.TotalTokens,
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, messages []Message) (chan string, error) {
	apiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		apiMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: apiMessages,
	})

	if err != nil {
		return nil, err
	}

	ch := make(chan string)

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
			ch <- response.Choices[0].Delta.Content
		}
	}()

	return ch, nil
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
