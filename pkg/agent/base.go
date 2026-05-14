package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
	"github.com/jonatas-dev080708/NexoRount-/pkg/memory"
	"github.com/jonatas-dev080708/NexoRount-/pkg/provider"
)

// HandlerFunc define a assinatura para processar eventos. Agora recebe o agente como primeiro parâmetro.
type HandlerFunc func(a *BaseAgent, ctx context.Context, e events.Event)

type BaseAgent struct {
	id       string
	category string
	handlers map[events.Type]HandlerFunc
	mu       sync.RWMutex
	nexus    events.Bus
	llm      provider.LLMProvider
	memory   memory.Store
	tools    *ToolManager
	instructions string
	middlewares  []ThinkMiddleware
	preFetchers  map[events.Type][]PreFetcherFunc
}

type ThinkMiddleware func(a *BaseAgent, next func(context.Context, string) (string, error)) func(context.Context, string) (string, error)
type PreFetcherFunc func(a *BaseAgent, ctx context.Context, e events.Event) (string, error)

func New(id, category string) *BaseAgent {
	return &BaseAgent{
		id:       id,
		category: category,
		handlers: make(map[events.Type]HandlerFunc),
		tools:    NewToolManager(),
		preFetchers: make(map[events.Type][]PreFetcherFunc),
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

// Run implementa a lógica de execução concorrente robusta
func (a *BaseAgent) Run(ctx context.Context, nexus events.Bus) error {
	a.nexus = nexus
	
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
		}
	}()

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
		a.nexus.Publish(events.Event{
			Type:    eventType,
			Source:  a.id,
			Payload: payload,
		})
	}
}

// WithMemory associa um armazenamento de memória ao agente
func (a *BaseAgent) WithMemory(s memory.Store) *BaseAgent {
	a.memory = s
	return a
}

// Remember salva uma informação na memória persistente do agente
func (a *BaseAgent) Remember(key string, value interface{}) error {
	if a.memory == nil {
		return fmt.Errorf("memória não configurada para o agente %s", a.id)
	}
	return a.memory.Save(a.id, key, value)
}

// Recall recupera uma informação da memória do agente
func (a *BaseAgent) Recall(key string) (interface{}, bool) {
	if a.memory == nil {
		return nil, false
	}
	return a.memory.Get(a.id, key)
}

// WithLLM associa um provedor de inteligência ao agente
func (a *BaseAgent) WithLLM(p provider.LLMProvider) *BaseAgent {
	a.llm = p
	return a
}

// WithInstructions define a persona e o comportamento base do agente (System Prompt)
func (a *BaseAgent) WithInstructions(instr string) *BaseAgent {
	a.instructions = instr
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
	// Cria a função base de execução
	execute := func(c context.Context, p string) (string, error) {
		if a.llm == nil {
			return "[Simulação: LLM não configurada]", nil
		}

		messages := []provider.Message{}
		if a.instructions != "" {
			messages = append(messages, provider.Message{Role: "system", Content: a.instructions})
		}
		messages = append(messages, provider.Message{Role: "user", Content: p})

		res, err := a.llm.Predict(c, messages)
		if err != nil {
			return "", err
		}
		return res.Content, nil
	}

	// Aplica Middlewares (em ordem inversa para que o primeiro adicionado seja o mais externo)
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		execute = a.middlewares[i](a, execute)
	}

	return execute(ctx, prompt)
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

	tokenChan, err := a.llm.Stream(ctx, messages)
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

	// Tentamos fazer o cast para PGVectorStore se disponível
	if vStore, ok := a.memory.(*memory.PGVectorStore); ok {
		return vStore.SaveVector(ctx, a.id, content, vector, metadata)
	}
	return fmt.Errorf("o armazenamento atual não suporta busca vetorial")
}

// Search busca informações semanticamente similares na memória do agente
func (a *BaseAgent) Search(ctx context.Context, query string, limit int) ([]memory.Document, error) {
	if a.llm == nil || a.memory == nil {
		return nil, fmt.Errorf("LLM ou Memória não configurados")
	}

	vector, err := a.llm.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	if vStore, ok := a.memory.(*memory.PGVectorStore); ok {
		return vStore.Search(ctx, a.id, vector, limit)
	}
	return nil, fmt.Errorf("o armazenamento atual não suporta busca vetorial")
}

// WithTool registra uma ferramenta que o agente pode usar
func (a *BaseAgent) WithTool(t *Tool) *BaseAgent {
	a.tools.Register(t)
	return a
}

// Do executa uma ferramenta pelo nome de forma simples
func (a *BaseAgent) Do(toolName string, args string) (string, error) {
	t, ok := a.tools.Get(toolName)
	if !ok {
		return "", fmt.Errorf("ferramenta %s não encontrada no agente %s", toolName, a.id)
	}

	fmt.Printf("🛠️  [Agente: %s] Executando ferramenta: %s\n", a.id, toolName)
	return t.Execute(args)
}
