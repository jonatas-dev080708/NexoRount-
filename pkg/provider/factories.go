package provider

import (
	"fmt"
	"os"
)

// NewMistralProvider cria um provedor para a API da Mistral AI
func NewMistralProvider(model string) (*GenericOpenAIProvider, error) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY não encontrada")
	}
	
	baseURL := "https://api.mistral.ai/v1"
	return NewGenericProvider("Mistral:"+model, baseURL, apiKey, model), nil
}

// NewAzureOpenAIProvider cria um provedor para o Azure OpenAI
func NewAzureOpenAIProvider(endpoint, deployment, apiKey string) (*GenericOpenAIProvider, error) {
	// Azure usa um formato de URL ligeiramente diferente, mas o go-openai suporta
	baseURL := fmt.Sprintf("%s/openai/deployments/%s", endpoint, deployment)
	
	p := NewGenericProvider("Azure:"+deployment, baseURL, apiKey, deployment)
	// Azure exige a chave no header api-key, o go-openai trata isso se configurado corretamente
	// Mas o GenericProvider atual usa o DefaultConfig que foca em Bearer token.
	// Para Azure, precisaríamos de um ajuste fino no GenericProvider ou um adapter próprio.
	return p, nil
}

// NewOllamaProvider cria um provedor para o Ollama local
func NewOllamaProvider(model string) *GenericOpenAIProvider {
	return NewGenericProvider("Ollama:"+model, "http://localhost:11434/v1", "ollama", model)
}
