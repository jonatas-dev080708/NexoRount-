package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// Tipos de Eventos Internos para os Blueprints Reativos
const (
	EventThinkStep      events.Type = "AGENT_THINK_STEP"
	EventToolRequest    events.Type = "AGENT_TOOL_REQUEST"
	EventToolResponse   events.Type = "AGENT_TOOL_RESPONSE"
	EventFinalAnswer    events.Type = "AGENT_FINAL_ANSWER"
	EventBranchStart    events.Type = "AGENT_BRANCH_START"
	EventBranchComplete events.Type = "AGENT_BRANCH_COMPLETE"
)

// EnableReactiveReAct transforma o agente em um organismo de raciocínio reativo.
// Em vez de um loop síncrono, ele usa o Nexus para gerenciar cada passo do raciocínio.
func EnableReactiveReAct(a *BaseAgent) {
	
	// Handler: Processa um passo de pensamento
	a.On(EventThinkStep, func(a *BaseAgent, ctx context.Context, e events.Event) {
		data, ok := e.Payload.(map[string]interface{})
		if !ok { return }

		query := data["query"].(string)
		history := data["history"].(string)
		step := data["step"].(int)

		if step > 10 {
			a.Emit(EventFinalAnswer, "Erro: Limite de passos atingido sem conclusão.")
			return
		}

		fmt.Printf("🧠 [%s] Reativo - Passo %d\n", a.id, step)

		// O agente pensa no próximo passo
		prompt := fmt.Sprintf(`
			Questão: %s
			Histórico de ações: %s
			
			Analise e decida o próximo passo.
		`, query, history)

		res, err := a.Think(ctx, prompt)
		if err != nil {
			a.Emit(EventFinalAnswer, "Erro no pensamento: "+err.Error())
			return
		}

		// Se encontrou a resposta final
		if strings.Contains(res, "Final Answer:") {
			a.Emit(EventFinalAnswer, res)
			return
		}

		// Se precisa de uma ferramenta
		action := a.extractField(res, "Action:")
		actionInput := a.extractField(res, "Action Input:")

		if action != "" {
			// Emitimos o pedido da ferramenta e liberamos a goroutine
			a.Emit(EventToolRequest, map[string]interface{}{
				"agent_id": a.id,
				"tool":     action,
				"input":    actionInput,
				"context":  data, // Passamos o estado completo para o próximo ciclo
			})
		} else {
			// Caso não entenda o que fazer, tenta forçar uma conclusão
			data["history"] = history + "\nSistema: Não entendi sua ação. Por favor, seja claro."
			data["step"] = step + 1
			a.Emit(EventThinkStep, data)
		}
	})

	// Handler: Reage ao resultado de uma ferramenta
	a.On(EventToolResponse, func(a *BaseAgent, ctx context.Context, e events.Event) {
		data, ok := e.Payload.(map[string]interface{})
		if !ok { return }

		result := data["result"].(string)
		ctxData := data["context"].(map[string]interface{})

		// Evolui o histórico e volta para o passo de pensamento
		ctxData["history"] = ctxData["history"].(string) + "\nObservation: " + result
		ctxData["step"] = ctxData["step"].(int) + 1

		a.Emit(EventThinkStep, ctxData)
	})
}

// EnableReactiveToT implementa o Tree of Thoughts de forma puramente reativa e paralela
func EnableReactiveToT(a *BaseAgent, branches int) {
	
	// Handler: Inicia a exploração da árvore
	a.On("TOT_START", func(a *BaseAgent, ctx context.Context, e events.Event) {
		question := e.Payload.(string)
		fmt.Printf("🌲 [%s] ToT Reativo - Iniciando %d ramos\n", a.id, branches)

		for i := 0; i < branches; i++ {
			// Disparamos eventos para cada ramo. O Nexus pode distribuir isso entre workers.
			a.Emit(EventBranchStart, map[string]interface{}{
				"branch_id": i + 1,
				"question":  question,
			})
		}
	})

	// Handler: Processa um único ramo da árvore
	a.On(EventBranchStart, func(a *BaseAgent, ctx context.Context, e events.Event) {
		data := e.Payload.(map[string]interface{})
		branchID := data["branch_id"].(int)
		question := data["question"].(string)

		fmt.Printf("   🌱 Explorando Ramo %d...\n", branchID)
		res, _ := a.Think(ctx, "Crie um raciocínio para: "+question)

		// Emite o resultado do ramo para ser coletado
		a.Emit(EventBranchComplete, map[string]interface{}{
			"branch_id": branchID,
			"result":    res,
			"question":  question,
		})
	})
	
	// Nota: Um Agente Supervisor ou um Collector Handler 
	// seria responsável por juntar os EventBranchComplete e emitir a decisão final.
}

// EnableReactiveSupervisor transforma o agente em um mestre de orquestração reativo.
// Ele delega tarefas via eventos e aguarda os resultados de forma assíncrona.
func EnableReactiveSupervisor(s *BaseAgent, team []string) {
	
	// Handler: Recebe uma tarefa complexa e decide quem deve executá-la
	s.On("SUPERVISOR_DELEGATE", func(a *BaseAgent, ctx context.Context, e events.Event) {
		task := e.Payload.(string)
		fmt.Printf("👨‍✈️ [%s] Supervisor Reativo - Analisando delegação para: %s\n", s.id, task)

		prompt := fmt.Sprintf(`
			Você é um Supervisor. Sua tarefa é delegar o seguinte pedido para o melhor agente do seu time.
			Time: %v
			Pedido: %s
			
			Responda apenas com o ID do agente escolhido.
		`, team, task)

		agentID, err := s.Think(ctx, prompt)
		if err != nil {
			s.Emit("SUPERVISOR_ERROR", err.Error())
			return
		}

		fmt.Printf("   👉 Delegando para: %s\n", agentID)

		// Emitimos a delegação. O agente alvo deve estar ouvindo TASK_ASSIGNED_{ID}
		s.Emit(events.Type("TASK_ASSIGNED_"+agentID), map[string]interface{}{
			"supervisor_id": s.id,
			"task":          task,
			"trace_id":      e.TraceID, // Mantemos o rastro
		})
	})

	// Handler: O supervisor ouve quando qualquer membro do time termina uma tarefa
	s.On("TASK_COMPLETED", func(a *BaseAgent, ctx context.Context, e events.Event) {
		data, ok := e.Payload.(map[string]interface{})
		if !ok { return }

		// Verifica se esta tarefa foi delegada por este supervisor
		if data["supervisor_id"] != s.id { return }

		workerID := e.Source
		result := data["result"].(string)

		fmt.Printf("✅ [%s] Supervisor recebeu conclusão de %s\n", s.id, workerID)

		// O supervisor agora pode sintetizar a resposta final ou decidir o próximo passo
		prompt := fmt.Sprintf(`
			O agente %s completou sua tarefa com o seguinte resultado: %s
			Sintetize a resposta final para o usuário ou decida se precisamos de mais alguma ação.
		`, workerID, result)

		final, _ := s.Think(ctx, prompt)
		s.Emit("SUPERVISOR_FINAL_ANSWER", final)
	})
}
