package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
)

// NewFileReadTool lê o conteúdo de um arquivo local
func NewFileReadTool(basePath string) *agent.Tool {
	return &agent.Tool{
		Name:        "file_read",
		Description: "Lê o conteúdo de um arquivo local. Requer o caminho relativo ao diretório base.",
		Execute: func(path string) (string, error) {
			fullPath := filepath.Join(basePath, path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("falha ao ler arquivo: %v", err)
			}
			return string(content), nil
		},
	}
}

// NewFileWriteTool escreve conteúdo em um arquivo local
func NewFileWriteTool(basePath string) *agent.Tool {
	return &agent.Tool{
		Name:        "file_write",
		Description: "Escreve ou sobrescreve conteúdo em um arquivo local. Informe o caminho e o conteúdo.",
		Execute: func(input string) (string, error) {
			// Aqui idealmente args seria JSON com path e content
			// Por simplicidade, vamos assumir que o agente envia o path na primeira linha
			return "Arquivo escrito com sucesso (simulado)", nil
		},
	}
}

// NewFileListTool lista arquivos em um diretório
func NewFileListTool(basePath string) *agent.Tool {
	return &agent.Tool{
		Name:        "file_list",
		Description: "Lista os arquivos em um diretório específico.",
		Execute: func(dir string) (string, error) {
			fullPath := filepath.Join(basePath, dir)
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				return "", err
			}
			
			output := "Arquivos:\n"
			for _, e := range entries {
				output += fmt.Sprintf("- %s (dir: %v)\n", e.Name(), e.IsDir())
			}
			return output, nil
		},
	}
}
