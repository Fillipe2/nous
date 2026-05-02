package bootstrap

import (
	"context"
	"log"
	"time"
)

const (
	intervaloExpurgoRefreshTokens = 24 * time.Hour
	retencaoRefreshTokens         = 24 * time.Hour
)

func (c *Container) StartCleanupJobs(ctx context.Context) {
	if c.UsuarioRepository == nil {
		return
	}
	go c.expurgarRefreshTokensPeriodicamente(ctx)
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
