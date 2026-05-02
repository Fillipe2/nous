package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/cliente"
)

// Cliente escaneia uma row do banco para a entidade de domínio Cliente.
func Cliente(row pgx.Row) (*cliente.Cliente, error) {
	var (
		id           string
		nome         string
		email        string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &nome, &email, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	return &cliente.Cliente{
		Id:           id,
		Nome:         nome,
		Email:        email,
		Status:       cliente.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}, nil
}
