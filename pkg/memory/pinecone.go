package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PineconeStore implementa a memória semântica usando Pinecone
type PineconeStore struct {
	APIKey string
	Host   string
}

func NewPineconeStore(apiKey, host string) *PineconeStore {
	return &PineconeStore{APIKey: apiKey, Host: host}
}

func (p *PineconeStore) SaveVector(ctx context.Context, agentID, content string, vector []float32, metadata map[string]interface{}) error {
	// Pinecone exige que o ID seja uma string única
	vectorData := map[string]interface{}{
		"vectors": []map[string]interface{}{
			{
				"id":     fmt.Sprintf("vec_%x", content),
				"values": vector,
				"metadata": map[string]interface{}{
					"agent_id": agentID,
					"content":  content,
					"extra":    metadata,
				},
			},
		},
	}

	body, _ := json.Marshal(vectorData)
	url := fmt.Sprintf("%s/vectors/upsert", p.Host)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Api-Key", p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pinecone error: %d", resp.StatusCode)
	}

	return nil
}

func (p *PineconeStore) Search(ctx context.Context, agentID string, vector []float32, limit int) ([]Document, error) {
	query := map[string]interface{}{
		"vector":          vector,
		"topK":            limit,
		"includeMetadata": true,
		"filter": map[string]interface{}{
			"agent_id": map[string]interface{}{"$eq": agentID},
		},
	}

	body, _ := json.Marshal(query)
	url := fmt.Sprintf("%s/query", p.Host)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Api-Key", p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Matches []struct {
			Metadata struct {
				Content string `json:"content"`
			} `json:"metadata"`
		} `json:"matches"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	var docs []Document
	for _, m := range result.Matches {
		docs = append(docs, Document{
			Content: m.Metadata.Content,
		})
	}

	return docs, nil
}
