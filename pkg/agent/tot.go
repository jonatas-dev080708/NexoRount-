package agent

import (
	"context"
	"fmt"
	"strings"
)

// ThinkToT implementa o padrão Tree of Thoughts (ToT)
// Ele explora múltiplos caminhos de raciocínio em paralelo e escolhe o melhor.
func (a *BaseAgent) ThinkToT(ctx context.Context, question string, branches int) (string, error) {
	fmt.Printf("🌲 [ToT] Iniciando Tree of Thoughts para: %s\n", question)

	// PASSO 1: Geração de Candidatos (Pensamentos)
	fmt.Printf("   🌱 Gerando %d caminhos de raciocínio...\n", branches)
	prompts := make([]string, branches)
	for i := 0; i < branches; i++ {
		prompts[i] = fmt.Sprintf("Apresente um caminho de raciocínio lógico (Opção %d) para resolver: %s", i+1, question)
	}

	caminhos, err := a.ThinkParallel(ctx, prompts)
	if err != nil {
		return "", err
	}

	// PASSO 2: Avaliação (Crítica)
	// O agente atua como um juiz de seus próprios pensamentos
	fmt.Println("   ⚖️  Avaliando e pontuando caminhos...")
	evalPrompts := make([]string, branches)
	for i, c := range caminhos {
		evalPrompts[i] = fmt.Sprintf(`
			Avalie o seguinte raciocínio para a questão: "%s"
			Raciocínio: %s
			
			Atribua uma nota de 0 a 10 e explique brevemente por que.
			Formato: NOTA: [valor]
		`, question, c)
	}

	avaliacoes, _ := a.ThinkParallel(ctx, evalPrompts)

	// PASSO 3: Seleção do Melhor Caminho
	bestIdx := 0
	maxScore := -1.0
	for i, eval := range avaliacoes {
		score := a.parseScore(eval)
		if score > maxScore {
			maxScore = score
			bestIdx = i
		}
	}

	fmt.Printf("   🏆 Melhor caminho escolhido: Opção %d (Nota: %.1f)\n", bestIdx+1, maxScore)

	// PASSO 4: Consolidação Final
	fmt.Println("   ✨ Consolidando resposta final baseada no melhor caminho...")
	finalPrompt := fmt.Sprintf(`
		Com base no melhor raciocínio identificado: %s
		Gere a resposta final para a questão: %s
	`, caminhos[bestIdx], question)

	return a.Think(ctx, finalPrompt)
}

func (a *BaseAgent) parseScore(text string) float64 {
	// Lógica simples de parse de nota
	if strings.Contains(text, "NOTA:") {
		var score float64
		fmt.Sscanf(text[strings.Index(text, "NOTA:"):], "NOTA: %f", &score)
		return score
	}
	return 0
}
