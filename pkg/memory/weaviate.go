package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WeaviateStore implementa a memória semântica usando Weaviate
type WeaviateStore struct {
	BaseURL string
	Class   string
	APIKey  string
}

func NewWeaviateStore(url, class, apiKey string) *WeaviateStore {
	return &WeaviateStore{BaseURL: url, Class: class, APIKey: apiKey}
}

func (w *WeaviateStore) SaveVector(ctx context.Context, agentID, content string, vector []float32, metadata map[string]interface{}) error {
	obj := map[string]interface{}{
		"class": w.Class,
		"properties": map[string]interface{}{
			"agent_id": agentID,
			"content":  content,
			"metadata": metadata,
		},
		"vector": vector,
	}

	body, _ := json.Marshal(obj)
	url := fmt.Sprintf("%s/v1/objects", w.BaseURL)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if w.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("weaviate error: %d", resp.StatusCode)
	}

	return nil
}

func (w *WeaviateStore) Search(ctx context.Context, agentID string, vector []float32, limit int) ([]Document, error) {
	// Weaviate usa GraphQL para buscas complexas
	query := fmt.Sprintf(`{
		Get {
			%s (
				nearVector: {vector: %v},
				limit: %d,
				where: {
					path: ["agent_id"],
					operator: Equal,
					valueString: "%s"
				}
			) {
				content
				agent_id
			}
		}
	}`, w.Class, vector, limit, agentID)

	qObj := map[string]string{"query": query}
	body, _ := json.Marshal(qObj)
	
	url := fmt.Sprintf("%s/v1/graphql", w.BaseURL)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if w.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.APIKey)
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Get map[string][]struct {
				Content string `json:"content"`
			} `json:"Get"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	var docs []Document
	if items, ok := result.Data.Get[w.Class]; ok {
		for _, item := range items {
			docs = append(docs, Document{Content: item.Content})
		}
	}

	return docs, nil
}
