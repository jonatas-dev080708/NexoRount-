package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/memory"
)

func main() {
	// 1. Criamos um armazenamento de memória em disco (pasta ./db)
	store, err := memory.NewFileStore("./db")
	if err != nil {
		log.Fatal(err)
	}

	runtime := engine.NewEngine()

	// 2. Agente que reconhece usuários
	greeter := agent.New("greeter-01", "ASSISTANT").
		WithMemory(store).
		On("USER_HELLO", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			userName := fmt.Sprintf("%v", e.Payload)

			// Tenta "Recordar" se já conhece este usuário
			lastSeen, ok := a.Recall("last_user")
			
			if ok {
				fmt.Printf("🤖 [Greeter] Olá de novo, %s! (Eu lembro que o último que passou por aqui foi %v)\n", userName, lastSeen)
			} else {
				fmt.Printf("🤖 [Greeter] Prazer em te conhecer, %s! Vou guardar seu nome na minha memória.\n", userName)
			}

			// "Lembra" do usuário atual para a próxima vez
			a.Remember("last_user", userName)
		})

	runtime.Register(greeter)
	runtime.Start()

	// Simula duas interações
	fmt.Println("--- Primeira Interação ---")
	runtime.Publish(events.Event{Type: "USER_HELLO", Payload: "Alice"})
	time.Sleep(2 * time.Second)

	fmt.Println("\n--- Segunda Interação ---")
	runtime.Publish(events.Event{Type: "USER_HELLO", Payload: "Bob"})
	
	time.Sleep(2 * time.Second)
	runtime.Stop()

	fmt.Println("\n📂 Verifique a pasta './db' para ver o arquivo JSON de memória do agente!")
}
