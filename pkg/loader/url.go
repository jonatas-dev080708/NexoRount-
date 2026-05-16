package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// URLLoader carrega o conteúdo de uma página web
type URLLoader struct {
	URL string
}

func NewURLLoader(url string) *URLLoader {
	return &URLLoader{URL: url}
}

func (l *URLLoader) Load(ctx context.Context) ([]Document, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", l.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status de erro ao carregar URL: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc := Document{
		Content: string(body),
		Metadata: map[string]interface{}{
			"source": l.URL,
			"type":   "web_page",
		},
	}

	return []Document{doc}, nil
}
