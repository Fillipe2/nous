package bootstrap

import (
	"context"
	"errors"
	"log"
	"time"

	"buildnix-backend/internal/infra/eventbus"
)

const (
	intervaloExpurgoRefreshTokens = 24 * time.Hour
	retencaoRefreshTokens         = 24 * time.Hour

	intervaloRetryOutbox = 5 * time.Minute
	limiteRetryOutbox    = 50

	intervaloAnonimizacaoLGPD = 24 * time.Hour
	retencaoLeadsDias         = 365
)

func (c *Container) StartCleanupJobs(ctx context.Context) {
	if c.UsuarioRepository == nil {
		return
	}
	go c.expurgarRefreshTokensPeriodicamente(ctx)
	go c.retentarEventosOutboxPeriodicamente(ctx)
	go c.anonimizarLeadsExpiradosPeriodicamente(ctx)
}

func (c *Container) expurgarRefreshTokensPeriodicamente(ctx context.Context) {
	executar := func() {
		cleanupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		antesDe := time.Now().Add(-retencaoRefreshTokens)
		if err := c.UsuarioRepository.ExpurgarRefreshTokens(cleanupCtx, antesDe); err != nil {
			log.Printf("expurgar refresh tokens: %v", err)
		}
	}

	executar()
	ticker := time.NewTicker(intervaloExpurgoRefreshTokens)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			executar()
		}
	}
}

// retentarEventosOutboxPeriodicamente tenta re-publicar eventos com status 'falha'
// ou 'pendente' travados (criados há mais de 2 minutos sem publicação).
func (c *Container) retentarEventosOutboxPeriodicamente(ctx context.Context) {
	if c.EventOutboxRepository == nil || c.Publisher == nil {
		return
	}

	executar := func() {
		cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		pendentes, err := c.EventOutboxRepository.BuscarPendentes(cleanupCtx, limiteRetryOutbox)
		if err != nil {
			log.Printf("buscar eventos outbox pendentes: %v", err)
			return
		}

		for _, ev := range pendentes {
			if err := c.Publisher.Publicar(cleanupCtx, ev); err != nil {
				// Erro já registrado pelo OutboxPublisher; apenas loga resumo.
				if !errors.Is(err, eventbus.ErrObserverNaoConfigurado) {
					log.Printf("retry outbox event_id=%s type=%s: %v", ev.EventId, ev.EventType, err)
				}
			}
		}
	}

	ticker := time.NewTicker(intervaloRetryOutbox)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			executar()
		}
	}
}

// anonimizarLeadsExpiradosPeriodicamente anonimiza leads cujo consentimento
// tenha expirado há mais de 365 dias, conforme política LGPD.
func (c *Container) anonimizarLeadsExpiradosPeriodicamente(ctx context.Context) {
	if c.LeadRepository == nil {
		return
	}

	executar := func() {
		cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		antesDe := time.Now().AddDate(0, 0, -retencaoLeadsDias)
		n, err := c.LeadRepository.AnonimizarExpirados(cleanupCtx, antesDe)
		if err != nil {
			log.Printf("anonimizar leads expirados LGPD: %v", err)
			return
		}
		if n > 0 {
			log.Printf("anonimizados %d leads expirados (LGPD)", n)
		}
	}

	executar()
	ticker := time.NewTicker(intervaloAnonimizacaoLGPD)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			executar()
		}
	}
}
