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
	SaveVersioned(id string, key string, value interface{}, version int) error
	Get(id string, key string) (interface{}, bool)
	GetVersioned(id string, key string) (interface{}, int, bool)
	GetAll(id string) (map[string]interface{}, error)
}

// Hierarchy organiza os diferentes níveis de memória de um agente
type Hierarchy struct {
	Working   map[string]interface{} // Memória de trabalho (volátil, por execução)
	ShortTerm Store                  // Memória de curto prazo (sessão, KV)
	LongTerm  Store                  // Memória de longo prazo (persistente, DB)
	Semantic  *PGVectorStore         // Memória semântica (conhecimento, vetores)
}

func NewHierarchy(short, long Store, semantic *PGVectorStore) *Hierarchy {
	return &Hierarchy{
		Working:   make(map[string]interface{}),
		ShortTerm: short,
		LongTerm:  long,
		Semantic:  semantic,
	}
}

type versionedValue struct {
	Value   interface{} `json:"value"`
	Version int         `json:"version"`
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
	return s.SaveVersioned(id, key, value, -1) // -1 ignora a versão (força escrita)
}

func (s *FileStore) SaveVersioned(id string, key string, value interface{}, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load(id)
	if err != nil {
		data = make(map[string]interface{})
	}

	// Verifica versão se não for -1
	current, ok := data[key]
	currentVersion := 0
	if ok {
		if vv, okvv := current.(map[string]interface{}); okvv {
			if v, okv := vv["version"].(float64); okv {
				currentVersion = int(v)
			}
		}
	}

	if version != -1 && currentVersion != version {
		return fmt.Errorf("conflito de versão para chave %s: esperado %d, atual %d", key, version, currentVersion)
	}

	data[key] = versionedValue{
		Value:   value,
		Version: currentVersion + 1,
	}
	return s.persist(id, data)
}

func (s *FileStore) Get(id string, key string) (interface{}, bool) {
	val, _, ok := s.GetVersioned(id, key)
	return val, ok
}

func (s *FileStore) GetVersioned(id string, key string) (interface{}, int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.load(id)
	if err != nil {
		return nil, 0, false
	}

	raw, ok := data[key]
	if !ok {
		return nil, 0, false
	}

	// Tenta converter de volta do formato versionado
	if vv, okvv := raw.(map[string]interface{}); okvv {
		val := vv["value"]
		ver := 0
		if v, okv := vv["version"].(float64); okv {
			ver = int(v)
		}
		return val, ver, true
	}

	return raw, 0, true
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
