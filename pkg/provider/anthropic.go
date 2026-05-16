package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(model string) (*AnthropicProvider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY não encontrada")
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

func (p *AnthropicProvider) Name() string {
	return "Anthropic:" + p.model
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		TotalTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *AnthropicProvider) Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error) {
	var system string
	var apiMessages []anthropicMessage

	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
		} else {
			apiMessages = append(apiMessages, anthropicMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	maxTokens := 1024
	if config != nil && config.MaxTokens > 0 {
		maxTokens = config.MaxTokens
	}

	reqBody := anthropicRequest{
		Model:     p.model,
		Messages:  apiMessages,
		MaxTokens: maxTokens,
		System:    system,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var anthroResp anthropicResponse
	if err := json.Unmarshal(body, &anthroResp); err != nil {
		return nil, err
	}

	if anthroResp.Error != nil {
		return nil, fmt.Errorf("anthropic api error: %s", anthroResp.Error.Message)
	}

	if len(anthroResp.Content) == 0 {
		return nil, fmt.Errorf("resposta vazia da anthropic")
	}

	return &Result{
		Content: anthroResp.Content[0].Text,
		Tokens:  anthroResp.Usage.TotalTokens,
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error) {
	return nil, fmt.Errorf("streaming não implementado para Anthropic ainda")
}

func (p *AnthropicProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("anthropic não suporta embeddings nativos via mensagens")
}
