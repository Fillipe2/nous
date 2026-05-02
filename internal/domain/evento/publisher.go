package evento

import (
	"context"
	"time"
)

// Evento representa uma mensagem versionada publicada para o Observer
// quando os outros servicos do ecossistema Sognix estiverem ativos.
type Evento struct {
	IdempotencyKey string                 `json:"idempotency_key"`
	Payload        map[string]interface{} `json:"payload"`
	EventId        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	Source         string                 `json:"source"`
	OccurredAt     time.Time              `json:"occurred_at"`
	SchemaVersion  int                    `json:"schema_version"`
}

// Publisher é a interface para publicar eventos para o Observer.
// Quando o Observer não estiver configurado, a implementação é no-op.
type Publisher interface {
	Publicar(ctx context.Context, evento Evento) error
}

// OutboxRepository persiste eventos localmente antes da tentativa de entrega externa.
type OutboxRepository interface {
	Salvar(ctx context.Context, evento Evento) error
	MarcarPublicado(ctx context.Context, eventId string) error
	MarcarFalha(ctx context.Context, eventId string, mensagemErro string) error
}
