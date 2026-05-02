package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/negocio"
)

// Negocio escaneia uma row do banco para a entidade de domínio Negocio.
func Negocio(row pgx.Row) (*negocio.Negocio, error) {
	var (
		id           string
		clienteID    string
		nome         string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &clienteID, &nome, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	return &negocio.Negocio{
		Id:           id,
		ClienteId:    clienteID,
		Nome:         nome,
		Status:       negocio.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}, nil
}
