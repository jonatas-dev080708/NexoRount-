package loader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ledongthuc/pdf"
)

// PDFLoader carrega e extrai texto de arquivos PDF
type PDFLoader struct {
	FilePath string
}

func NewPDFLoader(path string) *PDFLoader {
	return &PDFLoader{FilePath: path}
}

func (l *PDFLoader) Load(ctx context.Context) ([]Document, error) {
	f, r, err := pdf.Open(l.FilePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir PDF: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("falha ao extrair texto do PDF: %v", err)
	}

	_, err = io.Copy(&buf, b)
	if err != nil {
		return nil, err
	}

	doc := Document{
		Content: buf.String(),
		Metadata: map[string]interface{}{
			"source": l.FilePath,
			"type":   "pdf",
		},
	}

	return []Document{doc}, nil
}

// ReadPDF simple helper for quick reads
func ReadPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
