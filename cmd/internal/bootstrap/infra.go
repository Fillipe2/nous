package bootstrap

import (
	"context"
	"fmt"
	"log"

	"buildnix-backend/internal/infra/config"
	"buildnix-backend/internal/infra/database/postgres"
	"buildnix-backend/internal/infra/eventbus"
	infrajwt "buildnix-backend/internal/infra/jwt"
	"buildnix-backend/internal/infra/logging"
)

func (c *Container) initInfra(ctx context.Context) error {
	cfg := config.Carregar()
	c.Config = cfg
	if err := cfg.Validar(); err != nil {
		return fmt.Errorf("configuracao invalida: %w", err)
	}
	logCloser, err := logging.Configure(logging.Config{
		FilePath:   cfg.LogFilePath,
		MaxSizeMB:  cfg.LogMaxSizeMB,
		MaxAgeDays: cfg.LogMaxAgeDays,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
		Stdout:     cfg.LogStdout,
	})
	if err != nil {
		return fmt.Errorf("configurar logs: %w", err)
	}
	c.LogCloser = logCloser
	if cfg.ResetPasswordReturnToken {
		log.Printf("AVISO: RESET_PASSWORD_RETURN_TOKEN=true retorna token de redefinicao na resposta HTTP; use apenas em desenvolvimento/testes")
	}

	database, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("conectar banco de dados: %w", err)
	}
	c.Database = database

	c.Publisher = eventbus.NewObserverPublisher(cfg.ObserverURL)
	c.JWTServico = infrajwt.NewServico(cfg.JWTSecret, cfg.JWTRefreshSecret)

	return nil
}
