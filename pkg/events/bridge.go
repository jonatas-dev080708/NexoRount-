package events

import (
	"fmt"
)

// Bridge conecta dois barramentos de eventos diferentes
// Isso permite que eventos locais sejam replicados para um Nexus remoto (como NATS)
// e vice-versa, criando um runtime híbrido.
type Bridge struct {
	local  Bus
	remote Bus
	types  []Type
}

func NewBridge(local, remote Bus) *Bridge {
	return &Bridge{
		local:  local,
		remote: remote,
		types:  make([]Type, 0),
	}
}

// Forward ativa o encaminhamento bidirecional para tipos específicos de eventos
func (b *Bridge) Forward(eventTypes ...Type) {
	for _, et := range eventTypes {
		// Local -> Remote
		go func(t Type) {
			ch := b.local.Subscribe(t)
			for ev := range ch {
				// Evita loop infinito: não re-encaminha o que veio do próprio bridge
				if ev.Metadata != nil && ev.Metadata["bridged"] == true {
					continue
				}
				
				if ev.Metadata == nil {
					ev.Metadata = make(map[string]interface{})
				}
				ev.Metadata["bridged"] = true
				b.remote.Publish(ev)
			}
		}(et)

		// Remote -> Local
		go func(t Type) {
			ch := b.remote.Subscribe(t)
			for ev := range ch {
				if ev.Metadata != nil && ev.Metadata["bridged"] == true {
					continue
				}

				if ev.Metadata == nil {
					ev.Metadata = make(map[string]interface{})
				}
				ev.Metadata["bridged"] = true
				b.local.Publish(ev)
			}
		}(et)
		
		fmt.Printf("🌉 Bridge: Encaminhando eventos do tipo [%s] entre Local e Remote\n", et)
	}
}
