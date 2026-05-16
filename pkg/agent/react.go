package agent

import (
	"context"
	"fmt"
	"strings"
)

// ThinkReAct implementa o padrão Reasoning + Acting (ReAct)
func (a *BaseAgent) ThinkReAct(ctx context.Context, prompt string, maxSteps int) (string, error) {
	currentPrompt := prompt
	
	// Adiciona instruções de ReAct se não estiverem presentes
	reactInstructions := `
	Use o seguinte formato para resolver o problema:
	Thought: seu raciocínio sobre o que fazer agora.
	Action: o nome da ferramenta a usar (escolha entre: %s).
	Action Input: o input para a ferramenta.
	Observation: o resultado da ferramenta (será fornecido a você).
	... (repetir Thought/Action/Observation se necessário)
	Final Answer: a resposta final para o usuário.
	`
	toolsDesc := a.tools.ToDescription()
	systemPrompt := fmt.Sprintf(reactInstructions, toolsDesc)
	
	// Histórico interno do loop ReAct
	history := systemPrompt + "\nQuestion: " + prompt + "\n"

	for i := 0; i < maxSteps; i++ {
		fmt.Printf("🔄 [ReAct] Passo %d/%d para o agente %s\n", i+1, maxSteps, a.id)
		
		res, err := a.Think(ctx, history)
		if err != nil {
			return "", err
		}

		history += res + "\n"

		// Verifica se temos a resposta final
		if strings.Contains(res, "Final Answer:") {
			parts := strings.Split(res, "Final Answer:")
			return strings.TrimSpace(parts[len(parts)-1]), nil
		}

		// Extrai Action e Action Input
		action := a.extractField(res, "Action:")
		actionInput := a.extractField(res, "Action Input:")

		if action != "" {
			fmt.Printf("🛠️  [ReAct] Executando ferramenta: %s com input: %s\n", action, actionInput)
			obs, err := a.Do(action, actionInput)
			if err != nil {
				obs = fmt.Sprintf("Erro ao executar ferramenta: %v", err)
			}
			
			history += "Observation: " + obs + "\n"
		} else {
			// Se não encontrou ação nem resposta final, tenta forçar encerramento
			history += "Thought: Não identifiquei uma ação clara. Vou tentar concluir.\n"
		}
	}

	return "", fmt.Errorf("limite de passos ReAct atingido sem resposta final")
}

func (a *BaseAgent) extractField(text, field string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), field) {
			return strings.TrimSpace(strings.Replace(line, field, "", 1))
		}
	}
	return ""
}
