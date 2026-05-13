package main

import (
	"context"
	"fmt"
	"time"

	"github.com/autonome-ecosystem/framework/pkg/agent"
	"github.com/autonome-ecosystem/framework/pkg/engine"
	"github.com/autonome-ecosystem/framework/pkg/events"
	"github.com/autonome-ecosystem/framework/pkg/provider"
)

func main() {
	runtime := engine.NewEngine()
	ai := &provider.MockProvider{}

	// 1. Agente que gera o conteúdo em stream
	writer := agent.New("writer-01", "CONTENT_CREATOR").
		WithLLM(ai).
		On("START_WRITING", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Println("📝 [Writer] Começando a gerar conteúdo...")
			_, _ = a.StreamThink(ctx, "Escreva algo criativo.")
		})

	// 2. Agente de UI que reage a cada token no ecossistema
	display := agent.New("display-01", "UI").
		On("TOKEN_STREAM", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			isFinal := e.Metadata["is_final"].(bool)
			if !isFinal {
				// Imprime o token sem pular linha
				fmt.Print(e.Payload)
			} else {
				fmt.Println("\n✅ [Display] Stream finalizado.")
			}
		})

	runtime.Register(writer)
	runtime.Register(display)
	runtime.Start()

	// Dispara o processo
	runtime.Publish(events.Event{
		Type: "START_WRITING",
	})

	// Espera o streaming terminar
	time.Sleep(5 * time.Second)
	runtime.Stop()
}
