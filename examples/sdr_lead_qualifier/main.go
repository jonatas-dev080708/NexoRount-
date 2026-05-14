package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

func main() {
	runtime := engine.NewEngine()
	
	// Usaremos o MockProvider para demonstração rápida
	ai := &provider.MockProvider{}

	// 1. Definindo o Agente SDR com sua Persona e Ferramentas
	sdr := agent.New("sdr-01", "SALES").
		WithLLM(ai).
		WithInstructions(`
			Você é um SDR (Sales Development Representative) de uma empresa de tecnologia.
			Seu objetivo é qualificar leads. 
			Se o lead tiver uma empresa com mais de 10 funcionários, ele é QUALIFICADO.
			Se for qualificado, use a ferramenta 'book_meeting'.
			Sempre seja profissional e persuasivo.
		`).
		// Ferramenta de agendamento
		WithTool(&agent.Tool{
			Name:        "book_meeting",
			Description: "Agenda uma reunião com o time de vendas",
			Execute: func(leadInfo string) (string, error) {
				return "✅ Reunião agendada com sucesso para o lead: " + leadInfo, nil
			},
		}).
		// Reação ao receber um novo lead
		On("NEW_LEAD", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			leadMsg := e.Payload.(string)
			fmt.Printf("\n🚀 [SDR] Novo lead recebido: %s\n", leadMsg)

			// Passo 1: Pensar e Qualificar (LLM)
			fmt.Println("🤔 [SDR] Qualificando lead...")
			analysis, _ := a.Think(ctx, "Analise este lead e decida se é QUALIFICADO ou NAO: "+leadMsg)
			
			// Passo 2: Se qualificado, agir (Tool)
			if analysis == "SIM" || analysis == "QUALIFICADO" { // Mock retorna SIM por padrão
				fmt.Println("🎯 [SDR] Lead qualificado! Agendando reunião...")
				res, _ := a.Do("book_meeting", leadMsg)
				a.Emit("MEETING_BOOKED", res)
			}

			// Passo 3: Aprender (Persistência Vetorial/RAG)
			// No mundo real, aqui ele salvaria no pgvector
			fmt.Println("💾 [SDR] Salvando perfil do lead na memória semântica...")
			_ = a.Remember("last_lead_analysis", analysis)

			// Passo 4: Responder ao ecossistema
			a.Emit("SDR_RESPONSE", "Processamento de lead concluído.")
		})

	// 2. Agente de Log (Observador)
	logger := agent.New("logger", "MONITOR").
		On("MEETING_BOOKED", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("📢 [NOTIFICAÇÃO] %v\n", e.Payload)
		})

	runtime.Register(sdr)
	runtime.Register(logger)
	runtime.Start()

	// 3. Simulando a chegada de um Lead
	runtime.Publish(events.Event{
		Type:    "NEW_LEAD",
		Payload: "Olá, sou o Diretor da TechCorp, temos 50 funcionários e queremos automatizar processos.",
	})

	time.Sleep(5 * time.Second)
	runtime.Stop()
}
