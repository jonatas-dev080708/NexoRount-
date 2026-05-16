package tools

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
)

// NewHTTPGetTool cria uma ferramenta para fazer requisições GET simples
func NewHTTPGetTool() *agent.Tool {
	return &agent.Tool{
		Name:        "http_get",
		Description: "Faz uma requisição HTTP GET para uma URL e retorna o corpo da resposta em texto. Use para buscar informações em APIs ou sites.",
		Execute: func(args string) (string, error) {
			client := &http.Client{
				Timeout: 15 * time.Second,
			}

			resp, err := client.Get(args)
			if err != nil {
				return "", fmt.Errorf("falha na requisição HTTP: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("servidor retornou erro: %d %s", resp.StatusCode, resp.Status)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("falha ao ler corpo da resposta: %v", err)
			}

			// Limita o retorno para não estourar o contexto do agente
			content := string(body)
			if len(content) > 10000 {
				content = content[:10000] + "... [truncado]"
			}

			return content, nil
		},
	}
}
