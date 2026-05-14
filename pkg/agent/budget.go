package agent

import (
	"context"
	"fmt"
	"sync"
)

// BudgetManager controla o consumo de recursos (tokens/custo) por sessão ou agente
type BudgetManager struct {
	mu            sync.RWMutex
	maxTokens     int
	usedTokens    int
	maxCost       float64
	usedCost      float64
	stopOnExceed  bool
}

func NewBudgetManager(maxTokens int, maxCost float64) *BudgetManager {
	return &BudgetManager{
		maxTokens:    maxTokens,
		maxCost:      maxCost,
		stopOnExceed: true,
	}
}

// AddConsumption registra o uso de tokens e custo
func (b *BudgetManager) AddConsumption(tokens int, cost float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.usedTokens += tokens
	b.usedCost += cost

	if b.stopOnExceed {
		if b.maxTokens > 0 && b.usedTokens > b.maxTokens {
			return fmt.Errorf("limite de tokens excedido: %d/%d", b.usedTokens, b.maxTokens)
		}
		if b.maxCost > 0 && b.usedCost > b.maxCost {
			return fmt.Errorf("limite de custo excedido: %.4f/%.4f", b.usedCost, b.maxCost)
		}
	}
	return nil
}

// BudgetMiddleware retorna um middleware que monitora e limita o uso de tokens
func BudgetMiddleware(manager *BudgetManager) ThinkMiddleware {
	return func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			// Antes de pensar, verificamos se ainda há saldo (estimativa simples ou check prévio)
			// Aqui apenas executamos e validamos o retorno
			
			res, err := next(ctx, prompt)
			if err != nil {
				return "", err
			}

			// Simulação de cálculo de tokens (no futuro pegaremos do provider.Result)
			// Por enquanto, vamos estimar 1 token por 4 caracteres
			estimatedTokens := len(prompt)/4 + len(res)/4
			
			// Atualiza o manager
			if err := manager.AddConsumption(estimatedTokens, 0); err != nil {
				return res, fmt.Errorf("AVISO DE BUDGET: %v", err)
			}

			return res, nil
		}
	}
}
