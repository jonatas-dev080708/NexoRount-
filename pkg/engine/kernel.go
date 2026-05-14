package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// CognitiveKernel é o escalonador avançado de recursos mentais
type CognitiveKernel struct {
	mu                sync.RWMutex
	maxGlobalThoughts int
	activeThoughts    int
	agentCapacities   map[string]int // Limite de concorrência por agente (para evitar 'fadiga')
	agentActive       map[string]int
	
	// Fila de prioridade simplificada
	priorityQueue chan events.Event
}

func NewCognitiveKernel(maxGlobal int) *CognitiveKernel {
	return &CognitiveKernel{
		maxGlobalThoughts: maxGlobal,
		agentCapacities:   make(map[string]int),
		agentActive:       make(map[string]int),
		priorityQueue:     make(chan events.Event, 1000),
	}
}

// SetAgentCapacity define quantos pensamentos simultâneos um agente pode ter
func (k *CognitiveKernel) SetAgentCapacity(agentID string, capacity int) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.agentCapacities[agentID] = capacity
}

// Acquire slot para processamento (respeita prioridade e fadiga)
func (k *CognitiveKernel) Acquire(ctx context.Context, agentID string, priority int) error {
	for {
		k.mu.Lock()
		
		// 1. Verifica Global
		if k.activeThoughts >= k.maxGlobalThoughts && priority < 100 { // 100 é emergência, ignora limite
			k.mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Backoff simples
			continue
		}

		// 2. Verifica Fadiga do Agente
		cap, ok := k.agentCapacities[agentID]
		if !ok { cap = 5 } // Default cap
		
		active := k.agentActive[agentID]
		if active >= cap && priority < 100 {
			k.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Sucesso: Aloca recursos
		k.activeThoughts++
		k.agentActive[agentID]++
		k.mu.Unlock()
		return nil
	}
}

// Release devolve os recursos mentais ao kernel
func (k *CognitiveKernel) Release(agentID string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.activeThoughts--
	k.agentActive[agentID]--
	
	if k.activeThoughts < 0 { k.activeThoughts = 0 }
	if k.agentActive[agentID] < 0 { k.agentActive[agentID] = 0 }
}

// KernelMiddleware integra o escalonador no ciclo de vida do agente
func (k *CognitiveKernel) KernelMiddleware() agent.ThinkMiddleware {
	return func(a *agent.BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
		return func(ctx context.Context, prompt string) (string, error) {
			// Antes de pensar, o Kernel precisa autorizar
			// Aqui usamos uma prioridade padrão ou pegamos do contexto se disponível
			priority := 0 
			
			if err := k.Acquire(ctx, a.ID(), priority); err != nil {
				return "", fmt.Errorf("kernel recusou processamento: %v", err)
			}
			defer k.Release(a.ID())

			return next(ctx, prompt)
		}
	}
}
