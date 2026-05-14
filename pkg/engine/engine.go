package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/observability"
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
	ctx     context.Context
	cancel  context.CancelFunc
	journal *observability.Journal
	kernel  *CognitiveKernel
}

func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		nexus:   events.NewLocalNexus(),
		agents:  make(map[string]Agent),
		ctx:     ctx,
		cancel:  cancel,
		journal: observability.NewJournal(),
		kernel:  NewCognitiveKernel(100), // Default 100 global parallel thoughts
	}
}

// WithBus permite injetar um barramento diferente (ex: NATS)
func (e *Engine) WithBus(b events.Bus) *Engine {
	e.nexus = b
	return e
}

// WithRemoteBus ativa o modo híbrido, conectando o nexus local a um barramento remoto
func (e *Engine) WithRemoteBus(remote events.Bus, types ...events.Type) *Engine {
	bridge := events.NewBridge(e.nexus, remote)
	bridge.Forward(types...)
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
	
	// Injeta o Kernel e o Journal automaticamente se o agente for do tipo BaseAgent
	if ba, ok := a.(*agent.BaseAgent); ok {
		ba.WithJournal(e.journal)
		ba.WithMiddleware(e.kernel.KernelMiddleware())
	}

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

// Kernel retorna o acesso ao escalonador cognitivo
func (e *Engine) Kernel() *CognitiveKernel {
	return e.kernel
}
	// Se o evento não tiver TraceID, criamos um novo para iniciar a jornada
	if event.TraceID == "" {
		event.TraceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixNano()
	}

	// Registra no diário de bordo antes de publicar
	e.journal.RecordEvent(event)
	
	e.nexus.Publish(event)
}

// Journal retorna o acesso ao histórico para inspeção e debugging
func (e *Engine) Journal() *observability.Journal {
	return e.journal
}

// Stop desliga todos os agentes graciosamente
func (e *Engine) Stop() {
	fmt.Println("🛑 Desligando ecossistema...")
	e.cancel()
	e.wg.Wait()
	e.nexus.Close()
	fmt.Println("✨ Sistema encerrado com sucesso.")
}
