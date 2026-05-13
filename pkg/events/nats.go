package events

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSNexus é a implementação distribuída do barramento de eventos
type NATSNexus struct {
	nc *nats.Conn
}

func NewNATSNexus(url string) (*NATSNexus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no NATS: %v", err)
	}
	return &NATSNexus{nc: nc}, nil
}

func (n *NATSNexus) Subscribe(eventType Type) chan Event {
	ch := make(chan Event, 100)
	
	// Subscreve no tópico do NATS correspondente ao tipo de evento
	_, err := n.nc.Subscribe(string(eventType), func(m *nats.Msg) {
		var ev Event
		if err := json.Unmarshal(m.Data, &ev); err != nil {
			return
		}
		ch <- ev
	})

	if err != nil {
		fmt.Printf("⚠️ Erro ao assinar tópico NATS: %v\n", err)
	}

	return ch
}

func (n *NATSNexus) Publish(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	
	// Publica no NATS para que todas as instâncias do ecossistema recebam
	n.nc.Publish(string(event.Type), data)
}

func (n *NATSNexus) Close() {
	n.nc.Close()
}
