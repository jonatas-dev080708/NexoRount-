package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockProvider struct {
	client *bedrockruntime.Client
	model  string
}

func NewBedrockProvider(ctx context.Context, region, model string) (*BedrockProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar config AWS: %v", err)
	}

	return &BedrockProvider{
		client: bedrockruntime.NewFromConfig(cfg),
		model:  model,
	}, nil
}

func (p *BedrockProvider) Name() string {
	return "Bedrock:" + p.model
}

func (p *BedrockProvider) Predict(ctx context.Context, messages []Message, cfg *LLMConfig) (*Result, error) {
	// Bedrock usa formatos diferentes por modelo. Vamos focar no formato de mensagens (InvokeModel)
	// Este exemplo é simplificado para modelos Claude/Llama no Bedrock
	
	payload := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        1024,
		"messages":          messages,
	}

	if cfg != nil {
		if cfg.MaxTokens > 0 {
			payload["max_tokens"] = cfg.MaxTokens
		}
		payload["temperature"] = cfg.Temperature
	}

	payloadBytes, _ := json.Marshal(payload)

	output, err := p.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(p.model),
		ContentType: aws.String("application/json"),
		Body:        payloadBytes,
	})

	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	json.Unmarshal(output.Body, &resp)

	// O formato de resposta varia, aqui assumimos o padrão Anthropic no Bedrock
	content := ""
	if c, ok := resp["content"].([]interface{}); ok && len(c) > 0 {
		if m, ok := c[0].(map[string]interface{}); ok {
			content = m["text"].(string)
		}
	}

	return &Result{
		Content: content,
	}, nil
}

func (p *BedrockProvider) Stream(ctx context.Context, messages []Message, cfg *LLMConfig) (chan string, error) {
	return nil, fmt.Errorf("streaming não implementado para Bedrock ainda")
}

func (p *BedrockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("embeddings não implementados para Bedrock ainda")
}
