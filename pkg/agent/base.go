package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/memory"
	"github.com/jonatas-dev080708/NexoRount-/pkg/observability"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

// HandlerFunc define a assinatura para processar eventos. Agora recebe o agente como primeiro parâmetro.
type HandlerFunc func(a *BaseAgent, ctx context.Context, e events.Event)

type State string

const (
	StateSleeping   State = "sleeping"
	StateThinking   State = "thinking"
	StateWaiting    State = "waiting"
	StateBlocked    State = "blocked"
	StateFailed     State = "failed"
	StateRecovering State = "recovering"
)

type BaseAgent struct {
	id           string
	category     string
	state        State
	handlers     map[events.Type]HandlerFunc
	mu           sync.RWMutex
	nexus        events.Bus
	llm          provider.LLMProvider
	memory       *memory.Hierarchy
	tools        *ToolManager
	instructions string
	middlewares  []ThinkMiddleware
	preFetchers  map[events.Type][]PreFetcherFunc
	journal      *observability.Journal
	// Configurações de Geração
	temperature   float32
	maxTokens     int
	topP          float32
	stopSequences []string
}

type ThinkMiddleware func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error)
type PreFetcherFunc func(a *BaseAgent, ctx context.Context, e events.Event) (string, error)

// New cria uma nova instância de BaseAgent.
// O ID deve ser único no ecossistema e a categoria ajuda na organização e filtragem.
func New(id, category string) *BaseAgent {
	return &BaseAgent{
		id:          id,
		category:    category,
		state:       StateSleeping,
		handlers:    make(map[events.Type]HandlerFunc),
		tools:       NewToolManager(),
		preFetchers: make(map[events.Type][]PreFetcherFunc),
	}
}

// Status retorna o estado atual do agente (thinking, sleeping, blocked, etc).
func (a *BaseAgent) Status() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

func (a *BaseAgent) setState(s State) {
	a.mu.Lock()
	oldState := a.state
	a.state = s
	a.mu.Unlock()

	if oldState != s {
		a.Emit("AGENT_STATE_CHANGED", map[string]string{
			"agent_id":  a.id,
			"old_state": string(oldState),
			"new_state": string(s),
		})
	}
}

func (a *BaseAgent) ID() string   { return a.id }
func (a *BaseAgent) Type() string { return a.category }

func (a *BaseAgent) On(eventType events.Type, handler HandlerFunc) *BaseAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handlers[eventType] = handler
	return a
}

// Fluent API Extensions

// WithReactiveReAct ativa o padrão Reasoning + Acting via eventos assíncronos.
// Recomendado para tarefas complexas que exigem uso de ferramentas e raciocínio multi-etapa.
func (a *BaseAgent) WithReactiveReAct() *BaseAgent {
	EnableReactiveReAct(a)
	return a
}

// WithReactiveToT ativa a exploração paralela de múltiplos caminhos de raciocínio.
// O parâmetro branches define quantos caminhos simultâneos o agente deve explorar.
func (a *BaseAgent) WithReactiveToT(branches int) *BaseAgent {
	EnableReactiveToT(a, branches)
	return a
}

// WithSelfHealing permite que o agente aprenda com erros de ferramentas e tente corrigi-los automaticamente.
func (a *BaseAgent) WithSelfHealing() *BaseAgent {
	EnableSelfHealing(a)
	return a
}

// WithMemoryConsolidation ativa um processo em background que sumariza e indexa aprendizados no Vector Store.
func (a *BaseAgent) WithMemoryConsolidation() *BaseAgent {
	EnableMemoryConsolidation(a)
	return a
}

// WithBudget define um teto de gastos de tokens para o agente, agindo como um disjuntor de segurança.
func (a *BaseAgent) WithBudget(tokens int, manager *ReactiveBudgetManager) *BaseAgent {
	manager.SetBudget(a.id, tokens)
	EnableReactiveBudgeting(a, manager)
	return a
}

// WithInstructions define o prompt de sistema (personalidade e regras) do agente.
func (a *BaseAgent) WithInstructions(instr string) *BaseAgent {
	a.instructions = instr
	return a
}

// WithLLM associa um provedor de inteligência (OpenAI, Gemini, Claude) ao agente.
func (B *BaseAgent) WithLLM(llm provider.LLMProvider) *BaseAgent {
	B.llm = llm
	return B
}

// WithTool registra uma ferramenta que o agente pode utilizar durante o raciocínio.
func (a *BaseAgent) WithTool(t *Tool) *BaseAgent {
	a.tools.Register(t)
	return a
}

// Provedores Rápidos (Sintaxe Simplificada)

func (a *BaseAgent) WithClaude(model string) *BaseAgent {
	p, _ := provider.NewAnthropicProvider(model)
	a.llm = p
	return a
}

func (a *BaseAgent) WithOpenAI(model string) *BaseAgent {
	p, _ := provider.NewOpenAIProvider(model)
	a.llm = p
	return a
}

func (a *BaseAgent) WithGemini(ctx context.Context, model string) *BaseAgent {
	p, _ := provider.NewGeminiProvider(ctx, model)
	a.llm = p
	return a
}

func (a *BaseAgent) WithGroq(model string) *BaseAgent {
	p, _ := provider.NewGroqProvider(model)
	a.llm = p
	return a
}

func (a *BaseAgent) WithMaritaca(model string) *BaseAgent {
	p, _ := provider.NewMaritacaProvider(model)
	a.llm = p
	return a
}


// Configuração de Memória

// WithPostgresMemory conecta o agente ao PostgreSQL com suporte a PGVector.
// Isso habilita memória semântica e persistência de longo prazo.
func (a *BaseAgent) WithPostgresMemory(ctx context.Context, dsn string) *BaseAgent {
	store, err := memory.NewPGVectorStore(ctx, dsn)
	if err != nil {
		fmt.Printf("❌ Erro ao configurar PGVector: %v\n", err)
		return a
	}
	a.memory = memory.NewHierarchy(nil, nil, store)
	return a
}

// Run implementa a lógica de execução concorrente robusta
func (a *BaseAgent) Run(ctx context.Context, nexus events.Bus) error {
	a.nexus = nexus
	// Tenta capturar o Journal se o Nexus for um Engine (pattern common em Go)
	// Para simplicidade, vamos assumir que o usuário pode injetar o journal ou ele vem via Nexus
	// Aqui vamos apenas preparar o agente para receber o journal no Start
	
	var wg sync.WaitGroup
	
	// Para cada tipo de evento, criamos uma goroutine dedicada de escuta
	for eventType, handler := range a.handlers {
		ch := nexus.Subscribe(eventType)
		wg.Add(1)
		
		go func(t events.Type, h HandlerFunc, c chan events.Event) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ev, ok := <-c:
					if !ok {
						return
					}
					a.safeExecute(ctx, h, ev)
				}
			}
		}(eventType, handler, ch)
	}

	fmt.Printf("🤖 Agente [%s] online e ouvindo %d eventos\n", a.id, len(a.handlers))
	
	// O Run fica bloqueado até que o contexto seja cancelado
	wg.Wait()
	return nil
}

func (a *BaseAgent) safeExecute(ctx context.Context, handler HandlerFunc, ev events.Event) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 [ERRO CRÍTICO] Pânico no Agente [%s]: %v\n", a.id, r)
			a.setState(StateFailed)
		}
	}()

	a.setState(StateThinking)
	defer a.setState(StateSleeping)

	// Inicia um Span para a execução do Handler se houver journal
	var spanID string
	if a.journal != nil {
		spanID = a.journal.StartSpan(ev.TraceID, "handler_"+string(ev.Type), a.id)
		defer a.journal.EndSpan(spanID)
	}

	// Executa Pre-fetchers em paralelo se existirem para este tipo de evento
	a.mu.RLock()
	fetchers := a.preFetchers[ev.Type]
	a.mu.RUnlock()

	if len(fetchers) > 0 {
		var wg sync.WaitGroup
		for _, f := range fetchers {
			wg.Add(1)
			go func(prefetch PreFetcherFunc) {
				defer wg.Done()
				res, err := prefetch(a, ctx, ev)
				if err == nil && res != "" {
					// Salva o resultado do prefetch no contexto ou memória temporária do evento
					// Por simplicidade, vamos anexar ao Metadata do evento para que o handler use
					if ev.Metadata == nil {
						ev.Metadata = make(map[string]interface{})
					}
					ev.Metadata["prefetch_data"] = res
				}
			}(f)
		}
		wg.Wait()
	}

	handler(a, ctx, ev)
}

// Emit facilita o envio de eventos pelo próprio agente
func (a *BaseAgent) Emit(eventType events.Type, payload interface{}) {
	if a.nexus != nil {
		// Busca se existe um TraceID no contexto (simplificado por agora)
		// Em uma versão futura, usaríamos context values para propagar TraceID
		
		a.nexus.Publish(events.Event{
			Type:    eventType,
			Source:  a.id,
			Payload: payload,
			// Aqui no futuro propagaremos o TraceID do evento atual
		})
	}
}

// WithJournal associa um diário de bordo ao agente para rastreamento
func (a *BaseAgent) WithJournal(j *observability.Journal) *BaseAgent {
	a.journal = j
	return a
}

// WithMemory associa a hierarquia de memória ao agente
func (a *BaseAgent) WithMemory(h *memory.Hierarchy) *BaseAgent {
	a.memory = h
	return a
}

// Remember salva uma informação na memória persistente do agente
func (a *BaseAgent) Remember(key string, value interface{}) error {
	return a.RememberVersioned(key, value, -1)
}

// RememberVersioned salva com trava otimista (falha se a versão mudou)
func (a *BaseAgent) RememberVersioned(key string, value interface{}, version int) error {
	if a.memory == nil || a.memory.LongTerm == nil {
		return fmt.Errorf("memória de longo prazo não configurada para o agente %s", a.id)
	}
	return a.memory.LongTerm.SaveVersioned(a.id, key, value, version)
}

// Recall recupera uma informação da memória do agente
func (a *BaseAgent) Recall(key string) (interface{}, bool) {
	val, _, ok := a.RecallVersioned(key)
	return val, ok
}

// RecallVersioned recupera o valor e sua versão atual
func (a *BaseAgent) RecallVersioned(key string) (interface{}, int, bool) {
	if a.memory == nil || a.memory.LongTerm == nil {
		return nil, 0, false
	}
	return a.memory.LongTerm.GetVersioned(a.id, key)
}

// WorkMemory salva um dado temporário na memória de trabalho (volátil)
func (a *BaseAgent) WorkMemory(key string, value interface{}) {
	if a.memory != nil {
		a.memory.Working[key] = value
	}
}

// Search busca na memória semântica (vetorial)
func (a *BaseAgent) Search(ctx context.Context, query string, limit int) ([]memory.Document, error) {
	if a.llm == nil || a.memory == nil || a.memory.Semantic == nil {
		return nil, fmt.Errorf("LLM ou Memória Semântica não configurados")
	}

	vector, err := a.llm.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	return a.memory.Semantic.Search(ctx, a.id, vector, limit)
}

// RecallEpisode busca na memória episódica (histórico de eventos)
func (a *BaseAgent) RecallEpisode(traceID string) []events.Event {
	if a.journal != nil {
		return a.journal.GetTraceHistory(traceID)
	}
	return nil
}

// ConfigLLM define parâmetros finos de geração
func (a *BaseAgent) ConfigLLM(temp float32, maxTokens int) *BaseAgent {
	a.temperature = temp
	a.maxTokens = maxTokens
	return a
}

// WithMiddleware adiciona um interceptor para chamadas de LLM
func (a *BaseAgent) WithMiddleware(m ThinkMiddleware) *BaseAgent {
	a.middlewares = append(a.middlewares, m)
	return a
}

// WithPreFetcher registra uma função para preparar dados assim que um evento chega
func (a *BaseAgent) WithPreFetcher(t events.Type, f PreFetcherFunc) *BaseAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.preFetchers[t] = append(a.preFetchers[t], f)
	return a
}

// Think executa uma chamada de inteligência síncrona respeitando a persona
func (a *BaseAgent) Think(ctx context.Context, prompt string) (string, error) {
	res, err := a.thinkRaw(ctx, prompt, "text")
	if err != nil {
		return "", err
	}
	return res.Content, err
}

// ThinkJSON executa uma chamada e mapeia o resultado para uma struct
func (a *BaseAgent) ThinkJSON(ctx context.Context, prompt string, target interface{}) error {
	res, err := a.thinkRaw(ctx, prompt, "json_object")
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(res.Content), target)
}

func (a *BaseAgent) thinkRaw(ctx context.Context, prompt string, format string) (*provider.Result, error) {
	a.setState(StateThinking)
	defer a.setState(StateSleeping)

	// Cria a função base de execução
	execute := func(c context.Context, p string) (string, error) {
		if a.llm == nil {
			return "[Simulação: LLM não configurada]", nil
		}

		// Rastreamento do Pensamento
		if a.journal != nil {
			sid := a.journal.StartSpan("internal", "llm_predict", a.id)
			defer a.journal.EndSpan(sid)
		}

		messages := []provider.Message{}
		if a.instructions != "" {
			messages = append(messages, provider.Message{Role: "system", Content: a.instructions})
		}
		messages = append(messages, provider.Message{Role: "user", Content: p})

		config := &provider.LLMConfig{
			Temperature:    a.temperature,
			MaxTokens:      a.maxTokens,
			TopP:           a.topP,
			StopSequences:  a.stopSequences,
			ResponseFormat: format,
		}

		// Emite evento de início de predição para Budgeting e Observabilidade
		a.Emit("LLM_PREDICT_START", nil)

		res, err := a.llm.Predict(c, messages, config)
		if err != nil {
			return "", err
		}

		// Emite evento de fim de predição com contagem de tokens
		a.Emit("LLM_PREDICT_END", map[string]interface{}{
			"tokens": res.Tokens,
			"model":  a.llm.Name(),
		})

		return res.Content, nil
	}

	// Aplica Middlewares (em ordem inversa para que o primeiro adicionado seja o mais externo)
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		execute = a.middlewares[i](a, execute)
	}

	resContent, err := execute(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &provider.Result{Content: resContent}, nil
}

// ThinkParallel executa múltiplos pensamentos simultaneamente usando Goroutines
func (a *BaseAgent) ThinkParallel(ctx context.Context, prompts []string) ([]string, error) {
	results := make([]string, len(prompts))
	errs := make([]error, len(prompts))
	var wg sync.WaitGroup

	for i, p := range prompts {
		wg.Add(1)
		go func(idx int, pr string) {
			defer wg.Done()
			res, err := a.Think(ctx, pr)
			results[idx] = res
			errs[idx] = err
		}(i, p)
	}

	wg.Wait()

	// Retorna o primeiro erro encontrado, se houver
	for _, err := range errs {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// StreamThink executa uma chamada de inteligência e emite tokens em tempo real no Nexus
func (a *BaseAgent) StreamThink(ctx context.Context, prompt string) (chan string, error) {
	if a.llm == nil {
		return nil, fmt.Errorf("provedor de LLM não configurado")
	}

	messages := []provider.Message{}
	if a.instructions != "" {
		messages = append(messages, provider.Message{Role: "system", Content: a.instructions})
	}
	messages = append(messages, provider.Message{Role: "user", Content: prompt})

	tokenChan, err := a.llm.Stream(ctx, messages, &provider.LLMConfig{
		Temperature: a.temperature,
		MaxTokens:   a.maxTokens,
	})
	if err != nil {
		return nil, err
	}

	out := make(chan string, 100)
	streamID := fmt.Sprintf("stream-%d", time.Now().UnixNano())

	go func() {
		defer close(out)
		for token := range tokenChan {
			// Publicamos cada token no Nexus para que o ecossistema reaja
			a.nexus.Publish(events.Event{
				Type:    "TOKEN_STREAM",
				Source:  a.id,
				Payload: token,
				Metadata: map[string]interface{}{
					"stream_id": streamID,
					"is_final":  false,
				},
			})
			out <- token
		}
		// Sinaliza fim do stream
		a.nexus.Publish(events.Event{
			Type:   "TOKEN_STREAM",
			Source: a.id,
			Metadata: map[string]interface{}{
				"stream_id": streamID,
				"is_final":  true,
			},
		})
	}()

	return out, nil
}

// Learn permite que o agente aprenda uma nova informação salvando-a no pgvector
func (a *BaseAgent) Learn(ctx context.Context, content string, metadata map[string]interface{}) error {
	if a.llm == nil || a.memory == nil {
		return fmt.Errorf("LLM ou Memória não configurados")
	}

	vector, err := a.llm.Embed(ctx, content)
	if err != nil {
		return err
	}

	// Usamos a memória semântica se disponível
	if a.memory.Semantic != nil {
		return a.memory.Semantic.SaveVector(ctx, a.id, content, vector, metadata)
	}
	return fmt.Errorf("o armazenamento atual não suporta busca vetorial")
}


// Do executa uma ferramenta pelo nome de forma simples
func (a *BaseAgent) Do(toolName string, args string) (string, error) {
	t, ok := a.tools.Get(toolName)
	if !ok {
		return "", fmt.Errorf("ferramenta %s não encontrada no agente %s", toolName, a.id)
	}

	if t.IsAsync {
		fmt.Printf("⏳ [Agente: %s] Iniciando ferramenta ASSÍNCRONA: %s\n", a.id, toolName)
		go func() {
			a.setState(StateWaiting)
			res, err := t.Execute(args)
			a.setState(StateSleeping)

			status := "success"
			if err != nil {
				status = "error"
			}

			a.Emit("TOOL_ASYNC_RESPONSE", map[string]interface{}{
				"tool":    toolName,
				"result":  res,
				"error":   err,
				"status":  status,
				"payload": args,
			})
		}()
		return "[PENDING: Ferramenta assíncrona iniciada. O resultado chegará via evento TOOL_ASYNC_RESPONSE]", nil
	}

	a.setState(StateWaiting)
	defer a.setState(StateSleeping)

	fmt.Printf("🛠️  [Agente: %s] Executando ferramenta SÍNCRONA: %s\n", a.id, toolName)
	return t.Execute(args)
}
