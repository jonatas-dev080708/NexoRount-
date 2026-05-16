package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/loader"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
	"github.com/jonatas-dev080708/NexoRount-/pkg/tools"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// 1. Configurando os Provedores
	gemini, _ := provider.NewGeminiProvider(context.Background(), "gemini-1.5-flash")
	claude, _ := provider.NewAnthropicProvider("claude-3-5-sonnet-20240620")

	// 2. Criando o Router (Alterna entre modelos conforme a necessidade)
	// Começamos com o "Smart" (Claude) para análise complexa
	router := provider.NewRouterProvider(gemini, claude).WithStrategy(provider.StrategySmart)

	// 3. Preparando Ferramentas
	sqlTool := tools.NewSQLQueryTool(tools.SQLConfig{
		Driver: "pgx",
		DSN:    os.Getenv("DATABASE_URL"),
	})
	httpTool := tools.NewHTTPGetTool()

	// 4. Inicializando o Engine
	runtime := engine.NewEngine()

	// 5. Definindo o Agente Analista de Dados
	analista := agent.New("analista-pro", "BI").
		WithLLM(router).
		WithInstructions(`
			Você é um Analista de BI Sênior.
			Seu objetivo é cruzar dados de documentos PDF com consultas ao banco de dados SQL.
			Seja preciso e apresente conclusões baseadas em fatos.
		`).
		WithTool(sqlTool).
		WithTool(httpTool)

	// Handler de exemplo: Processar Relatório
	analista.On("PROCESS_REPORT", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
		pdfPath := e.Payload.(string)
		fmt.Printf("📂 [Analista] Lendo relatório PDF: %s\n", pdfPath)

		// Usando o novo PDFLoader
		pdfLoader := loader.NewPDFLoader(pdfPath)
		docs, err := pdfLoader.Load(ctx)
		if err != nil {
			fmt.Printf("❌ Erro ao ler PDF: %v\n", err)
			return
		}

		// O agente analisa o conteúdo do PDF e decide se precisa consultar o SQL
		prompt := fmt.Sprintf("Analise este conteúdo de relatório: %s. \nSe encontrar menção a IDs de clientes, use a ferramenta sql_query para buscar o faturamento deles.", docs[0].Content)
		
		res, _ := a.Think(ctx, prompt)
		fmt.Printf("\n🤖 [Conclusão do Analista]: %s\n", res)
	})

	runtime.Register(analista)
	runtime.Start()

	// Simulando o disparo de um processo
	// runtime.Publish(events.Event{Type: "PROCESS_REPORT", Payload: "relatorio_vendas.pdf"})

	fmt.Println("🚀 Engine Analista de Dados rodando...")
	// Mantém rodando para demonstração
	select {}
}
