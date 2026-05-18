package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	domain "buildnix-backend/internal/domain/lead_aquisicao"
)

// LeadAquisicao escaneia uma row do banco para a entidade de domínio LeadAquisicao.
func LeadAquisicao(row pgx.Row) (*domain.LeadAquisicao, error) {
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
	la := &domain.LeadAquisicao{
		Id:           id,
		Nome:         nome,
		Email:        email,
		StatusFunil:  domain.StatusFunil(statusFunil),
		Interesses:   json.RawMessage(interesses),
		CriadoEm:    criadoEm,
		AtualizadoEm: atualizadoEm,
	}
	if telefone != nil {
		la.Telefone = *telefone
	}
	if categoriaPretendidaId != nil {
		la.CategoriaPretendidaId = *categoriaPretendidaId
	}
	if categoriaPretendidaLivre != nil {
		la.CategoriaPretendidaLivre = *categoriaPretendidaLivre
	}
	if cidade != nil {
		la.Cidade = *cidade
	}
	if estado != nil {
		la.Estado = *estado
	}
	if porte != nil {
		la.Porte = domain.Porte(*porte)
	}
	if mensagem != nil {
		la.Mensagem = *mensagem
	}
	if origem != nil {
		la.Origem = *origem
	}
	if utmSource != nil {
		la.UTMSource = *utmSource
	}
	if utmMedium != nil {
		la.UTMMedium = *utmMedium
	}
	if utmCampaign != nil {
		la.UTMCampaign = *utmCampaign
	}
	if utmContent != nil {
		la.UTMContent = *utmContent
	}
	if utmTerm != nil {
		la.UTMTerm = *utmTerm
	}
	if convertidoEmUsuarioId != nil {
		la.ConvertidoEmUsuarioId = *convertidoEmUsuarioId
	}
	if ip != nil {
		la.IP = *ip
	}
	if userAgent != nil {
		la.UserAgent = *userAgent
	}
	return la, nil
}
