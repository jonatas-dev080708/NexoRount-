package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jonatas-dev080708/NexoRount-/pkg/events"
)

// Bind associa um evento a um handler com tipagem forte usando Generics.
// Isso elimina a necessidade de fazer casts manuais de interface{} dentro dos handlers.
func Bind[T any](a *BaseAgent, eventType events.Type, handler func(a *BaseAgent, ctx context.Context, payload T)) {
	a.On(eventType, func(agent *BaseAgent, ctx context.Context, e events.Event) {
		var payload T

		// 1. Tenta cast direto (mais performático)
		if val, ok := e.Payload.(T); ok {
			handler(agent, ctx, val)
			return
		}

		// 2. Tenta conversão via JSON (mais flexível para eventos vindos de fora/Nexus remoto)
		// Isso permite que um map[string]interface{} seja convertido em uma struct tipada.
		jsonData, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Printf("⚠️ [Bind] Erro ao serializar payload para %s: %v\n", eventType, err)
			return
		}

		if err := json.Unmarshal(jsonData, &payload); err != nil {
			fmt.Printf("⚠️ [Bind] Erro ao desserializar payload para o tipo esperado em %s: %v\n", eventType, err)
			return
		}

		handler(agent, ctx, payload)
	})
}

// BindResponse é uma variação que facilita o envio de uma resposta de volta para o Nexus
func BindResponse[T any, R any](a *BaseAgent, eventType events.Type, handler func(a *BaseAgent, ctx context.Context, p T) (R, error)) {
	a.On(eventType, func(agent *BaseAgent, ctx context.Context, e events.Event) {
		var payload T
		
		// Lógica de conversão similar ao Bind...
		jsonData, _ := json.Marshal(e.Payload)
		json.Unmarshal(jsonData, &payload)

		result, err := handler(agent, ctx, payload)
		if err != nil {
			agent.Emit(events.Type(string(eventType)+"_ERROR"), err.Error())
			return
		}

		agent.Emit(events.Type(string(eventType)+"_COMPLETED"), result)
	})
}
