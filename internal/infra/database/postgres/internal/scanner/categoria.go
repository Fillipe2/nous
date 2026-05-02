package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/categoria"
)

// Categoria escaneia uma row do banco para a entidade de domínio Categoria.
func Categoria(row pgx.Row) (*categoria.Categoria, error) {
	var (
		id           string
		nome         string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &nome, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	return &categoria.Categoria{
		Id:           id,
		Nome:         nome,
		Status:       categoria.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}, nil
}
