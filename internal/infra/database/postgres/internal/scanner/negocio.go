package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/negocio"
)

// Negocio escaneia uma row do banco para a entidade de domínio Negocio.
func Negocio(row pgx.Row) (*negocio.Negocio, error) {
	var (
		id           string
		nome         string
		email        string
		telefone     *string
		whatsapp     *string
		documento    *string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &nome, &email, &telefone, &whatsapp, &documento, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	c := &negocio.Negocio{
		Id:           id,
		Nome:         nome,
		Email:        email,
		Status:       negocio.Status(status),
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
	if documento != nil {
		c.Documento = *documento
	}
	return c, nil
}
