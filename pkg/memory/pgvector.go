package memory

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
)

type Document struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
}

type PGVectorStore struct {
	conn *pgx.Conn
}

func NewPGVectorStore(ctx context.Context, connectionString string) (*PGVectorStore, error) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	// Habilita a extensão pgvector e cria a tabela base se não existir
	_, err = conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS agent_memory (
		id serial PRIMARY KEY,
		agent_id TEXT,
		content TEXT,
		metadata JSONB,
		embedding vector(1536) -- Dimensão padrão para OpenAI/Gemini
	);`
	
	_, err = conn.Exec(ctx, schema)
	return &PGVectorStore{conn: conn}, err
}

func (s *PGVectorStore) Save(id string, key string, value interface{}) error {
	// Implementação simples de KV no Postgres
	_, err := s.conn.Exec(context.Background(), 
		`INSERT INTO agent_memory (agent_id, content, metadata) 
		 VALUES ($1, $2, $3) 
		 ON CONFLICT DO NOTHING`, // Simplificado para o exemplo
		id, key, map[string]interface{}{"value": value})
	return err
}

func (s *PGVectorStore) Get(id string, key string) (interface{}, bool) {
	var val interface{}
	err := s.conn.QueryRow(context.Background(),
		"SELECT metadata->>'value' FROM agent_memory WHERE agent_id = $1 AND content = $2 LIMIT 1",
		id, key).Scan(&val)
	
	if err != nil {
		return nil, false
	}
	return val, true
}

func (s *PGVectorStore) GetAll(id string) (map[string]interface{}, error) {
	return nil, nil // Opcional para este momento
}

func (s *PGVectorStore) SaveVector(ctx context.Context, agentID string, content string, embedding []float32, metadata map[string]interface{}) error {
	_, err := s.conn.Exec(ctx, 
		"INSERT INTO agent_memory (agent_id, content, embedding, metadata) VALUES ($1, $2, $3, $4)",
		agentID, content, pgvector.NewVector(embedding), metadata)
	return err
}

func (s *PGVectorStore) Search(ctx context.Context, agentID string, queryEmbedding []float32, limit int) ([]Document, error) {
	rows, err := s.conn.Query(ctx,
		`SELECT content, metadata FROM agent_memory 
		 WHERE agent_id = $1 
		 ORDER BY embedding <=> $2 
		 LIMIT $3`,
		agentID, pgvector.NewVector(queryEmbedding), limit)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.Content, &doc.Metadata); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
