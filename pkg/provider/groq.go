package provider

import (
	"os"
)

// NewGroqProvider cria um provedor para a API do Groq (ultra-rápida)
func NewGroqProvider(model string) (*GenericOpenAIProvider, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		apiKey = "" // Permite passar depois se necessário, mas idealmente vem do env
	}
	
	baseURL := "https://api.groq.com/openai/v1"
	return NewGenericProvider("Groq:"+model, baseURL, apiKey, model), nil
}

// Modelos comuns do Groq
const (
	GroqLlama3_8b  = "llama3-8b-8192"
	GroqLlama3_70b = "llama3-70b-8192"
	GroqMixtral8x7b = "mixtral-8x7b-32768"
)
