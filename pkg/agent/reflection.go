package agent

import (
	"context"
	"fmt"
)

// ThinkWithReflection gera uma resposta e depois a revisa e refina
func (a *BaseAgent) ThinkWithReflection(ctx context.Context, prompt string) (string, error) {
	// 1. Geração Inicial
	fmt.Printf("🖋️  [Reflection] Gerando resposta inicial...\n")
	initial, err := a.Think(ctx, prompt)
	if err != nil {
		return "", err
	}

	// 2. Reflexão / Crítica
	fmt.Printf("🔍 [Reflection] Revisando e criticando a resposta...\n")
	critiquePrompt := fmt.Sprintf(`
		Analise criticamente a seguinte resposta para o prompt: "%s"
		Resposta a ser analisada: "%s"
		
		Identifique falhas, imprecisões ou pontos de melhoria.
		Se a resposta estiver perfeita, diga apenas "PERFEITO".
	`, prompt, initial)
	
	critique, err := a.Think(ctx, critiquePrompt)
	if err != nil {
		return initial, nil // Retorna a inicial em caso de erro na crítica
	}

	if critique == "PERFEITO" {
		return initial, nil
	}

	// 3. Refinamento Final
	fmt.Printf("✨ [Reflection] Refinando resposta com base na crítica...\n")
	refinePrompt := fmt.Sprintf(`
		Melhore a resposta inicial com base na crítica fornecida.
		Resposta Inicial: %s
		Crítica: %s
		
		Forneça apenas a versão final aprimorada.
	`, initial, critique)

	return a.Think(ctx, refinePrompt)
}
