package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/evento"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"
)

type EventOutboxRepository struct {
	connection *Postgres
}

var _ evento.OutboxRepository = (*EventOutboxRepository)(nil)

func NewEventOutboxRepository(connection *Postgres) (*EventOutboxRepository, error) {
	return &EventOutboxRepository{connection: connection}, nil
}

func (r *EventOutboxRepository) Salvar(ctx context.Context, ev evento.Evento) error {
	payload, err := json.Marshal(ev.Payload)
	if err != nil {
		return fmt.Errorf("serializar payload outbox: %w", err)
	}

	stms := statement.Get()
	if err := r.connection.Exec(ctx, stms.SalvarEventoOutbox, pgx.NamedArgs{
		"event_id":        ev.EventId,
		"event_type":      ev.EventType,
		"schema_version":  ev.SchemaVersion,
		"source":          ev.Source,
		"occurred_at":     ev.OccurredAt,
		"idempotency_key": ev.IdempotencyKey,
		"payload":         payload,
	}); err != nil {
		return fmt.Errorf("salvar evento outbox: %w", err)
	}
	return nil
}

func (r *EventOutboxRepository) MarcarPublicado(ctx context.Context, eventId string) error {
	stms := statement.Get()
	if err := r.connection.Exec(ctx, stms.MarcarEventoOutboxPublicado, pgx.NamedArgs{"event_id": eventId}); err != nil {
		return fmt.Errorf("marcar evento outbox publicado: %w", err)
	}
	return nil
}

func (r *EventOutboxRepository) MarcarFalha(ctx context.Context, eventId string, mensagemErro string) error {
	stms := statement.Get()
	if err := r.connection.Exec(ctx, stms.MarcarEventoOutboxFalha, pgx.NamedArgs{
		"event_id":      eventId,
		"mensagem_erro": mensagemErro,
	}); err != nil {
		return fmt.Errorf("marcar evento outbox falha: %w", err)
	}
	return nil
}
