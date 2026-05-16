package agent

import (
	"context"
	"fmt"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// EnableSelfHealing ativa a capacidade de auto-correção reativa
func EnableSelfHealing(a *BaseAgent) {
	
	// Ouve falhas em ferramentas
	a.On("TOOL_ERROR", func(agent *BaseAgent, ctx context.Context, e events.Event) {
		data := e.Payload.(map[string]interface{})
		errStr := data["error"].(string)
		toolName := data["tool"].(string)
		prevState := data["context"].(map[string]interface{})

		fmt.Printf("🩹 [Self-Healing] Detectada falha na ferramenta %s: %s. Tentando correção...\n", toolName, errStr)

		// O agente analisa o erro e sugere uma correção para si mesmo
		prompt := fmt.Sprintf(`
			A ferramenta %s falhou com o erro: %s
			Analise o motivo e sugira como corrigir o comando ou uma abordagem alternativa.
		`, toolName, errStr)

		correction, _ := agent.Think(ctx, prompt)

		// Evolui o histórico com a correção e re-emite o passo de pensamento
		prevState["history"] = prevState["history"].(string) + fmt.Sprintf("\nErro na %s: %s. Sugestão: %s", toolName, errStr, correction)
		
		// Tenta novamente emitindo o evento de pensamento
		agent.Emit("AGENT_THINK_STEP", prevState)
	})
}
