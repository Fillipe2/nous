package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/prospecto"
)

// Prospecto escaneia uma row do banco para a entidade de domínio Prospecto.
func Prospecto(row pgx.Row) (*prospecto.Prospecto, error) {
	var (
		id                       string
		nome                     string
		email                    string
		telefone                 *string
		categoriaPretendidaId    *string
		categoriaPretendidaLivre *string
		cidade                   *string
		estado                   *string
		porte                    *string
		interesses               []byte
		mensagem                 *string
		origem                   *string
		utmSource                *string
		utmMedium                *string
		utmCampaign              *string
		utmContent               *string
		utmTerm                  *string
		statusFunil              string
		convertidoEmUsuarioId    *string
		ip                       *string
		userAgent                *string
		criadoEm                 time.Time
		atualizadoEm             time.Time
	)
	if err := row.Scan(
		&id, &nome, &email, &telefone,
		&categoriaPretendidaId, &categoriaPretendidaLivre,
		&cidade, &estado, &porte, &interesses, &mensagem, &origem,
		&utmSource, &utmMedium, &utmCampaign, &utmContent, &utmTerm,
		&statusFunil, &convertidoEmUsuarioId,
		&ip, &userAgent, &criadoEm, &atualizadoEm,
	); err != nil {
		return nil, err
	}
	p := &prospecto.Prospecto{
		Id:          id,
		Nome:        nome,
		Email:       email,
		StatusFunil: prospecto.StatusFunil(statusFunil),
		Interesses:  json.RawMessage(interesses),
		CriadoEm:   criadoEm,
		AtualizadoEm: atualizadoEm,
	}
	if telefone != nil {
		p.Telefone = *telefone
	}
	if categoriaPretendidaId != nil {
		p.CategoriaPretendidaId = *categoriaPretendidaId
	}
	if categoriaPretendidaLivre != nil {
		p.CategoriaPretendidaLivre = *categoriaPretendidaLivre
	}
	if cidade != nil {
		p.Cidade = *cidade
	}
	if estado != nil {
		p.Estado = *estado
	}
	if porte != nil {
		p.Porte = prospecto.Porte(*porte)
	}
	if mensagem != nil {
		p.Mensagem = *mensagem
	}
	if origem != nil {
		p.Origem = *origem
	}
	if utmSource != nil {
		p.UTMSource = *utmSource
	}
	if utmMedium != nil {
		p.UTMMedium = *utmMedium
	}
	if utmCampaign != nil {
		p.UTMCampaign = *utmCampaign
	}
	if utmContent != nil {
		p.UTMContent = *utmContent
	}
	if utmTerm != nil {
		p.UTMTerm = *utmTerm
	}
	if convertidoEmUsuarioId != nil {
		p.ConvertidoEmUsuarioId = *convertidoEmUsuarioId
	}
	if ip != nil {
		p.IP = *ip
	}
	if userAgent != nil {
		p.UserAgent = *userAgent
	}
	return p, nil
}
