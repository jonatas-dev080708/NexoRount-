package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jonatas-dev080708/NexoRount/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount/pkg/events"
	"github.com/jonatas-dev080708/NexoRount/pkg/provider"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	runtime := engine.NewEngine()
	// Tentativa de usar Gemini, fallback para Mock se falhar
	var ai provider.LLMProvider
	ai, err := provider.NewGeminiProvider(context.Background(), "gemini-1.5-flash")
	if err != nil {
		fmt.Println("⚠️  Gemini não configurado. Usando MockProvider para demonstração.")
		ai = &provider.MockProvider{}
	}

	// 1. Agente Financeiro
	finance := agent.New("finance-expert", "FINANCE").WithLLM(ai).
		On("USER_INPUT", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			prompt := fmt.Sprintf("Analise se esta mensagem é sobre dinheiro ou reembolso: '%v'. Responda apenas SIM ou NAO", e.Payload)
			decision, _ := a.Think(ctx, prompt)
			if decision == "SIM" {
				fmt.Println("💰 [Financeiro] Atuando sobre reembolso...")
				a.Emit("INTERNAL_ADVICE", "Regra: Reembolso permitido.")
			}
		})

	// 2. Agente Técnico
	tech := agent.New("tech-support", "TECHNICAL").WithLLM(ai).
		On("USER_INPUT", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			prompt := fmt.Sprintf("Analise se esta mensagem é sobre erro técnico: '%v'. Responda apenas SIM ou NAO", e.Payload)
			decision, _ := a.Think(ctx, prompt)
			if decision == "SIM" {
				fmt.Println("🛠️ [Técnico] Atuando sobre erro de sistema...")
				a.Emit("INTERNAL_ADVICE", "Suporte: Pedir para limpar cache.")
			}
		})

	// 3. Agente Atendente Consolidador
	attendant := agent.New("main-attendant", "ORCHESTRATOR").WithLLM(ai).
		On("USER_INPUT", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("🤔 [Atendente] Pensando na resposta final para: %v\n", e.Payload)
			
			// Em um sistema real, aqui ele esperaria pelos INTERNAL_ADVICE
			// Para o exemplo, vamos apenas gerar a resposta final
			response, _ := a.Think(ctx, fmt.Sprintf("O usuário disse: %v. Responda de forma curta ajudando com reembolso e erro técnico.", e.Payload))
			
			fmt.Printf("\n🤖 RESPOSTA FINAL AO CLIENTE:\n%s\n", response)
		})

	runtime.Register(finance)
	runtime.Register(tech)
	runtime.Register(attendant)

	runtime.Start()

	fmt.Println("\n--- ENTRADA: 'App crashou ao pedir reembolso' ---")
	runtime.Publish(events.Event{
		Type: "USER_INPUT",
		Payload: "App crashou ao pedir reembolso",
	})

	time.Sleep(10 * time.Second)
	runtime.Stop()
}
