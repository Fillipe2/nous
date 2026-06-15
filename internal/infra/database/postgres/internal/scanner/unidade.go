package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/unidade"
)

// Unidade escaneia uma row do banco para a entidade de domínio Unidade.
func Unidade(row pgx.Row) (*unidade.Unidade, error) {
	var (
		id           string
		negocioID    string
		nome         string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &negocioID, &nome, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	return &unidade.Unidade{
		Id:           id,
		NegocioId:    negocioID,
		Nome:         nome,
		Status:       unidade.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}, nil
}
