package main

import (
	"context"
	"fmt"
	"time"

	"github.com/autonome-ecosystem/framework/pkg/agent"
	"github.com/autonome-ecosystem/framework/pkg/engine"
	"github.com/autonome-ecosystem/framework/pkg/events"
	"github.com/autonome-ecosystem/framework/pkg/provider"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	runtime := engine.NewEngine()

	// Usaremos o MockProvider para focar na lógica da ferramenta
	ai := &provider.MockProvider{}

	// 1. Criando o Agente com uma Ferramenta de Saldo
	bankAgent := agent.New("bank-bot", "FINANCE").
		WithLLM(ai).
		WithTool(&agent.Tool{
			Name:        "get_balance",
			Description: "Busca o saldo atual do cliente",
			Execute: func(accountID string) (string, error) {
				// Aqui seria uma chamada de API ou Banco de Dados real
				return fmt.Sprintf("O saldo da conta %s é R$ 5.430,20", accountID), nil
			},
		}).
		On("CHECK_BALANCE", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			accountID := fmt.Sprintf("%v", e.Payload)
			
			// O agente executa sua própria ferramenta de forma elegante
			result, err := a.Do("get_balance", accountID)
			
			if err != nil {
				fmt.Printf("❌ Erro ao executar ferramenta: %v\n", err)
				return
			}

			fmt.Printf("💰 [BankBot] Resultado da Ferramenta: %s\n", result)
			a.Emit("BALANCE_REPORT", result)
		})

	runtime.Register(bankAgent)
	runtime.Start()

	// Injetando evento para testar a ferramenta
	runtime.Publish(events.Event{
		Type:    "CHECK_BALANCE",
		Payload: "12345-X",
	})

	time.Sleep(5 * time.Second)
	runtime.Stop()
}
