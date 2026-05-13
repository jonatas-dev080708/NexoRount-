package main

import (
	"context"
	"fmt"
	"time"

	"github.com/autonome-ecosystem/framework/pkg/engine"
	"github.com/autonome-ecosystem/framework/pkg/events"
)

// --- Agente Observador ---
type ObserverAgent struct {
	id string
}

func (a *ObserverAgent) ID() string   { return a.id }
func (a *ObserverAgent) Type() string { return "OBSERVER" }

func (a *ObserverAgent) Run(ctx context.Context, nexus *events.Nexus) error {
	// Inscreve-se em todos os tipos de eventos (usando um exemplo aqui)
	tickEvents := nexus.Subscribe("HEARTBEAT")
	
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-tickEvents:
			fmt.Printf("🔍 [Observer] Percebi evento: %s vindo de %s com carga: %v\n", ev.Type, ev.Source, ev.Payload)
		}
	}
}

// --- Agente de Ritmo (Heartbeat) ---
type HeartbeatAgent struct {
	id string
}

func (a *HeartbeatAgent) ID() string   { return a.id }
func (a *HeartbeatAgent) Type() string { return "HEARTBEAT_GENERATOR" }

func (a *HeartbeatAgent) Run(ctx context.Context, nexus *events.Nexus) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			// Publica um evento no ecossistema
			nexus.Publish(events.Event{
				Type:    "HEARTBEAT",
				Source:  a.id,
				Payload: fmt.Sprintf("Pulso detectado às %s", t.Format("15:04:05")),
			})
		}
	}
}

func main() {
	// Cria o Engine
	runtime := engine.NewEngine()

	// Registra os agentes
	runtime.Register(&ObserverAgent{id: "obs-1"})
	runtime.Register(&HeartbeatAgent{id: "heart-1"})

	// Inicia o ecossistema (cada agente em sua goroutine)
	runtime.Start()

	// Mantém o sistema vivo por 10 segundos
	time.Sleep(10 * time.Second)

	// Desliga tudo graciosamente
	runtime.Stop()
}
