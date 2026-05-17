package agent

import (
	"context"
	"fmt"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// EnableMemoryConsolidation cria um processo que observa o Nexus e salva aprendizados
func EnableMemoryConsolidation(a *BaseAgent) {
	// Ouve quando qualquer agente (ou ele mesmo) termina uma tarefa
	a.On("AGENT_FINAL_ANSWER", func(agent *BaseAgent, ctx context.Context, e events.Event) {
		result := e.Payload.(string)
		
		fmt.Printf("🧠 [MemoryConsolidation] Analisando resultado de %s para indexação...\n", e.Source)

		// O consolidificador "pensa" se isso deve ser lembrado
		prompt := fmt.Sprintf(`
			Extraia o conhecimento crucial, fatos ou decisões desta resposta: %s
			Resuma em uma frase curta para ser salva na memória de longo prazo.
			Se não houver nada útil para o futuro, responda "IGNORAR".
		`, result)

		summary, err := agent.Think(ctx, prompt)
		if err != nil || summary == "IGNORAR" {
			return
		}

		fmt.Printf("✅ [MemoryConsolidation] Novo aprendizado salvo: %s\n", summary)
		
		// Aqui salvaríamos no Vector Store (via evento ou chamada direta)
		agent.Emit("MEMORY_STORE_REQUEST", map[string]interface{}{
			"content":  summary,
			"agent_id": e.Source,
			"metadata": map[string]interface{}{"source": "consolidation", "original_trace": e.TraceID},
		})
	})
}
