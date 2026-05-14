package provider

import (
	"context"
)

// Message representa uma entrada/saída de chat para as LLMs
type Message struct {
	Role    string
	Content string
}

// LLMConfig define parâmetros de geração
type LLMConfig struct {
	Temperature    float32
	MaxTokens      int
	TopP           float32
	StopSequences  []string
	ResponseFormat string // "text" ou "json_object"
}

// Result contém a resposta final e metadados de uso
type Result struct {
	Content string
	Tokens  int
}

// LLMProvider é a interface que qualquer modelo (OpenAI, Gemini, Ollama) deve implementar
type LLMProvider interface {
	Name() string
	Predict(ctx context.Context, messages []Message, config *LLMConfig) (*Result, error)
	// Stream permite receber a resposta token a token
	Stream(ctx context.Context, messages []Message, config *LLMConfig) (chan string, error)
	// Embed converte texto em um vetor numérico (embedding)
	Embed(ctx context.Context, text string) ([]float32, error)
}
