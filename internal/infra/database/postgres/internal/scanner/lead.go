package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/lead"
)

// Lead escaneia uma row do banco para a entidade de domínio Lead.
func Lead(row pgx.Row) (*lead.Lead, error) {
	var (
		id                   string
		landingPageID        string
		nome                 string
		email                string
		telefone             *string
		whatsapp             *string
		mensagem             *string
		camposPersonalizados []byte
		origem               string
		ip                   *string
		userAgent            *string
		status               string
		criadoEm             time.Time
	)
	if err := row.Scan(
		&id, &landingPageID, &nome, &email, &telefone, &whatsapp, &mensagem,
		&camposPersonalizados, &origem, &ip, &userAgent, &status, &criadoEm,
	); err != nil {
		return nil, err
	}
	l := &lead.Lead{
		Id:                   id,
		LandingPageId:        landingPageID,
		Nome:                 nome,
		Email:                email,
		CamposPersonalizados: json.RawMessage(camposPersonalizados),
		Origem:               origem,
		Status:               lead.Status(status),
		CriadoEm:             criadoEm,
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
	return l, nil
}
