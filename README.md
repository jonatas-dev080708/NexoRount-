# NexoRount 🚀: O Guia Definitivo do Runtime Agentico
### Sistema Operacional Cognitivo para Ecossistemas de Inteligência Distribuída

---

## 📑 Sumário

1. [Visão Geral e Filosofia](#-visão-geral-e-filosofia)
2. [Arquitetura de Sistemas](#-arquitetura-de-sistemas)
3. [Instalação e Configuração](#-instalação-e-configuração)
4. [O Cérebro: pkg/agent](#-o-cérebro-pkgagent)
5. [O Sistema Nervoso: pkg/events](#-o-sistema-nervoso-pkgevents)
6. [O Kernel: pkg/engine](#-o-kernel-pkgengine)
7. [A Memória Hierárquica: pkg/memory](#-a-memória-hierárquica-pkgmemory)
8. [Observabilidade: pkg/observability](#-observabilidade-pkgobservability)
9. [Provedores de IA: pkg/provider](#-provedores-de-ia-pkgprovider)
10. [Tutorial: Criando seu Primeiro Agente Multitarefa](#-tutorial-passo-a-passo)
11. [Padrões Avançados (Enterprise Patterns)](#-padrões-avançados)
12. [Performance e Escalabilidade](#-performance-e-escalabilidade)
13. [Solução de Problemas (Troubleshooting)](#-solução-de-problemas)
14. [NexoDashboard (Fase 5)](#-observabilidade-pkgobservability)
15. [Checkpoints e Durabilidade (Persistent Nexus)](#-checkpoints-e-durabilidade)
16. [Sintaxe Moderna e DX (Fluent API & Generics)](#-sintaxe-moderna-e-dx)
17. [Evolução e Roadmap](#-contribuição-e-roadmap)

---

## 🌪️ 1. Visão Geral e Filosofia

O **NexoRount** não foi criado para ser apenas mais uma biblioteca de agentes. Ele nasceu de uma insatisfação profunda com os frameworks atuais (LangChain, LangGraph) que, apesar de poderosos, carregam a herança de sistemas determinísticos e síncronos da era pré-IA.

### A Crise dos Grafos
A maioria dos frameworks hoje tenta forçar a inteligência em um "fluxograma". Mas a inteligência humana não é um grafo rígido de nós e arestas. Ela é um **processo biológico concorrente**. Enquanto você fala, seu cérebro está processando visão, memória, emoção e lógica ao mesmo tempo. 

### A Resposta: Reatividade Pura
No NexoRount, abandonamos o conceito de "Chains" e "Graphs" em favor do **Organismo Reativo**. Nossos agentes são processos vivos que habitam um barramento de eventos (o Nexus). Eles não esperam que alguém os chame; eles **despertam** quando o ambiente sinaliza algo relevante.

### Por que Go?
A escolha do Go não foi estética. Para um runtime agentico de verdade, você precisa de:
- **Concorrência Massiva:** Goroutines permitem rodar 100.000 agentes em um laptop.
- **Transparência de Localização:** O modelo de canais do Go se traduz perfeitamente para sistemas distribuídos (NATS).
- **Performance:** Redução drástica de latência em relação ao Python, permitindo que a IA gaste tempo "pensando" e não esperando o interpretador.

---

## 🧠 2. Arquitetura de Sistemas

O NexoRount é construído como um **Cognitive Operating System**. Ele gerencia os "recursos de inteligência" (LLMs) da mesma forma que um kernel gerencia ciclos de CPU.

### Camada 1: O Nexus (Comunicação)
Tudo no NexoRount é um evento. O Nexus é o meio físico por onde esses eventos viajam. Ele é agnóstico à localização: pode ser um canal em memória ou um cluster global de NATS.

### Camada 2: O Kernel (Escalonamento)
O Kernel impede o caos. Ele gerencia a "Fadiga" dos agentes, as prioridades e o orçamento (budget) de tokens. Sem o Kernel, um ecossistema concorrente colapsaria rapidamente sob seu próprio peso cognitivo.

### Camada 3: A Memória (Hierarquia)
O sistema de memória segue o modelo biológico:
1. **Trabalho (Working):** Rápida e volátil.
2. **Curto Prazo (Short-Term):** Contexto da conversa.
3. **Longo Prazo (Long-Term):** Identidade e fatos.
4. **Semântica (Semantic):** Conhecimento vetorial (RAG).
5. **Episódica (Episodic):** O histórico vivido.

---

## ⚙️ 3. Instalação e Configuração

### Requisitos
- Go 1.21 ou superior.
- PostgreSQL com extensão pgvector (opcional, para memória semântica).
- NATS Server (opcional, para runtime híbrido/distribuído).

### Instalando o Framework
```bash
go get github.com/jonatas-dev080708/NexoRount-
```

### Configuração Inicial
O framework utiliza variáveis de ambiente para chaves de API:
```bash
export OPENAI_API_KEY="sua_chave"
export GEMINI_API_KEY="sua_chave"
export DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

---

## 🤖 4. O Cérebro: pkg/agent

Este é o pacote mais utilizado pelo desenvolvedor. Ele contém a definição do **BaseAgent**.

### Estrutura do Agente
O `BaseAgent` não é apenas um System Prompt. Ele é uma máquina de estados:
- **StateSleeping:** Ocioso, economizando recursos.
- **StateThinking:** Ativo, consumindo recursos de IA.
- **StateWaiting:** Aguardando uma ferramenta externa ou input.
- **StateFailed:** Erro crítico em recuperação.

### O Método `On`
O registro de comportamento é declarativo:
```go
meuAgente.On("TIPO_DE_EVENTO", func(a *BaseAgent, ctx context.Context, e Event) {
    // Reação lógica
})
```

### ⚡ Novo: Bind Genérico (Safe Typing)
Elimine o `interface{}` e use tipos fortes para seus eventos:
```go
type Lead struct { Nome string }

agent.Bind(meuAgente, "NEW_LEAD", func(a *BaseAgent, ctx context.Context, p Lead) {
    fmt.Println(p.Nome) // Tipagem garantida pelo compilador
})
```

### Multitarefa Interna: `ThinkParallel`
Este é o recurso mais poderoso para agentes complexos. Ele permite que o agente explore caminhos de raciocínio divergentes antes de convergir para uma resposta.
```go
res, err := a.ThinkParallel(ctx, []string{"Analise X", "Compare com Y"})
```

### Middlewares
Os middlewares permitem injetar lógica em cada chamada de `Think`:
- **LogMiddleware:** Rastreia o prompt e a resposta.
- **SafetyMiddleware:** Filtra conteúdo inadequado.
- **BudgetMiddleware:** Controla o gasto financeiro em tempo real.

---

## ⚡ 5. O Sistema Nervoso: pkg/events

O pacote `events` define como a informação flui.

### A Interface `Bus`
Qualquer coisa que implemente `Publish` e `Subscribe` pode ser o Nexus.
```go
type Bus interface {
    Subscribe(eventType Type) chan Event
    Publish(event Event)
}
```

### LocalNexus vs NATSNexus
- **Local:** Usa `channels` do Go. É imbatível em velocidade para aplicações single-binary.
- **NATS:** Usa `JetStream` para persistência e distribuição. Ideal para arquiteturas de microsserviços.

### 💾 Novo: Persistent Nexus (Checkpoints Estilo LangGraph)
O NexoRount agora possui durabilidade nativa de eventos. Se o processo cair, os eventos não processados são restaurados automaticamente.
```go
engine.NewEngine().
    WithPersistence("backlog.db"). // Usa SQLite para imortalidade de eventos
    Start()
```

### O Bridge Híbrido
O `Bridge` permite conectar dois barramentos. Imagine um agente rodando em uma Raspberry Pi (Edge) que se conecta via Bridge ao Nexus principal rodando na Cloud.

---

## 🖥️ 6. O Kernel: pkg/engine

O Engine é o coração do runtime. É ele quem "carrega" os agentes e os coloca para rodar.

### O Gerenciamento de Fadiga
O Kernel evita que um agente seja sobrecarregado por eventos simultâneos. Você pode configurar a "Capacidade Mental" de cada agente.
```go
engine.Kernel().SetAgentCapacity("agente-01", 5) // Máximo 5 pensamentos paralelos
```

### O Diário de Bordo (Journal)
Integrado na Observabilidade, o Journal grava a sequência causal de eventos. Se o agente agendou uma reunião errada, o Journal permite ver exatamente qual evento desencadeou aquela decisão.

---

## 💾 7. A Memória Hierárquica: pkg/memory

O NexoRount resolve o problema do "esquecimento" da IA com uma arquitetura de múltiplas camadas.

### Camada 1: Working Memory
Implementada como um `map[string]interface{}` volátil no agente. Use para flags de controle de fluxo de pensamento.

### Camada 2: Short-Term (Session)
Persistida via `FileStore` ou `KV Store`. Armazena o contexto imediato da interação atual.

### Camada 3: Long-Term (Optimistic Locking)
A maior inovação aqui é o **Versioning**.
- Ao ler um dado: `val, version, _ := a.RecallVersioned("perfil")`
- Ao salvar: `err := a.RememberVersioned("perfil", novoVal, version)`
- Se outra Goroutine alterou o perfil enquanto o agente "pensava", a versão não baterá e o agente será forçado a reavaliar. Isso evita o caos em escritas simultâneas.

### Camada 4: Semantic (pgvector)
Usa busca vetorial (Cosine Similarity) para RAG. O agente busca conhecimento em manuais ou bases de dados gigantes.

### Camada 5: Episodic (Journal)
O agente pode consultar seu próprio passado: "O que eu fiz quando o usuário reclamou na semana passada?".

---

## 🕵️ 8. Observabilidade: pkg/observability

Desenvolver agentes sem observabilidade é como voar às cegas.

### Journaling de Eventos
Cada evento ganha um `TraceID` único. Isso permite criar uma "linhagem" do evento. 
Exemplo de Trace:
`LeadChegou -> AgenteSDR_Iniciado -> BuscaHistorico_Span -> Think_Span -> RespostaEmitida`

### Spans e Tracing
Usamos o conceito de Spans para medir o tempo de cada etapa. Você pode identificar gargalos: "Por que a busca vetorial está levando 2 segundos?".

---

## 🤖 9. Provedores de IA: pkg/provider

O framework é agnóstico à LLM.

### Provedores Nativos
- **OpenAI:** GPT-4, GPT-3.5.
- **Gemini:** Google Gemini Pro e Flash (Altamente recomendado para alta performance e custo baixo).
- **Mock:** Para testes automatizados sem gastar créditos.

### Streaming
O NexoRount trata o streaming de tokens como eventos no Nexus. Outros agentes podem "ouvir" o pensamento enquanto ele está sendo gerado.

---

## 🚀 10. Tutorial: Criando seu Primeiro Agente Multitarefa

Neste tutorial, vamos construir um **Agente de Atendimento Médico** que faz triagem, consulta protocolos e verifica segurança simultaneamente.

### Passo 1: Configuração do Engine
```go
runtime := engine.NewEngine()
provider, _ := provider.NewGeminiProvider("gemini-1.5-flash")
```

### Passo 2: Definindo o Agente
```go
atendente := agent.New("medico-bot", "SAUDE").
    WithLLM(provider).
    WithInstructions("Você é um atendente de triagem médica.")
```

### Passo 3: Implementando a Multitarefa
```go
atendente.On("MENSAGEM_PACIENTE", func(a *BaseAgent, ctx context.Context, e Event) {
    msg := e.Payload.(string)
    
    // Dispara 3 linhas de pensamento paralelas
    reflexoes, _ := a.ThinkParallel(ctx, []string{
        "Quais são os sinais vitais mencionados?",
        "Consulte protocolos para estes sintomas.",
        "Existe risco de vida imediato?",
    })
    
    // Sintetiza e responde
    conclusao, _ := a.Think(ctx, "Sintetize estes dados: " + strings.Join(reflexoes, " | "))
    a.Emit("RESPOSTA_WPP", conclusao)
})
```

### 10.4 Saída Estruturada (ThinkJSON)
Para automações que exigem dados tipados, use o `ThinkJSON`.
```go
type LeadScore struct {
    Score int    `json:"score"`
    Setor string `json:"setor"`
}

var result LeadScore
err := a.ThinkJSON(ctx, "Analise este lead: "+msg, &result)
// O framework força o JSON e faz o Unmarshal automaticamente.
```

---

## 🏢 11. Padrões Avançados (Enterprise Patterns)

### Padrão de Orquestrador Reativo
Em vez de um grafo central, use eventos de "necessidade". O Orquestrador emite `NEED_RESEARCH` e qualquer agente capaz de pesquisar atende o pedido.

### Padrão Sentinela (Guardrails)
Crie agentes invisíveis que monitoram o Nexus em busca de anomalias. Ou use o middleware nativo:
```go
agente.WithMiddleware(agent.SafetyMiddleware()) // Protege contra injeção e PII
```

### 🧬 Novo: Blueprints Reativos (Deep Reasoning)
Padrões de raciocínio que operam de forma puramente assíncrona:
- **`WithReactiveReAct()`**: Loop de pensamento/ferramenta que libera a Goroutine durante a espera.
- **`WithReactiveToT(5)`**: Exploração concorrente de 5 ramos de pensamento (Tree of Thoughts).
- **`WithReactiveSupervisor()`**: Delegação de tarefas para sub-agentes via eventos.

### 🩹 Novo: Self-Healing & Budgeting
- **Self-Healing:** O agente detecta erros de ferramentas e tenta se auto-corrigir re-analisando a falha.
- **Budgeting:** Define um teto de tokens. Se exceder, o sistema emite um sinal de veto, protegendo seu faturamento.

### Resiliência Total (LLM Fallback)
Nunca fique na mão se um provedor falhar.
```go
multi := provider.NewMultiProvider(openai, gemini)
agente.WithLLM(multi) // Tenta OpenAI, se falhar pula para Gemini
```

---

## 📈 12. Performance e Escalabilidade

### Gerenciamento de Memória
Go é excelente em gerenciar memória, mas agentes com históricos gigantes podem ser pesados. Recomendamos o uso de **Memória Episódica Summarizada** para conversas longas.

### Otimização de Goroutines
O NexoRount usa `sync.Pool` e gerenciamento fino de canais para garantir que o Nexus não se torne um gargalo mesmo com milhões de eventos.

---

## 🔍 13. Solução de Problemas (Troubleshooting)

### "Meu agente não responde"
1. Verifique se o `Engine.Start()` foi chamado.
2. Verifique se o tipo de evento emitido corresponde ao tipo esperado no `.On()`.
3. Verifique o Journal para ver se o evento chegou ao Nexus.

### "Conflito de Versão na Memória"
Isso acontece quando dois processos tentam salvar a mesma chave ao mesmo tempo. A solução é implementar um **Retry Exponential Backoff** no handler do agente.

---

## 🗺️ 14. Contribuição e Roadmap

### Próximos Passos (v1.5)
- Interface Visual de Debugging (NexoDashboard).
- Integração nativa com Vector Clocks para causalidade distribuída.
- Suporte a Agentes Multimodais (Visão e Áudio como eventos).

### Como Contribuir
O NexoRount é Open Source. Sinta-se livre para abrir Issues e Pull Requests.

---

## 🏗️ 15. Manual Técnico Profundo (The Deep Manual)

Esta seção é dedicada a desenvolvedores que desejam estender as capacidades core do framework ou entender as entranhas da inteligência reativa.

### 15.1 Internals do Nexus (O Coração do Sistema)
O Nexus não é apenas um pub/sub. Ele foi desenhado para garantir que a ordem causal dos pensamentos seja respeitada, mesmo em um ambiente assíncrono. No `LocalNexus`, utilizamos canais com buffer e goroutines de despacho protegidas por `recover()`. Isso significa que se um agente falhar (sofrer um panic) durante a reação a um evento, o Nexus não morre. Ele isola a falha e continua entregando eventos para os outros agentes.

#### Algoritmo de Despacho:
```mermaid
graph TD
    A[Publicador] --> B[Nexus]
    B --> C{Tipo de Evento?}
    C --> D[Subscriber 1]
    C --> E[Subscriber 2]
    C --> F[Subscriber N]
    D --> G[Goroutine Independente]
    E --> H[Goroutine Independente]
    F --> I[Goroutine Independente]
```

Cada subscriber recebe o evento em sua própria trilha de execução. Isso é fundamental para evitar que um agente lento trave o ecossistema inteiro.

### 15.2 Gestão de Estado e Versionamento (pkg/memory)
A maior dificuldade em sistemas multi-agente é o estado compartilhado. Se o Agente A e o Agente B estão lendo e escrevendo no perfil do usuário ao mesmo tempo, quem vence?
No NexoRount, implementamos o **Optimistic Locking**. 
Quando você faz um `RecallVersioned`, o sistema retorna o dado e um número de versão (ex: 5). Se você tentar salvar o dado com a versão 5, mas outro agente já salvou e a versão atual no banco é 6, o framework retornará um erro de conflito.

**Por que isso é melhor que travas (Locks)?**
Porque travas matam a performance. Com travas otimistas, os agentes operam em velocidade máxima e só lidam com o conflito se ele realmente acontecer. É o modelo usado pelos bancos de dados mais modernos do mundo.

### 15.3 Desenvolvimento de Middlewares Customizados
Você pode estender o `Think` do agente criando seus próprios middlewares. O middleware segue o padrão `Chain of Responsibility`.

```go
func CustomLoggerMiddleware() agent.ThinkMiddleware {
    return func(a *agent.BaseAgent, next func(ctx, prompt) (string, error)) func(ctx, prompt) (string, error) {
        return func(ctx context.Context, prompt string) (string, error) {
            start := time.Now()
            res, err := next(ctx, prompt)
            fmt.Printf("Pensamento levou: %v\n", time.Since(start))
            return res, err
        }
    }
}
```

### 15.4 O Kernel Cognitivo e o Escalonamento de Pensamento
O Kernel funciona como um gerenciador de semáforos inteligentes. Cada agente, ao nascer, é registrado no Kernel. O Kernel mantém um mapa de "Slots Mentais".
- Se um agente tem capacidade 5 e tenta processar o 6º pensamento, o Kernel o suspende.
- O Kernel utiliza um `Context` para permitir o cancelamento de pensamentos que estão levando tempo demais ou que se tornaram irrelevantes (ex: o usuário cancelou a query).

### 15.5 Conectividade Híbrida (NATS + Local)
O pacote `events/bridge.go` é o que torna o NexoRount único. Ele permite criar uma topologia de "Estrela" ou "Malha".
- **Topologia em Estrela:** Um Nexus central na Cloud conectado a múltiplos Engines no Edge.
- **Topologia em Malha:** Múltiplos clusters NATS trocando eventos entre regiões geográficas.

### 15.6 Boas Práticas de Engenharia Agentica
1. **Evite Estados Gigantes:** Quanto maior a memória que você passa para a LLM, mais caro e lento é o pensamento. Use a Memória Semântica para buscar apenas o necessário.
2. **Handlers Pequenos:** Um handler de evento não deveria fazer tudo. Ele deveria processar uma parte da lógica e emitir um novo evento para que outro agente (ou o mesmo) continue o trabalho.
3. **Use IDs Únicos:** Sempre gere IDs únicos para suas entidades. O NexoRount usa UUIDs internamente para garantir que o Journal seja auditável.

### 15.7 O Futuro: A Versão 2.0
Estamos trabalhando em **Self-Evolving Systems**, onde agentes podem emitir eventos para o Kernel solicitando a criação de "filhos" (sub-agentes) dinamicamente para resolver problemas complexos, e depois os destruindo para economizar recursos.

---

## 📜 16. Glossário Técnico para Desenvolvedores

- **Nexus:** O barramento de eventos concorrente. O "éter" onde os agentes vivem.
- **Engine:** O runtime que orquestra a execução e o ciclo de vida.
- **Kernel:** O gestor de recursos cognitivos (LLMs e tokens).
- **TraceID:** O fio condutor que une múltiplos eventos em uma única jornada.
- **Span:** Uma unidade de medida de tempo para uma operação atômica.
- **Tool:** Uma função externa que o agente pode invocar para interagir com o mundo real.
- **Memory Tier:** Um nível da hierarquia de memória (Working, Short, Long, Semantic, Episodic).

---

## 🏛️ 17. Guia de Implementação de Provedores (LLM Adapters)

Se você deseja usar um modelo que o NexoRount ainda não suporta (ex: Anthropic Claude ou um modelo local via Llama.cpp), você só precisa implementar a interface `LLMProvider` no pacote `pkg/provider`.

```go
type LLMProvider interface {
    Predict(ctx context.Context, messages []Message) (string, error)
    Embed(ctx context.Context, text string) ([]float32, error)
}
```

Implementar isso permite que seu modelo se beneficie automaticamente de todos os middlewares de budget, kernel e observabilidade que já construímos.

---

## 📈 18. Benchmark e Métricas de Performance

Em nossos testes internos em um servidor Standard (4 vCPUs, 8GB RAM):
- **Event Latency:** < 500 microsegundos (LocalNexus).
- **Concurrent Agents:** 50,000 agentes ativos sem degradação de performance do Nexus.
- **Memory Overhead:** Aproximadamente 1.2MB por agente em repouso.

---

---

## 🏛️ 19. Sintaxe Moderna e DX (The Fluent API)

Para tornar o desenvolvimento "gostoso", o NexoRount agora suporta encadeamento total de capacidades. Veja um agente configurado com 100% dos recursos:

```go
sdr := agent.New("sdr-01", "SDR").
    WithClaude("claude-3-5-sonnet"). // Atalho para provedor
    WithInstructions("Venda o produto X.").
    WithPostgresMemory(ctx, dsn).    // RAG nativo
    WithReactiveReAct().             // Deep reasoning
    WithSelfHealing().               // Auto-correção
    WithMemoryConsolidation().       // Aprendizado orgânico
    WithBudget(10000, budgetManager) // Controle de custo
```

**NexoRount: Porque a inteligência não deveria ser um grafo, mas um organismo.** 🚀

*(Este documento é o manual oficial para o desenvolvimento de sistemas vivos de inteligência artificial.)*
