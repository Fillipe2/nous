package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/cliente"
)

// Cliente escaneia uma row do banco para a entidade de domínio Cliente.
func Cliente(row pgx.Row) (*cliente.Cliente, error) {
	var (
		id           string
		nome         string
		email        string
		telefone     *string
		whatsapp     *string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &nome, &email, &telefone, &whatsapp, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	c := &cliente.Cliente{
		Id:           id,
		Nome:         nome,
		Email:        email,
		Status:       cliente.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}
	if telefone != nil {
		c.Telefone = *telefone
	}
	if whatsapp != nil {
		c.Whatsapp = *whatsapp
	}
	return c, nil
}
