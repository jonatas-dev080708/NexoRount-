package agent

import (
	"context"
	"fmt"
)

// Step representa um passo em um plano
type Step struct {
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
}

// Plan representa uma sequência de passos
type Plan struct {
	Steps []Step `json:"steps"`
}

// ThinkPlan gera um plano de ação estruturado
func (a *BaseAgent) ThinkPlan(ctx context.Context, task string) (Plan, error) {
	fmt.Printf("🗺️  [Planner] Gerando plano estratégico para: %s\n", task)
	
	prompt := fmt.Sprintf(`
		Crie um plano de passos para resolver a seguinte tarefa: %s
		Ferramentas disponíveis: %s
		Retorne um JSON no formato: {"steps": [{"tool": "nome", "arguments": "input"}]}
	`, task, a.tools.ToDescription())

	var plan Plan
	err := a.ThinkJSON(ctx, prompt, &plan)
	return plan, err
}

// ExecutePlan executa sequencialmente os passos de um plano
func (a *BaseAgent) ExecutePlan(ctx context.Context, plan Plan) (string, error) {
	fmt.Printf("⚙️  [Executor] Iniciando execução de %d passos...\n", len(plan.Steps))
	
	results := ""
	for i, step := range plan.Steps {
		fmt.Printf("   ▶️ Passo %d: %s(%s)\n", i+1, step.Tool, step.Arguments)
		res, err := a.Do(step.Tool, step.Arguments)
		if err != nil {
			return "", fmt.Errorf("falha no passo %d: %v", i+1, err)
		}
		results += fmt.Sprintf("Resultado Passo %d: %s\n", i+1, res)
	}

	return results, nil
}
