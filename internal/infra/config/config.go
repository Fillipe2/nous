package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	segredoPadraoAccess  = "troque-em-producao"
	segredoPadraoRefresh = "troque-refresh-em-producao"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTRefreshSecret   string
	ObserverURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	ServerPort         string
	MigrationsDirPath  string
	AdminNome          string
	AdminEmail         string
	AdminSenha         string
	LogFilePath        string
	LogMaxSizeMB       int
	LogMaxAgeDays      int
	LogMaxBackups      int
	LogCompress        bool
	LogStdout          bool
}

func Carregar() *Config {
	return &Config{
		DatabaseURL:        obterEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/buildnix?sslmode=disable"),
		JWTSecret:          obterEnv("JWT_SECRET", ""),
		JWTRefreshSecret:   obterEnv("JWT_REFRESH_SECRET", ""),
		ObserverURL:        obterEnv("OBSERVER_URL", ""),
		GoogleClientID:     obterEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: obterEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  obterEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		ServerPort:         obterEnv("SERVER_PORT", "8080"),
		MigrationsDirPath:  obterEnv("MIGRATIONS_DIR", "./migrations/postgres"),
		AdminNome:          obterEnv("ADMIN_NOME", "Administrador"),
		AdminEmail:         obterEnv("ADMIN_EMAIL", ""),
		AdminSenha:         obterEnv("ADMIN_SENHA", ""),
		LogFilePath:        obterEnv("LOG_FILE_PATH", "./logs/buildnix-backend.log"),
		LogMaxSizeMB:       obterEnvInt("LOG_MAX_SIZE_MB", 50),
		LogMaxAgeDays:      obterEnvInt("LOG_MAX_AGE_DAYS", 30),
		LogMaxBackups:      obterEnvInt("LOG_MAX_BACKUPS", 10),
		LogCompress:        obterEnvBool("LOG_COMPRESS", true),
		LogStdout:          obterEnvBool("LOG_STDOUT", true),
	}
}

func (c *Config) Validar() error {
	if segredoInseguro(c.JWTSecret, segredoPadraoAccess) {
		return fmt.Errorf("JWT_SECRET obrigatorio e deve ser forte")
	}
	if segredoInseguro(c.JWTRefreshSecret, segredoPadraoRefresh) {
		return fmt.Errorf("JWT_REFRESH_SECRET obrigatorio e deve ser forte")
	}
	if c.JWTSecret == c.JWTRefreshSecret {
		return fmt.Errorf("JWT_SECRET e JWT_REFRESH_SECRET devem ser diferentes")
	}
	if strings.TrimSpace(c.LogFilePath) != "" && c.LogMaxSizeMB <= 0 {
		return fmt.Errorf("LOG_MAX_SIZE_MB deve ser maior que zero")
	}
	if c.LogMaxAgeDays < 0 {
		return fmt.Errorf("LOG_MAX_AGE_DAYS nao pode ser negativo")
	}
	if c.LogMaxBackups < 0 {
		return fmt.Errorf("LOG_MAX_BACKUPS nao pode ser negativo")
	}
	return nil
}

func segredoInseguro(valor, padrao string) bool {
	valor = strings.TrimSpace(valor)
	return valor == "" || valor == padrao || len(valor) < 32
}

func obterEnv(chave, padrao string) string {
	if valor, ok := os.LookupEnv(chave); ok {
		return valor
	}
	return padrao
}

func obterEnvInt(chave string, padrao int) int {
	valor, ok := os.LookupEnv(chave)
	if !ok || strings.TrimSpace(valor) == "" {
		return padrao
	}
	inteiro, err := strconv.Atoi(valor)
	if err != nil {
		return padrao
	}
	return inteiro
}

func obterEnvBool(chave string, padrao bool) bool {
	valor, ok := os.LookupEnv(chave)
	if !ok || strings.TrimSpace(valor) == "" {
		return padrao
	}
	booleano, err := strconv.ParseBool(valor)
	if err != nil {
		return padrao
	}
	return booleano
}
