package observability

import (
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// Span representa uma unidade de trabalho ou pensamento no tempo
type Span struct {
	ID        string
	TraceID   string
	Name      string
	Source    string
	StartTime int64
	EndTime   int64
	Metadata  map[string]interface{}
}

// Journal é o componente responsável por registrar a história do sistema
type Journal struct {
	mu     sync.RWMutex
	events []events.Event
	spans  []Span
}

func NewJournal() *Journal {
	return &Journal{
		events: make([]events.Event, 0),
		spans:  make([]Span, 0),
	}
}

// RecordEvent salva um evento no histórico para inspeção posterior
func (j *Journal) RecordEvent(e events.Event) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	// Garante que o evento tenha um timestamp se não tiver
	if e.Timestamp == 0 {
		e.Timestamp = time.Now().UnixNano()
	}
	
	j.events = append(j.events, e)
	fmt.Printf("🔍 [OBSERVER] Evento: %s | Source: %s | TraceID: %s\n", e.Type, e.Source, e.TraceID)
}

// StartSpan inicia o rastreio de uma operação
func (j *Journal) StartSpan(traceID, name, source string) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	spanID := fmt.Sprintf("span-%d", time.Now().UnixNano())
	span := Span{
		ID:        spanID,
		TraceID:   traceID,
		Name:      name,
		Source:    source,
		StartTime: time.Now().UnixNano(),
		Metadata:  make(map[string]interface{}),
	}
	j.spans = append(j.spans, span)
	return spanID
}

// EndSpan finaliza o rastreio de uma operação
func (j *Journal) EndSpan(spanID string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	for i := range j.spans {
		if j.spans[i].ID == spanID {
			j.spans[i].EndTime = time.Now().UnixNano()
			return
		}
	}
}

// GetTraceHistory retorna todos os eventos vinculados a uma jornada específica
func (j *Journal) GetTraceHistory(traceID string) []events.Event {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	var history []events.Event
	for _, e := range j.events {
		if e.TraceID == traceID {
			history = append(history, e)
		}
	}
	return history
}
