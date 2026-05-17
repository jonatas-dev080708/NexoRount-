package main

import (
	"fmt"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount-/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	runtime := engine.NewEngine()

	// 1. Criando o provedor da Maritaca AI (ou OpenAI como fallback caso prefira)
	// Em produção, certifique-se de configurar as chaves de API correspondentes no ambiente.
	maritaca, err := provider.NewMaritacaProvider(provider.MaritacaSabia3)
	if err != nil {
		fmt.Printf("⚠️  Não foi possível inicializar a Maritaca AI (usando MockProvider para demonstração): %v\n", err)
	}

	var llm provider.LLMProvider
	if maritaca != nil {
		llm = maritaca
	} else {
		llm = &provider.MockProvider{}
	}

	// 2. Definindo a Ferramenta (Tool) de forma tipada
	buscaSintomas := &agent.Tool{
		Name:        "buscar_sintomas",
		Description: "Busca protocolos médicos oficiais na internet. O parâmetro deve ser o termo de pesquisa.",
		Execute: func(sintoma string) (string, error) {
			// Simulação de retorno de banco de dados médico ou API externa
			return fmt.Sprintf("Protocolo Médico para '%s': Recomendar repouso, hidratação abundante e monitorar temperatura.", sintoma), nil
		},
	}

	// 3. Inicializando o Agente de forma Declarativa e Elegante
	// O framework assume todo o controle procedural de forma automática e assíncrona.
	atendente := agent.New("medico-bot", "SAUDE").
		WithLLM(llm).
		WithInstructions("Você é um atendente de triagem médica experiente. Use a ferramenta 'buscar_sintomas' se necessário.").
		WithTool(buscaSintomas).
		WithReactiveReAct() // ✨ Toda a complexidade de Think -> Do -> Synthesize agora é nativa!

	// 4. Registrando e Iniciando o Ecossistema
	runtime.Register(atendente)
	runtime.Start()

	// Pequena pausa para garantir a prontidão do barramento de eventos
	time.Sleep(1 * time.Second)

	// 5. Testando o Ecossistema (Publicando uma pergunta no Nexus)
	fmt.Println("\n📢 [Nexus] Enviando mensagem do paciente para o barramento...")
	runtime.Nexus().Publish(events.Event{
		Type:    "NEED_RESEARCH", // Evento de gatilho que ativa o ReAct cognitivo do agente
		Payload: "Estou com febre de 38°C e dor de cabeça há 2 dias. O que devo fazer?",
	})

	// Mantém rodando por um tempo para ver as interações e depois desliga
	time.Sleep(5 * time.Second)
	runtime.Stop()
}
