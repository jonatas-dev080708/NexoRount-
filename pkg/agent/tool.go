package agent

import (
	"fmt"
)

// Tool representa uma função que o agente pode executar
type Tool struct {
	Name        string
	Description string
	// Params é um exemplo do JSON schema esperado (pode ser evoluído)
	Params      interface{}
	IsAsync     bool
	Execute     func(args string) (string, error)
}

// ToolManager gerencia as ferramentas de um agente
type ToolManager struct {
	tools map[string]*Tool
}

func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]*Tool),
	}
}

func (tm *ToolManager) Register(t *Tool) {
	tm.tools[t.Name] = t
}

func (tm *ToolManager) Get(name string) (*Tool, bool) {
	t, ok := tm.tools[name]
	return t, ok
}

// ToOpenAITools converte nossas ferramentas para o formato que a OpenAI/Gemini esperam
func (tm *ToolManager) ToDescription() string {
	desc := "Ferramentas disponíveis:\n"
	for _, t := range tm.tools {
		desc += fmt.Sprintf("- %s: %s\n", t.Name, t.Description)
	}
	return desc
}
