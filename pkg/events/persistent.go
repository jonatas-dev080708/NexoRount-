package events

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// PersistentNexus é uma camada de persistência sobre o LocalNexus
type PersistentNexus struct {
	*LocalNexus
	db *sql.DB
	mu sync.Mutex
}

func NewPersistentNexus(dbPath string) (*PersistentNexus, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Cria tabela de eventos se não existir
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS nexo_events (
			id TEXT PRIMARY KEY,
			type TEXT,
			source TEXT,
			payload TEXT,
			trace_id TEXT,
			timestamp INTEGER,
			processed INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return nil, err
	}

	return &PersistentNexus{
		LocalNexus: NewLocalNexus(),
		db:         db,
	}, nil
}

// Publish salva o evento antes de propagar
func (pn *PersistentNexus) Publish(e Event) {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	// 1. Salva no banco
	payloadJSON, _ := json.Marshal(e.Payload)
	_, err := pn.db.Exec(`
		INSERT INTO nexo_events (id, type, source, payload, trace_id, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, e.ID, string(e.Type), e.Source, string(payloadJSON), e.TraceID, e.Timestamp)

	if err != nil {
		fmt.Printf("⚠️ [PersistentNexus] Erro ao persistir evento: %v\n", err)
	}

	// 2. Propaga no Nexus local
	pn.LocalNexus.Publish(e)
}

// Ack marca o evento como processado
func (pn *PersistentNexus) Ack(eventID string) error {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	_, err := pn.db.Exec("UPDATE nexo_events SET processed = 1 WHERE id = ?", eventID)
	return err
}

// Restore recupera eventos não processados e os re-emite
func (pn *PersistentNexus) Restore() error {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	rows, err := pn.db.Query("SELECT id, type, source, payload, trace_id, timestamp FROM nexo_events WHERE processed = 0")
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var e Event
		var payloadStr string
		var eventType string
		
		err := rows.Scan(&e.ID, &eventType, &e.Source, &payloadStr, &e.TraceID, &e.Timestamp)
		if err != nil {
			continue
		}
		
		e.Type = Type(eventType)
		json.Unmarshal([]byte(payloadStr), &e.Payload)

		// Re-emite no Nexus local
		pn.LocalNexus.Publish(e)
		count++
	}

	if count > 0 {
		fmt.Printf("♻️  [PersistentNexus] %d eventos pendentes restaurados com sucesso.\n", count)
	}
	return nil
}

func (pn *PersistentNexus) Close() {
	pn.db.Close()
	pn.LocalNexus.Close()
}
