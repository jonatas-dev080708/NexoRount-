package loader

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// WebCrawlerLoader carrega múltiplas páginas de um site recursivamente
type WebCrawlerLoader struct {
	BaseURL     string
	MaxDepth    int
	visited     map[string]bool
	mu          sync.Mutex
}

func NewWebCrawlerLoader(url string, depth int) *WebCrawlerLoader {
	return &WebCrawlerLoader{
		BaseURL:  url,
		MaxDepth: depth,
		visited:  make(map[string]bool),
	}
}

func (l *WebCrawlerLoader) Load(ctx context.Context) ([]Document, error) {
	var docs []Document
	err := l.crawl(ctx, l.BaseURL, 0, &docs)
	return docs, err
}

func (l *WebCrawlerLoader) crawl(ctx context.Context, url string, depth int, docs *[]Document) error {
	if depth > l.MaxDepth {
		return nil
	}

	l.mu.Lock()
	if l.visited[url] {
		l.mu.Unlock()
		return nil
	}
	l.visited[url] = true
	l.mu.Unlock()

	fmt.Printf("🕷️  Crawling: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Extrai texto limpo
	content := doc.Find("body").Text()
	*docs = append(*docs, Document{
		Content:  strings.TrimSpace(content),
		Metadata: map[string]interface{}{"url": url, "depth": depth},
	})

	// Busca links internos
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, l.BaseURL) {
			l.crawl(ctx, href, depth+1, docs)
		}
	})

	return nil
}
