package eventbus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"nous/internal/domain/evento"
)

var ErrObserverNaoConfigurado = errors.New("observer nao configurado")

// ObserverPublisher publica eventos HTTP para o Observer central do ecossistema Sognix.
// Quando OBSERVER_URL não está configurado, opera em modo no-op (log apenas).
type ObserverPublisher struct {
	observerURL string
	httpClient  *http.Client
}

func NewObserverPublisher(observerURL string) evento.Publisher {
	return &ObserverPublisher{
		observerURL: observerURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *ObserverPublisher) Publicar(ctx context.Context, ev evento.Evento) error {
	if p.observerURL == "" {
		return ErrObserverNaoConfigurado
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("observer publisher serializar evento: %w", err)
	}

	backoffs := []time.Duration{0, 200 * time.Millisecond, time.Second}
	var lastErr error
	for _, backoff := range backoffs {
		if backoff > 0 {
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.observerURL+"/api/v1/eventos", bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("observer publisher criar request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return nil
		}
		lastErr = fmt.Errorf("observer retornou status %d", resp.StatusCode)
	}

	return fmt.Errorf("observer publisher publicar evento %s: %w", ev.EventId, lastErr)
}
