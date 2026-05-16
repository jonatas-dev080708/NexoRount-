package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

// LeadQualification representa a estrutura de saída para a qualificação do lead
type LeadQualification struct {
	Score      int    `json:"score"`
	Status     string `json:"status"` // "QUALIFIED", "UNQUALIFIED", "NURTURE"
	Reason     string `json:"reason"`
	Industry   string `json:"industry"`
	CompanySize string `json:"company_size"`
}

func main() {
	runtime := engine.NewEngine()
	
	// Configurando Resiliência: MultiProvider (Mock + Simulação de outro)
	// Em produção seria: provider.NewMultiProvider(openaiProv, geminiProv)
	ai := provider.NewMultiProvider(
		&provider.MockProvider{}, 
		&provider.MockProvider{}, // Fallback para outro mock
	)

	// 1. Definindo o Agente SDR com as novas capacidades
	sdr := agent.New("sdr-enterprise", "SALES").
		WithLLM(ai).
		WithInstructions(`
			Você é um SDR de Elite especializado em prospecção B2B.
			Seu objetivo é qualificar leads com precisão cirúrgica.
			Criterios de Qualificação:
			- Empresa > 10 funcionários.
			- Interesse em automação ou IA.
			- Decisor identificado (Diretor, CEO, Manager).
		`).
		// Adicionando Segurança
		WithMiddleware(agent.SafetyMiddleware()).
		// Ferramenta de agendamento
		WithTool(&agent.Tool{
			Name:        "book_meeting",
			Description: "Agenda uma reunião estratégica com o time de vendas",
			Execute: func(leadInfo string) (string, error) {
				return "📅 [Agenda] Reunião estratégica marcada para o lead: " + leadInfo, nil
			},
		}).
		// Reação Reativa e Concorrente
		On("NEW_LEAD", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			leadMsg := e.Payload.(string)
			fmt.Printf("\n🚀 [SDR-ELITE] Analisando novo lead: %s\n", leadMsg)

			// PASSO 1: Análise Paralela (Diferentes ângulos)
			// O agente explora caminhos divergentes simultaneamente
			fmt.Println("🧠 [SDR] Iniciando análise concorrente (ThinkParallel)...")
			reflexoes, _ := a.ThinkParallel(ctx, []string{
				"Qual o tamanho da empresa mencionada?",
				"Qual o nível de autoridade da pessoa que entrou em contato?",
				"Existe um sinal claro de dor ou necessidade técnica?",
			})

			for i, r := range reflexoes {
				fmt.Printf("   💡 Reflexão %d: %s\n", i+1, r)
			}

			// PASSO 2: Qualificação Estruturada (ThinkJSON)
			// Consolidando as reflexões em um objeto JSON tipado
			fmt.Println("📊 [SDR] Consolidando qualificação estruturada (ThinkJSON)...")
			var qual LeadQualification
			promptQual := fmt.Printf("Com base nessas reflexões: %v, qualifique o lead: %s", reflexoes, leadMsg)
			
			// MockProvider retornará um JSON válido se format=json_object for passado (simulado)
			_ = a.ThinkJSON(ctx, fmt.Sprintf("Consolide a qualificação em JSON: %s", promptQual), &qual)
			
			// Simulando preenchimento do JSON pelo Mock para demonstração
			if qual.Status == "" {
				qual = LeadQualification{
					Score:       95,
					Status:      "QUALIFIED",
					Reason:      "Empresa grande com necessidade explícita de automação.",
					Industry:    "Tecnologia",
					CompanySize: "50+ funcionários",
				}
			}

			fmt.Printf("✅ [SDR] Resultado: [%s] Score: %d | Motivo: %s\n", qual.Status, qual.Score, qual.Reason)

			// PASSO 3: Tomada de Decisão Baseada em Dados
			if qual.Status == "QUALIFIED" {
				fmt.Println("🎯 [SDR] Lead Alta Prioridade! Executando agendamento...")
				res, _ := a.Do("book_meeting", leadMsg)
				a.Emit("MEETING_BOOKED", res)
			}

			// PASSO 4: Persistência com Versionamento
			// Usando o novo modelo de memória com trava otimista
			fmt.Println("💾 [SDR] Persistindo perfil versionado na memória...")
			_ = a.RememberVersioned("lead_profile_"+qual.Industry, qual, 1)

			a.Emit("SDR_COMPLETED", map[string]interface{}{
				"lead":   leadMsg,
				"result": qual,
			})
		})

	// 2. Agente Monitor (Observabilidade)
	monitor := agent.New("monitor-01", "OPS").
		On("MEETING_BOOKED", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			fmt.Printf("📢 [NOTIFICAÇÃO OPS] %v\n", e.Payload)
		}).
		On("AGENT_STATE_CHANGED", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			data := e.Payload.(map[string]string)
			fmt.Printf("📉 [STATUS] Agente %s mudou de %s para %s\n", data["agent_id"], data["old_state"], data["new_state"])
		})

	runtime.Register(sdr)
	runtime.Register(monitor)
	runtime.Start()

	// 3. Simulando Lead de Alto Valor
	runtime.Publish(events.Event{
		Type:    "NEW_LEAD",
		Payload: "Diretor da GlobalTech aqui. Temos 200 funcionários e precisamos migrar nossos processos para agentes de IA concorrentes.",
	})

	time.Sleep(6 * time.Second)
	runtime.Stop()
	
	finalJSON, _ := json.MarshalIndent(runtime, "", "  ")
	_ = finalJSON // Apenas para debug se necessário
}
