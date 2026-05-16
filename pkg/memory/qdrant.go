package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// QdrantStore implementa a memória semântica usando o banco vetorial Qdrant
type QdrantStore struct {
	BaseURL    string
	Collection string
}

func NewQdrantStore(url, collection string) *QdrantStore {
	return &QdrantStore{BaseURL: url, Collection: collection}
}

func (q *QdrantStore) SaveVector(ctx context.Context, agentID, content string, vector []float32, metadata map[string]interface{}) error {
	// Padrão do Qdrant: PUT /collections/{name}/points
	point := map[string]interface{}{
		"points": []map[string]interface{}{
			{
				"id":     fmt.Sprintf("%x", content), // ID simples baseado no hash
				"vector": vector,
				"payload": map[string]interface{}{
					"agent_id": agentID,
					"content":  content,
					"metadata": metadata,
				},
			},
		},
	}

	body, _ := json.Marshal(point)
	url := fmt.Sprintf("%s/collections/%s/points", q.BaseURL, q.Collection)
	req, _ := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdrant error: %d", resp.StatusCode)
	}

	return nil
}

func (q *QdrantStore) Search(ctx context.Context, agentID string, vector []float32, limit int) ([]Document, error) {
	search := map[string]interface{}{
		"vector":      vector,
		"limit":       limit,
		"with_payload": true,
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "agent_id", "match": map[string]interface{}{"value": agentID}},
			},
		},
	}

	body, _ := json.Marshal(search)
	url := fmt.Sprintf("%s/collections/%s/points/search", q.BaseURL, q.Collection)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			Payload struct {
				Content  string                 `json:"content"`
				Metadata map[string]interface{} `json:"metadata"`
			} `json:"payload"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	var docs []Document
	for _, r := range result.Result {
		docs = append(docs, Document{
			Content:  r.Payload.Content,
			Metadata: r.Payload.Metadata,
		})
	}

	return docs, nil
}
