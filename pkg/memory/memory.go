package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Store define o contrato para persistência de estado dos agentes
type Store interface {
	Save(id string, key string, value interface{}) error
	Get(id string, key string) (interface{}, bool)
	GetAll(id string) (map[string]interface{}, error)
}

// FileStore é uma implementação simples que salva o estado em arquivos JSON
type FileStore struct {
	mu       sync.RWMutex
	basePath string
}

func NewFileStore(path string) (*FileStore, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	return &FileStore{basePath: path}, nil
}

func (s *FileStore) filePath(id string) string {
	return fmt.Sprintf("%s/%s.json", s.basePath, id)
}

func (s *FileStore) Save(id string, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load(id)
	if err != nil {
		data = make(map[string]interface{})
	}

	data[key] = value
	return s.persist(id, data)
}

func (s *FileStore) Get(id string, key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.load(id)
	if err != nil {
		return nil, false
	}

	val, ok := data[key]
	return val, ok
}

func (s *FileStore) GetAll(id string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.load(id)
}

func (s *FileStore) load(id string) (map[string]interface{}, error) {
	path := s.filePath(id)
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(file, &data)
	return data, err
}

func (s *FileStore) persist(id string, data map[string]interface{}) error {
	path := s.filePath(id)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}
