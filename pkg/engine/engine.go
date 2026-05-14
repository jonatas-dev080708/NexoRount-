package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// Agent define o contrato para uma entidade autônoma no ecossistema
type Agent interface {
	ID() string
	Type() string
	// Run inicia a goroutine do agente. Ele deve observar o Context para shutdown.
	Run(ctx context.Context, nexus events.Bus) error
}

// Engine é o runtime operacional
type Engine struct {
	nexus  events.Bus
	agents map[string]Agent
	mu     sync.RWMutex
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		nexus:  events.NewLocalNexus(),
		agents: make(map[string]Agent),
		ctx:    ctx,
		cancel: cancel,
	}
}

// WithBus permite injetar um barramento diferente (ex: NATS)
func (e *Engine) WithBus(b events.Bus) *Engine {
	e.nexus = b
	return e
}

// Nexus expõe o barramento de eventos para assinaturas externas
func (e *Engine) Nexus() events.Bus {
	return e.nexus
}

// Register adiciona um agente ao ecossistema
func (e *Engine) Register(a Agent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.agents[a.ID()] = a
}

// Start liga o runtime e inicia todos os agentes em suas próprias goroutines
func (e *Engine) Start() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	fmt.Printf("🚀 Iniciando Runtime com %d agentes...\n", len(e.agents))

	for _, agent := range e.agents {
		e.wg.Add(1)
		go func(a Agent) {
			defer e.wg.Done()
			if err := a.Run(e.ctx, e.nexus); err != nil {
				fmt.Printf("⚠️ Erro no Agente [%s]: %v\n", a.ID(), err)
			}
		}(agent)
	}

	// Sincronização de inicialização: dá um tempo para as goroutines registrarem seus handlers no Nexus
	time.Sleep(50 * time.Millisecond)
	fmt.Println("✅ Todos os agentes estão em posição.")
}

// Publish expõe o Nexus para injeção de eventos externos
func (e *Engine) Publish(event events.Event) {
	e.nexus.Publish(event)
}

// Stop desliga todos os agentes graciosamente
func (e *Engine) Stop() {
	fmt.Println("🛑 Desligando ecossistema...")
	e.cancel()
	e.wg.Wait()
	e.nexus.Close()
	fmt.Println("✨ Sistema encerrado com sucesso.")
}
