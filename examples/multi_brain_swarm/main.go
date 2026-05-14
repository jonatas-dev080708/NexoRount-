package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	// 1. Inicializando os dois "cérebros"
	gemini, err := provider.NewGeminiProvider(ctx, "gemini-1.5-flash")
	if err != nil {
		log.Printf("Aviso: Gemini não configurado: %v", err)
	}

	openaiProv, err := provider.NewOpenAIProvider("gpt-3.5-turbo")
	if err != nil {
		log.Printf("Aviso: OpenAI não configurada: %v", err)
	}

	runtime := engine.NewEngine()

	// 2. Agente Analista (Usa Gemini)
	analyst := agent.New("analyst-01", "RESEARCHER").
		WithLLM(gemini).
		On("TOPIC_RECEIVED", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("🔭 [Analista - Gemini] Analisando: %v\n", e.Payload)
			
			prompt := fmt.Sprintf("Faça uma análise técnica detalhada em 3 pontos sobre: %v", e.Payload)
			analysis, err := a.Think(ctx, prompt)
			if err != nil {
				fmt.Printf("❌ Erro no Analista: %v\n", err)
				return
			}

			// Publica a análise para que outros agentes possam refinar
			a.Emit("ANALYSIS_READY", analysis)
		})

	// 3. Agente Refinador (Usa OpenAI)
	refiner := agent.New("refiner-01", "EDITOR").
		WithLLM(openaiProv).
		On("ANALYSIS_READY", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Println("✍️ [Refinador - OpenAI] Refinando a análise do Gemini...")
			
			prompt := fmt.Sprintf("Resuma esta análise técnica em um parágrafo executivo e impactante: %v", e.Payload)
			refined, err := a.Think(ctx, prompt)
			if err != nil {
				fmt.Printf("❌ Erro no Refinador: %v\n", err)
				return
			}

			a.Emit("FINAL_REPORT", refined)
		})

	// 4. Agente de Display
	display := agent.New("display", "UI").
		On("FINAL_REPORT", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("\n✨ RELATÓRIO FINAL CONSOLIDADO:\n%s\n", e.Payload)
		})

	runtime.Register(analyst)
	runtime.Register(refiner)
	runtime.Register(display)

	runtime.Start()

	// Inicia a cadeia de inteligência
	runtime.Publish(events.Event{
		Type:    "TOPIC_RECEIVED",
		Source:  "USER",
		Payload: "O impacto da computação quântica na criptografia moderna",
	})

	time.Sleep(15 * time.Second)
	runtime.Stop()
}
