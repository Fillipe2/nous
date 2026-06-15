package bootstrap

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"nous/internal/infra/config"
	"nous/internal/infra/database/postgres"
	"nous/internal/infra/eventbus"
	infrajwt "nous/internal/infra/jwt"
	"nous/internal/infra/logging"
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
	if cfg.EmailVerificationReturnToken {
		log.Printf("AVISO: EMAIL_VERIFICATION_RETURN_TOKEN=true retorna token de verificacao na resposta HTTP; use apenas em desenvolvimento/testes")
	}

	database, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("conectar banco de dados: %w", err)
	}
	c.Database = database

	c.Publisher = eventbus.NewObserverPublisher(cfg.ObserverURL)

	if cfg.JWTRSAPrivateKey != "" {
		rsaKey, err := parsearChaveRSA(cfg.JWTRSAPrivateKey)
		if err != nil {
			return fmt.Errorf("JWT_RSA_PRIVATE_KEY invalida: %w", err)
		}
		c.JWTServico = infrajwt.NewServicoRSA(cfg.JWTSecret, cfg.JWTRefreshSecret, rsaKey)
		log.Printf("JWT: usando RS256 com chave RSA (kid=%s)", c.JWTServico.KidRSA())
	} else {
		c.JWTServico = infrajwt.NewServico(cfg.JWTSecret, cfg.JWTRefreshSecret)
	}

	return nil
}

func parsearChaveRSA(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("bloco PEM invalido")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Tenta PKCS1 se PKCS8 falhar
		rsaKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parsear chave privada RSA: %w", err)
		}
		return rsaKey, nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("chave PEM nao e RSA")
	}
	return rsaKey, nil
}
