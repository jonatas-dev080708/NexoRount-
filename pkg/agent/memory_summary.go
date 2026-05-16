package agent

import (
	"context"
	"fmt"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

// MemorySummaryMiddleware monitora o histórico e cria resumos automáticos
// Isso evita que o contexto cresça infinitamente, mantendo apenas a essência.
func MemorySummaryMiddleware(threshold int) ThinkMiddleware {
	return func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			// Se o agente tem memória de sessão e ela está ficando grande
			// Aqui simulamos a lógica de checagem de tamanho de tokens
			if len(prompt) > threshold {
				fmt.Printf("🧠 [Memory] Histórico atingiu o limite (%d). Sumarizando...\n", len(prompt))
				
				summaryPrompt := fmt.Sprintf("Sumarize os pontos principais desta conversa de forma que um agente de IA possa continuar o trabalho sem perder o contexto essencial:\n\n%s", prompt)
				
				res, err := a.llm.Predict(ctx, []provider.Message{
					{Role: "user", Content: summaryPrompt},
				}, nil)

				if err == nil {
					fmt.Printf("✅ [Memory] Contexto compactado com sucesso.\n")
					prompt = "Resumo da conversa anterior: " + res.Content
				}
			}

			return next(ctx, prompt)
		}
	}
}
