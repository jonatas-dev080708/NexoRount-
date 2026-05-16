package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
)

// NewBrowserTool cria uma ferramenta de navegação web real usando Chrome
func NewBrowserTool() *agent.Tool {
	return &agent.Tool{
		Name:        "browser_navigate",
		Description: "Navega para uma URL usando um navegador real, aguarda o carregamento e extrai todo o texto visível da página. Use para sites complexos (SPA) que exigem JavaScript.",
		Execute: func(url string) (string, error) {
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			// Timeout de segurança
			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			var body string
			err := chromedp.Run(ctx,
				chromedp.Navigate(url),
				chromedp.Sleep(2*time.Second), // Aguarda renderização JS
				chromedp.Text("body", &body, chromedp.ByQuery),
			)

			if err != nil {
				return "", fmt.Errorf("erro na navegação browser: %v", err)
			}

			if len(body) > 15000 {
				body = body[:15000] + "... [conteúdo muito longo]"
			}

			return body, nil
		},
	}
}

// NewBrowserScreenshotTool (Opcional para o futuro)
func NewBrowserScreenshotTool() *agent.Tool {
	return &agent.Tool{
		Name:        "browser_screenshot",
		Description: "Tira um print da página. Útil para agentes visuais.",
		Execute: func(url string) (string, error) {
			return "Funcionalidade de screenshot agendada para v2.0", nil
		},
	}
}
