package main

import (
	"context"
	"fmt"
	"time"

	"github.com/autonome-ecosystem/framework/pkg/agent"
	"github.com/autonome-ecosystem/framework/pkg/engine"
	"github.com/autonome-ecosystem/framework/pkg/events"
)

func main() {
	runtime := engine.NewEngine()

	// 1. Criando um Agente Pesquisador de forma elegante
	searcher := agent.New("search-01", "RESEARCHER").
		On("QUERY", func(ctx context.Context, e events.Event) {
			fmt.Printf("🔎 [%s] Pesquisando sobre: %v\n", e.Source, e.Payload)
			// Simula um delay de processamento
			time.Sleep(1 * time.Second)
			// Responde com um novo evento
			runtime.Publish(events.Event{
				Type:    "RESULT",
				Source:  "search-01",
				Payload: "Informação encontrada sobre " + fmt.Sprintf("%v", e.Payload),
			})
		})

	// 2. Criando um Agente de Interface (Simulando um usuário)
	uiAgent := agent.New("ui-01", "INTERFACE").
		On("RESULT", func(ctx context.Context, e events.Event) {
			fmt.Printf("✅ [UI] Recebi o resultado final: %v\n", e.Payload)
		})

	// Registro
	runtime.Register(searcher)
	runtime.Register(uiAgent)

	// Início
	runtime.Start()

	// Injetando um evento inicial no ecossistema
	fmt.Println("🚀 Injetando comando: 'Como funciona concorrência em Go?'")
	runtime.Publish(events.Event{
		Type:    "QUERY",
		Source:  "USER",
		Payload: "Como funciona concorrência em Go?",
	})

	time.Sleep(5 * time.Second)
	runtime.Stop()
}
