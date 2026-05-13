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

## 🧩 Arquitetura: O Organismo Operacional

No NexoRount, tudo gira em torno do **Nexus** (Nosso Sistema Nervoso). 

1. **Agentes** observam o Nexus.
2. Quando um **Evento** relevante aparece, o agente desperta.
3. O agente **Pensa** (LLM), **Recorda** (Memory) e **Age** (Tools).
4. O agente **Publica** o resultado de volta no Nexus, alimentando o ecossistema.

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
