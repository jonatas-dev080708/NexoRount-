package provider

import (
	"errors"
	"os"
)

// NewMaritacaProvider cria um provedor para a API da Maritaca AI (focada em pt-BR)
func NewMaritacaProvider(model string) (*GenericOpenAIProvider, error) {
	apiKey := os.Getenv("MARITACA_API_KEY")
	if apiKey == "" {
		return nil, errors.New("MARITACA_API_KEY não encontrada no ambiente")
	}
	
	// A Maritaca AI fornece uma API compatível com OpenAI
	baseURL := "https://chat.maritaca.ai/api"
	return NewGenericProvider("Maritaca:"+model, baseURL, apiKey, model), nil
}

// Modelos comuns da Maritaca AI
const (
	MaritacaSabia3 = "sabia-3"
	MaritacaSabia2 = "sabia-2-small"
)
