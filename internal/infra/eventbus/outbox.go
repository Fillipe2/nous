package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log"

	"nous/internal/domain/evento"
)

type OutboxPublisher struct {
	repository evento.OutboxRepository
	next       evento.Publisher
}

func NewOutboxPublisher(repository evento.OutboxRepository, next evento.Publisher) evento.Publisher {
	return &OutboxPublisher{repository: repository, next: next}
}

func (p *OutboxPublisher) Publicar(ctx context.Context, ev evento.Evento) error {
	if err := p.repository.Salvar(ctx, ev); err != nil {
		return fmt.Errorf("salvar evento outbox: %w", err)
	}

	if err := p.next.Publicar(ctx, ev); err != nil {
		if errors.Is(err, ErrObserverNaoConfigurado) {
			return nil
		}
		if markErr := p.repository.MarcarFalha(ctx, ev.EventId, err.Error()); markErr != nil {
			log.Printf("[eventbus] erro ao marcar falha no outbox: %v", markErr)
		}
		return err
	}

	if err := p.repository.MarcarPublicado(ctx, ev.EventId); err != nil {
		return fmt.Errorf("marcar evento outbox publicado: %w", err)
	}
	return nil
}
