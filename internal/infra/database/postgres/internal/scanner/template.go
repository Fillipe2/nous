package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/template"
)

// Template escaneia uma row do banco para a entidade de domínio Template.
func Template(row pgx.Row) (*template.Template, error) {
	var (
		id           string
		clienteID    *string
		categoriaID  *string
		tipo         string
		origem       string
		nome         string
		arquivoURL   *string
		conteudo     string
		status       string
		criadoEm     time.Time
		atualizadoEm time.Time
		version      int
	)
	if err := row.Scan(&id, &clienteID, &categoriaID, &tipo, &origem, &nome, &arquivoURL, &conteudo, &status, &criadoEm, &atualizadoEm, &version); err != nil {
		return nil, err
	}
	t := &template.Template{
		Id:           id,
		Tipo:         template.Tipo(tipo),
		Origem:       template.Origem(origem),
		Nome:         nome,
		Conteudo:     conteudo,
		Status:       template.Status(status),
		CriadoEm:     criadoEm,
		AtualizadoEm: atualizadoEm,
		Version:      version,
	}
	if clienteID != nil {
		t.ClienteId = *clienteID
	}
	if categoriaID != nil {
		t.CategoriaId = *categoriaID
	}
	if arquivoURL != nil {
		t.ArquivoURL = *arquivoURL
	}
	return t, nil
}
