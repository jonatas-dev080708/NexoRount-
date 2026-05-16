package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
	"github.com/jonatas-dev080708/NexoRount-/pkg/tools"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// 1. Provedor Inteligente (Claude 3.5 Sonnet para raciocínio profundo)
	ai, err := provider.NewAnthropicProvider("claude-3-5-sonnet-20240620")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Inicializando ferramentas de Deep Research
	browser := tools.NewBrowserTool() // O coração da navegação profunda
	search := tools.NewGoogleSearchTool()

	// 3. Criando o Agente Deep Researcher
	researcher := agent.New("deep-researcher", "RESEARCH").
		WithLLM(ai).
		WithInstructions(`
			Você é um Deep Researcher especializado em análises exaustivas.
			Sua missão é investigar temas complexos na web.
			
			Sempre use o navegador (browser_navigate) quando encontrar links interessantes
			em uma busca para ler o conteúdo real da página.
			Seja detalhista e não aceite respostas superficiais.
		`).
		WithTool(browser).
		WithTool(search)

	runtime := engine.NewEngine()
	runtime.Register(researcher)
	runtime.Start()

	// 4. Executando uma pesquisa profunda com ReAct
	ctx := context.Background()
	question := "Quais são as últimas tendências em agentes de IA concorrentes em Go para 2026?"
	
	fmt.Printf("\n🧐 [Deep Researcher] Iniciando pesquisa profunda...\n")
	
	// O ThinkReAct permite que ele decida quando buscar e quando navegar
	resposta, err := researcher.ThinkReAct(ctx, question, 5)
	if err != nil {
		log.Printf("Erro na pesquisa: %v", err)
	}

	fmt.Printf("\n🏆 [RESPOSTA FINAL]:\n%s\n", resposta)

	// Opcional: Se a resposta for complexa, podemos passar pelo ToT
	// fmt.Println("\n🌲 Aplicando ToT para refinar o conhecimento...")
	// refina, _ := researcher.ThinkToT(ctx, "Sintetize os pontos técnicos de: " + resposta, 3)
	// fmt.Println(refina)
}
