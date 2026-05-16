package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/observability"
	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
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
	persistentNexus *events.PersistentNexus
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

// WithPersistence ativa a camada de durabilidade dos eventos usando SQLite.
// Isso garante que eventos não processados sejam restaurados em caso de queda do sistema.
func (e *Engine) WithPersistence(dbPath string) *Engine {
	p, err := events.NewPersistentNexus(dbPath)
	if err != nil {
		fmt.Printf("❌ Erro ao ativar persistência: %v\n", err)
		return e
	}
	e.persistentNexus = p
	e.nexus = p // Substitui o nexus pelo persistente
	return e
}

// Nexus retorna o barramento de eventos central do Engine.
func (e *Engine) Nexus() events.Bus {
	return e.nexus
}

// Register adiciona um agente ao ecossistema (antes do Start)
func (e *Engine) Register(a Agent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.internalRegister(a)
}

// Spawn adiciona e inicia um agente em tempo real (após o Start)
func (e *Engine) Spawn(a Agent) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.internalRegister(a)

	e.wg.Add(1)
	go func(agent Agent) {
		defer e.wg.Done()
		if err := agent.Run(e.ctx, e.nexus); err != nil {
			fmt.Printf("⚠️ Erro no Agente Dinâmico [%s]: %v\n", agent.ID(), err)
		}
	}(a)
}

func (e *Engine) internalRegister(a Agent) {
	// Injeta o Kernel e o Journal automaticamente se o agente for do tipo BaseAgent
	if ba, ok := a.(*agent.BaseAgent); ok {
		ba.WithJournal(e.journal)
		ba.WithMiddleware(e.kernel.KernelMiddleware())
	}

	e.agents[a.ID()] = a
}

// Start liga o motor do NexoRount. 
// Ele restaura o estado persistente, inicializa os agentes e começa a orquestração de eventos.
func (e *Engine) Start() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Se houver persistência, restauramos os eventos antes de começar
	if e.persistentNexus != nil {
		fmt.Println("💾 Restaurando estado persistente do Nexus...")
		e.persistentNexus.Restore()
	}

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

	// Listener para criação dinâmica de agentes via Nexus
	e.nexus.Subscribe("SPAWN_AGENT", func(event events.Event) {
		data := event.Payload.(map[string]interface{})
		id := data["id"].(string)
		role := data["role"].(string)
		
		fmt.Printf("🏗️  [Engine] Criando novo agente dinâmico: %s (%s)\n", id, role)
		
		newAgent := agent.New(id, role)
		if inst, ok := data["instructions"].(string); ok {
			newAgent.WithInstructions(inst)
		}
		
		e.Spawn(newAgent)
	})

	fmt.Println("✅ Todos os agentes estão em posição.")
}

func (e *Engine) Kernel() *CognitiveKernel {
	return e.kernel
}

// Publish envia um evento para o Nexus. 
// Se o Engine estiver com persistência ativa, o evento será salvo em disco antes da entrega.
func (e *Engine) Publish(event events.Event) {
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
