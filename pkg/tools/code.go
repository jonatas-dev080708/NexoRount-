package tools

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
)

// NewPythonRunnerTool executa um script Python localmente
// ATENÇÃO: Esta ferramenta é insegura para uso em produção com inputs não confiáveis
func NewPythonRunnerTool() *agent.Tool {
	return &agent.Tool{
		Name:        "python_run",
		Description: "Executa um script Python e retorna o output do terminal. Use para cálculos complexos ou processamento de dados.",
		Execute: func(code string) (string, error) {
			// No Windows, costuma ser 'python', no Linux/Mac 'python3'
			cmdName := "python3"
			if runtime.GOOS == "windows" {
				cmdName = "python"
			}

			cmd := exec.Command(cmdName, "-c", code)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return string(output), fmt.Errorf("erro na execução python: %v", err)
			}

			return string(output), nil
		},
	}
}
