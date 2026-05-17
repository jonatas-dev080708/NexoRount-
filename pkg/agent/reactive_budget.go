package agent

import (
	"context"
	"fmt"
	"sync"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// ReactiveBudgetManager controla o gasto de tokens via eventos
type ReactiveBudgetManager struct {
	mu      sync.Mutex
	budgets map[string]int // agentID -> remainingTokens
}

func NewReactiveBudgetManager() *ReactiveBudgetManager {
	return &ReactiveBudgetManager{budgets: make(map[string]int)}
}

func (bm *ReactiveBudgetManager) SetBudget(agentID string, tokens int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.budgets[agentID] = tokens
}

// EnableReactiveBudgeting ativa o controle de custos no Nexus
func EnableReactiveBudgeting(a *BaseAgent, manager *ReactiveBudgetManager) {
	
	// Ouve o início de qualquer predição
	a.On("LLM_PREDICT_START", func(agent *BaseAgent, ctx context.Context, e events.Event) {
		manager.mu.Lock()
		remaining, ok := manager.budgets[e.Source]
		manager.mu.Unlock()

		if ok && remaining <= 0 {
			fmt.Printf("🛑 [Budget] Agente %s esgotou o orçamento! Bloqueando...\n", e.Source)
			// Emite um evento de veto
			agent.Emit("LLM_PREDICT_VETO", map[string]interface{}{
				"agent_id": e.Source,
				"reason":   "Budget exceeded",
			})
		}
	})

	// Ouve o resultado para atualizar o saldo
	a.On("LLM_PREDICT_END", func(agent *BaseAgent, ctx context.Context, e events.Event) {
		data := e.Payload.(map[string]interface{})
		tokens := data["tokens"].(int)

		manager.mu.Lock()
		if _, ok := manager.budgets[e.Source]; ok {
			manager.budgets[e.Source] -= tokens
			fmt.Printf("📉 [Budget] Agente %s gastou %d tokens. Restante: %d\n", e.Source, tokens, manager.budgets[e.Source])
		}
		manager.mu.Unlock()
	})
}
