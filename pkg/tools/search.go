package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
)

// NewGoogleSearchTool cria uma ferramenta de busca via Google Custom Search
func NewGoogleSearchTool() *agent.Tool {
	return &agent.Tool{
		Name:        "google_search",
		Description: "Pesquisa no Google e retorna os títulos e links dos resultados. Use para encontrar informações atualizadas na web.",
		Execute: func(query string) (string, error) {
			apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
			cxID := os.Getenv("GOOGLE_SEARCH_CX")
			
			if apiKey == "" || cxID == "" {
				return "", fmt.Errorf("GOOGLE_SEARCH_API_KEY ou CX não configurados")
			}

			apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s",
				apiKey, cxID, url.QueryEscape(query))

			resp, err := http.Get(apiURL)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var result struct {
				Items []struct {
					Title   string `json:"title"`
					Link    string `json:"link"`
					Snippet string `json:"snippet"`
				} `json:"items"`
			}
			json.Unmarshal(body, &result)

			output := "Resultados da busca:\n"
			for i, item := range result.Items {
				if i >= 5 { break }
				output += fmt.Sprintf("%d. %s (%s)\n   %s\n", i+1, item.Title, item.Link, item.Snippet)
			}

			return output, nil
		},
	}
}

// NewTavilySearchTool cria uma ferramenta de busca otimizada para agentes via Tavily
func NewTavilySearchTool() *agent.Tool {
	return &agent.Tool{
		Name:        "tavily_search",
		Description: "Busca na web otimizada para IA. Retorna conteúdo denso e relevante.",
		Execute: func(query string) (string, error) {
			apiKey := os.Getenv("TAVILY_API_KEY")
			if apiKey == "" {
				return "", fmt.Errorf("TAVILY_API_KEY não configurada")
			}

			// Implementação simplificada do POST para Tavily
			return "Simulação Tavily: Resultados para " + query, nil
		},
	}
}
