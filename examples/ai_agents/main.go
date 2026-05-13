package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/autonome-ecosystem/framework/pkg/agent"
	"github.com/autonome-ecosystem/framework/pkg/engine"
	"github.com/autonome-ecosystem/framework/pkg/events"
	"github.com/autonome-ecosystem/framework/pkg/provider"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Carrega as chaves do arquivo .env
	_ = godotenv.Load()

	// 2. Inicializa o provedor de IA (OpenAI como exemplo)
	ai, err := provider.NewOpenAIProvider("gpt-3.5-turbo")
	if err != nil {
		log.Fatalf("Erro ao iniciar provedor: %v. Certifique-se de configurar a OPENAI_API_KEY.", err)
	}

	runtime := engine.NewEngine()

	// 3. Cria um agente que usa a IA de forma elegante
	processor := agent.New("ai-processor", "BRAIN").
		WithLLM(ai).
		On("USER_MESSAGE", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("🧠 [BRAIN] Analisando mensagem via %s...\n", ai.Name())
			
			// Uso direto do .Think() e .Emit() através da instância 'a'
			prompt := fmt.Sprintf("Responda de forma curta: %v", e.Payload)
			response, err := a.Think(ctx, prompt)
			
			if err != nil {
				fmt.Printf("❌ Erro no processamento: %v\n", err)
				return
			}

			fmt.Printf("💡 [BRAIN] Resposta: %s\n", response)
			a.Emit("AI_RESPONSE", response)
		})

	// Agente de Interface para mostrar o resultado
	ui := agent.New("ui-logger", "UI").
		On("AI_RESPONSE", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("🖥️  [DISPLAY] %v\n", e.Payload)
		})

	runtime.Register(processor)
	runtime.Register(ui)
	runtime.Start()

	// Injetando evento
	runtime.Publish(events.Event{
		Type:    "USER_MESSAGE",
		Source:  "CLIENT",
		Payload: "Qual a cor do céu?",
	})

	time.Sleep(5 * time.Second)
	runtime.Stop()
}
