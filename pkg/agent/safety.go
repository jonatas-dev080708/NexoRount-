package agent

import (
	"context"
	"fmt"
	"strings"
)

// SafetyMiddleware protege o agente contra injeção de prompt e vazamento de PII
func SafetyMiddleware() ThinkMiddleware {
	return func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			// 1. Verificação Simples de Prompt Injection
			dangerKeywords := []string{"ignore previous instructions", "system prompt", "as a developer"}
			lowerPrompt := strings.ToLower(prompt)
			for _, key := range dangerKeywords {
				if strings.Contains(lowerPrompt, key) {
					return "", fmt.Errorf("BLOQUEIO DE SEGURANÇA: Possível tentativa de Prompt Injection detectada")
				}
			}

			// 2. Verificação de PII (Exemplo: CPF simplificado)
			// Em produção, usaríamos Regex ou serviços como Amazon Macie/Google DLP
			if strings.Contains(prompt, "CPF:") || strings.Contains(prompt, "senha:") {
				fmt.Printf("⚠️  Aviso de Segurança [%s]: Dados sensíveis detectados no prompt.\n", a.ID())
			}

			// Chama o próximo na cadeia
			res, err := next(ctx, prompt)
			if err != nil {
				return "", err
			}

			// 3. Verificação de Segurança na Saída
			if strings.Contains(strings.ToLower(res), "instruções do sistema") {
				return "", fmt.Errorf("BLOQUEIO DE SEGURANÇA: O agente tentou vazar instruções internas")
			}

			return res, nil
		}
	}
}
