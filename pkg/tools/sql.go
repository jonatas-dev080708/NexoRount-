package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jonatas-dev080708/NexoRount-/pkg/agent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SQLConfig define as configurações para a ferramenta SQL
type SQLConfig struct {
	Driver string
	DSN    string
}

// NewSQLQueryTool cria uma ferramenta para executar consultas SELECT
func NewSQLQueryTool(config SQLConfig) *agent.Tool {
	return &agent.Tool{
		Name:        "sql_query",
		Description: "Executa uma consulta SQL (apenas SELECT) no banco de dados e retorna os resultados em formato JSON. Use para buscar dados estruturados.",
		Execute: func(query string) (string, error) {
			db, err := sql.Open(config.Driver, config.DSN)
			if err != nil {
				return "", fmt.Errorf("falha ao conectar no banco: %v", err)
			}
			defer db.Close()

			rows, err := db.Query(query)
			if err != nil {
				return "", fmt.Errorf("erro na execução da query: %v", err)
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				return "", err
			}

			var results []map[string]interface{}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					return "", err
				}

				rowMap := make(map[string]interface{})
				for i, col := range columns {
					rowMap[col] = values[i]
				}
				results = append(results, rowMap)
			}

			jsonData, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return "", err
			}

			return string(jsonData), nil
		},
	}
}
