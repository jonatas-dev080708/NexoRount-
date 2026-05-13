package events

import (
	"context"
	"sync"
)

// Type define a categoria do evento para roteamento
type Type string

// Event é a unidade fundamental de comunicação no ecossistema
type Event struct {
	Type     Type
	Source   string
	Payload  interface{}
	ID       string
	Metadata map[string]interface{}
}

// Handler é uma função que processa um evento de forma concorrente
type Handler func(ctx context.Context, event Event)

// Bus define o contrato para qualquer barramento de eventos (Local ou Distribuído)
type Bus interface {
	Subscribe(eventType Type) chan Event
	Publish(event Event)
	Close()
}

// LocalNexus é o barramento de eventos concorrente local (canais de Go)
type LocalNexus struct {
	mu          sync.RWMutex
	subscribers map[Type][]chan Event
	closed      bool
}

func NewLocalNexus() *LocalNexus {
	return &LocalNexus{
		subscribers: make(map[Type][]chan Event),
	}
}

// Subscribe cria uma assinatura para um tipo específico de evento
// Retorna um canal por onde o agente receberá os eventos
func (n *LocalNexus) Subscribe(eventType Type) chan Event {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch := make(chan Event, 100) // Buffer para evitar bloqueio imediato
	n.subscribers[eventType] = append(n.subscribers[eventType], ch)
	return ch
}

// Publish envia um evento para todos os interessados de forma não-bloqueante
func (n *LocalNexus) Publish(event Event) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return
	}

	for _, ch := range n.subscribers[event.Type] {
		// Executamos o envio em uma goroutine ou de forma protegida
		// para garantir que um subscriber lento não trave o sistema todo
		go func(c chan Event, e Event) {
			c <- e
		}(ch, event)
	}
}

func (n *LocalNexus) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	for _, subs := range n.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
}
