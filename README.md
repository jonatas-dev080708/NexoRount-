# NexoRount 🚀
### O Runtime Concorrente para Ecossistemas de Agentes Autônomos em Go

O NexoRount não é apenas um "wrapper" para LLMs ou um framework de pipelines sequenciais. Ele é um **Runtime Concorrente** desenhado para criar sistemas autônomos vivos, reativos e escaláveis.

Ao contrário de abordagens tradicionais baseadas em grafos rígidos ou fluxos lineares, o NexoRount trata agentes como entidades independentes que coexistem, colaboram e evoluem em tempo real através de uma arquitetura orientada a eventos (Event-Driven).

---

## 🌪️ Por que NexoRount?

A maioria dos frameworks atuais (LangChain, LangGraph) foi influenciada pelo mundo de notebooks e pesquisa. Eles funcionam bem para sequências simples, mas falham em sistemas complexos que exigem:

- **Concorrência Real:** Milhares de agentes operando simultaneamente sem travar o sistema.
- **Reatividade:** Agentes que decidem agir baseados em eventos do ambiente, não em fluxogramas pré-definidos.
- **Escalabilidade Distribuída:** Orquestração nativa entre múltiplos servidores.
- **Memória Evolutiva:** Integração profunda com pgvector para aprendizado contínuo (RAG).

---

## ✨ Principais Características

- [x] **Engine Concorrente:** Ciclo de vida de agentes gerenciado por Goroutines.
- [x] **Event Nexus:** Barramento de eventos ultra-rápido (Local e NATS).
- [x] **Multi-LLM:** Suporte nativo para OpenAI, Google Gemini e Provedores Customizados.
- [x] **Streaming Nativo:** Tokens fluem pelo ecossistema como eventos em tempo real.
- [x] **Hybrid Memory:** Persistência local (JSON) e semântica (PostgreSQL + pgvector).
- [x] **Ação via Tools:** Sistema elegante para dar "mãos" aos seus agentes.
- [x] **ThinkParallel:** API nativa para execução simultânea de múltiplos pensamentos.
- [x] **Middlewares de IA:** Interceptores globais para segurança, log e cache.
- [x] **Pre-fetching:** Enriquecimento automático de contexto antes da ativação do agente.

---

## 🚀 Começo Rápido

### Instalando
```bash
go get github.com/jonatas-dev080708/NexoRount
```

### Exemplo: Criando um Agente SDR Reativo

```go
package main

import (
	"github.com/jonatas-dev080708/NexoRount/pkg/agent"
	"github.com/jonatas-dev080708/NexoRount/pkg/engine"
	"github.com/jonatas-dev080708/NexoRount/pkg/provider"
)

func main() {
	runtime := engine.NewEngine()
	ai, _ := provider.NewOpenAIProvider("SUA_CHAVE")

	sdr := agent.New("sdr-01", "SALES").
		WithLLM(ai).
		WithInstructions("Você é um SDR focado em qualificação de leads.").
		On("NEW_LEAD", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
			// Pensamento e Ação Concorrente
			analise, _ := a.Think(ctx, "Qualifique: " + e.Payload.(string))
			a.Emit("LEAD_QUALIFIED", analise)
		})

	runtime.Register(sdr)
	runtime.Start()
}
```

---

## 🧠 Superpoder: Multitarefa Interna (O Cérebro Paralelo)

Diferente de frameworks baseados em grafos lineares, o NexoRount permite que um único agente execute múltiplas linhas de pensamento simultaneamente usando a concorrência nativa do Go.

```go
atendente.On("WPP_MESSAGE", func(a *agent.BaseAgent, ctx context.Context, e events.Event) {
    // Dispara 3 pensamentos simultâneos no "subconsciente" do agente
    resultados, _ := a.ThinkParallel(ctx, []string{
        "Analise o histórico deste paciente",
        "Busque protocolos clínicos relevantes",
        "Verifique sinais de emergência",
    })

    // Consolida tudo e responde
    resposta, _ := a.Think(ctx, "Com base nos resultados, responda ao paciente...")
    a.Emit("RESPOSTA_FINAL", resposta)
})
```

---

## 🛠️ Recursos Avançados

### 1. Middlewares (Observabilidade e Segurança)
Adicione camadas de controle em todas as chamadas de IA.
```go
agente.WithMiddleware(func(a *agent.BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error) {
    return func(ctx context.Context, prompt string) (string, error) {
        // Log, Validação ou Cache aqui
        return next(ctx, prompt)
    }
})
```

### 2. Pre-fetchers (Otimização de Contexto)
Prepare dados em paralelo assim que o evento chega ao Nexus.
```go
agente.WithPreFetcher("NEW_LEAD", func(a *agent.BaseAgent, ctx context.Context, e events.Event) (string, error) {
    // Busca dados no CRM antes mesmo do agente 'acordar'
    return "Dados do CRM: ...", nil
})
```

---

## 🧩 Arquitetura: O Organismo Operacional

No NexoRount, tudo gira em torno do **Nexus** (Nosso Sistema Nervoso). 

1. **Eventos** são publicados no **Nexus**.
2. **Pre-fetchers** enriquecem o evento com dados externos em paralelo.
3. **Agentes** despertam e executam suas **Handlers** reativos.
4. **ThinkParallel** permite múltiplos pensamentos simultâneos.
5. **Middlewares** interceptam e validam a comunicação com a LLM.
6. O agente **Publica** o resultado ou executa **Tools**, reiniciando o ciclo.

---

## 🛠️ Filosofia Go-First

O NexoRount aproveita o que Go faz de melhor:
- **Simplicidade:** API fluida e declarativa.
- **Performance:** Goroutines para milhares de agentes.
- **Escalabilidade:** NATS para orquestração distribuída.
- **Robustez:** Tipagem estática e segurança em concorrência.

---

## 📄 Licença
Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.

---

**Construído para quem acredita que agentes não são apenas prompts, são sistemas distribuídos vivos.** 🚀
