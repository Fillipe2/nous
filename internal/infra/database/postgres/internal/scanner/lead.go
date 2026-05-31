package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/lead"
)

// Lead escaneia uma row do banco para a entidade de domínio Lead.
func Lead(row pgx.Row) (*lead.Lead, error) {
	var (
		id                              string
		landingPageID                   string
		clienteID                       *string
		nome                            string
		email                           string
		telefone                        *string
		whatsapp                        *string
		mensagem                        *string
		camposPersonalizados            []byte
		origem                          string
		ip                              *string
		userAgent                       *string
		referrer                        *string
		utmSource                       *string
		utmMedium                       *string
		utmCampaign                     *string
		utmContent                      *string
		utmTerm                         *string
		consentimento                   bool
		consentimentoEm                 *time.Time
		politicaVersao                  *string
		consentimentoMarketingParceiros bool
		consentimentoVersaoMarketing    *string
		status                          string
		perdidoMotivo                   *string
		anonimizadoEm                   *time.Time
		criadoEm                        time.Time
		atualizadoEm                    time.Time
	)
	if err := row.Scan(
		&id, &landingPageID, &clienteID, &nome, &email, &telefone, &whatsapp, &mensagem,
		&camposPersonalizados, &origem, &ip, &userAgent, &referrer,
		&utmSource, &utmMedium, &utmCampaign, &utmContent, &utmTerm,
		&consentimento, &consentimentoEm, &politicaVersao,
		&consentimentoMarketingParceiros, &consentimentoVersaoMarketing,
		&status, &perdidoMotivo, &anonimizadoEm, &criadoEm, &atualizadoEm,
	); err != nil {
		return nil, err
	}
	l := &lead.Lead{
		Id:                              id,
		LandingPageId:                   landingPageID,
		Nome:                            nome,
		Email:                           email,
		CamposPersonalizados:            json.RawMessage(camposPersonalizados),
		Origem:                          origem,
		Consentimento:                   consentimento,
		ConsentimentoEm:                 consentimentoEm,
		ConsentimentoMarketingParceiros: consentimentoMarketingParceiros,
		Status:                          lead.Status(status),
		AnonimizadoEm:                   anonimizadoEm,
		CriadoEm:                        criadoEm,
		AtualizadoEm:                    atualizadoEm,
	}
	if clienteID != nil {
		l.ClienteId = *clienteID
	}
	if telefone != nil {
		l.Telefone = *telefone
	}
	if whatsapp != nil {
		l.Whatsapp = *whatsapp
	}
	if mensagem != nil {
		l.Mensagem = *mensagem
	}
	if ip != nil {
		l.IP = *ip
	}
	if userAgent != nil {
		l.UserAgent = *userAgent
	}
	if referrer != nil {
		l.Referrer = *referrer
	}
	if utmSource != nil {
		l.UTM.Source = *utmSource
	}
	if utmMedium != nil {
		l.UTM.Medium = *utmMedium
	}
	if utmCampaign != nil {
		l.UTM.Campaign = *utmCampaign
	}
	if utmContent != nil {
		l.UTM.Content = *utmContent
	}
	if utmTerm != nil {
		l.UTM.Term = *utmTerm
	}
	if politicaVersao != nil {
		l.PoliticaVersao = *politicaVersao
	}
	if consentimentoVersaoMarketing != nil {
		l.ConsentimentoVersaoMarketing = *consentimentoVersaoMarketing
	}
	if perdidoMotivo != nil {
		l.PerdidoMotivo = *perdidoMotivo
	}
	return l, nil
}
