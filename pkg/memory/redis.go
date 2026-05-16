package memory

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/redis/go-redis/v9"
)

// RedisStore implementa a memória semântica usando Redis com RediSearch
type RedisStore struct {
	client *redis.Client
	index  string
}

func NewRedisStore(addr, password string, index string) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
		index: index,
	}
}

func (r *RedisStore) SaveVector(ctx context.Context, agentID, content string, vector []float32, metadata map[string]interface{}) error {
	key := fmt.Sprintf("vec:%s:%x", agentID, content)
	
	// Redis exige o vetor como blob binário
	vecBlob := make([]byte, len(vector)*4)
	for i, v := range vector {
		binary.LittleEndian.PutUint32(vecBlob[i*4:], math.Float32bits(v))
	}

	data := map[string]interface{}{
		"agent_id": agentID,
		"content":  content,
		"vector":   vecBlob,
	}

	return r.client.HSet(ctx, key, data).Err()
}

func (r *RedisStore) Search(ctx context.Context, agentID string, vector []float32, limit int) ([]Document, error) {
	// Converte vetor para blob
	vecBlob := make([]byte, len(vector)*4)
	for i, v := range vector {
		binary.LittleEndian.PutUint32(vecBlob[i*4:], math.Float32bits(v))
	}

	// Comando FT.SEARCH do RediSearch para KNN
	query := fmt.Sprintf("(@agent_id:{%s})=>[KNN %d @vector $vec AS score]", agentID, limit)
	
	res := r.client.Do(ctx, "FT.SEARCH", r.index, query, "PARAMS", 2, "vec", vecBlob, "DIALECT", 2)
	if res.Err() != nil {
		return nil, res.Err()
	}

	// O parse do retorno do FT.SEARCH é complexo (lista de listas), simplificado aqui
	return []Document{}, nil
}
