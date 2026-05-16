package agent

import (
	"context"
	"fmt"
	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// SupervisorAgent orquestra tarefas complexas delegando para outros agentes
type SupervisorAgent struct {
	*BaseAgent
	subAgents []string
}

func NewSupervisor(id string, subAgents []string) *SupervisorAgent {
	return &SupervisorAgent{
		BaseAgent: New(id, "SUPERVISOR"),
		subAgents: subAgents,
	}
}

// Delegate analisa uma tarefa e emite o evento para o agente mais apto
func (s *SupervisorAgent) Delegate(ctx context.Context, task string) error {
	prompt := fmt.Sprintf(`
		Você é um Supervisor. Sua tarefa é delegar o seguinte pedido para o melhor agente disponível.
		Pedido: %s
		Agentes Disponíveis: %v
		
		Responda apenas com o ID do agente escolhido.
	`, task, s.subAgents)

	agentID, err := s.Think(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Printf("👨‍✈️ [Supervisor] Delegando tarefa para: %s\n", agentID)
	
	// Emite o evento de tarefa para o agente específico
	s.Emit(events.Type("TASK_ASSIGNED_"+agentID), task)
	return nil
}
