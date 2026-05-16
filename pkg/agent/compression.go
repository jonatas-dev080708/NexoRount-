package agent

import (
	"context"
	"fmt"

	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

// PromptCompressionMiddleware reduz o tamanho do contexto se ele exceder um limite
func PromptCompressionMiddleware(maxChars int) ThinkMiddleware {
	return func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			if len(prompt) <= maxChars {
				return next(ctx, prompt)
			}

			fmt.Printf("📉 [Compression] Prompt muito longo (%d chars). Comprimindo...\n", len(prompt))

			// Estratégia de compressão: Pede para a própria LLM sumarizar o prompt mantendo os pontos chave
			// Usamos uma instrução interna para isso
			compressionPrompt := fmt.Sprintf("Sumarize o seguinte contexto de forma ultra-concisa, mantendo todos os fatos, nomes e requisitos cruciais para uma tarefa de IA:\n\n%s", prompt)
			
			// Chamada direta para não entrar em loop de middleware
			summary, err := a.llm.Predict(ctx, []provider.Message{
				{Role: "user", Content: compressionPrompt},
			}, nil)

			if err != nil {
				// Se falhar a compressão, tenta seguir com o original (pode dar erro de context window)
				return next(ctx, prompt)
			}

			fmt.Printf("✅ [Compression] Reduzido para %d chars\n", len(summary.Content))
			return next(ctx, summary.Content)
		}
	}
}

// TokenCounterMiddleware apenas loga o uso estimado (pode ser expandido)
func TokenCounterMiddleware() ThinkMiddleware {
	return func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			res, err := next(ctx, prompt)
			// No futuro, aqui poderíamos atualizar o Budget do agente
			return res, err
		}
	}
}
