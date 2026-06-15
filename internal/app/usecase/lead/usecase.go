package lead

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"nous/internal/app/dto"
	"nous/internal/domain/evento"
	"nous/internal/domain/landingpage"
	domain "nous/internal/domain/lead"
)

const (
	limiteMensagem            = 2000
	limiteCamposPersonalizado = 8 * 1024
	janelaDedupe              = 5 * time.Minute
	retencaoLeadPadrao        = 365 * 24 * time.Hour
)

type UseCase struct {
	repository            domain.Repository
	landingPageRepository landingpage.Repository
	publisher             evento.Publisher
}

func NewUseCase(repository domain.Repository, landingPageRepository landingpage.Repository, publisher evento.Publisher) *UseCase {
	return &UseCase{repository: repository, landingPageRepository: landingPageRepository, publisher: publisher}
}

func (u *UseCase) CapturarLead(ctx context.Context, input dto.CapturarLeadInput, ip, userAgent string) (*dto.LeadOutput, error) {
	if strings.TrimSpace(input.Honeypot) != "" {
		return nil, domain.ErrSpamDetectado
	}

	input.Nome = strings.TrimSpace(input.Nome)
	if input.Nome == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Email == "" || !emailValido(input.Email) {
		return nil, fmt.Errorf("email valido e obrigatorio")
	}
	input.Telefone = normalizarTelefone(input.Telefone)
	input.Whatsapp = normalizarTelefone(input.Whatsapp)
	if input.Telefone == "" && input.Whatsapp == "" {
		return nil, fmt.Errorf("telefone ou whatsapp e obrigatorio")
	}
	if input.LandingPageId == "" {
		return nil, fmt.Errorf("landingpage_id e obrigatorio")
	}
	if !input.Consentimento {
		return nil, domain.ErrConsentimentoObrigatorio
	}
	if len(input.Mensagem) > limiteMensagem {
		return nil, domain.ErrPayloadGrande
	}
	if len(input.CamposPersonalizados) > limiteCamposPersonalizado {
		return nil, domain.ErrPayloadGrande
	}

	lp, err := u.landingPageRepository.BuscarPorID(ctx, input.LandingPageId)
	if err != nil {
		if errors.Is(err, landingpage.ErrNaoEncontrada) {
			return nil, landingpage.ErrNaoEncontrada
		}
		return nil, fmt.Errorf("capturar lead validar landingpage: %w", err)
	}
	if lp.Status != landingpage.StatusPublicada {
		return nil, domain.ErrLandingPageNaoPublicada
	}

	agora := time.Now().UTC()

	if duplicado, err := u.repository.BuscarDuplicado(ctx, input.Email, input.LandingPageId, agora.Add(-janelaDedupe)); err == nil && duplicado != nil {
		return mapearLeadOutput(duplicado), nil
	} else if err != nil && !errors.Is(err, domain.ErrNaoEncontrado) {
		log.Printf("dedupe lead falhou, prosseguindo: %v", err)
	}

	utm := utmDeInput(input.UTM)
	consentimentoEm := agora

	l := &domain.Lead{
		LandingPageId:        input.LandingPageId,
		NegocioId:            lp.NegocioId,
		Nome:                 input.Nome,
		Email:                input.Email,
		Telefone:             input.Telefone,
		Whatsapp:             input.Whatsapp,
		Mensagem:             strings.TrimSpace(input.Mensagem),
		CamposPersonalizados: input.CamposPersonalizados,
		Origem:               "buildnix",
		IP:                   ip,
		UserAgent:            userAgent,
		Referrer:             strings.TrimSpace(input.Referrer),
		UTM:                  utm,
		Consentimento:                   true,
		ConsentimentoEm:                 &consentimentoEm,
		PoliticaVersao:                  strings.TrimSpace(input.PoliticaVersao),
		ConsentimentoMarketingParceiros: input.ConsentimentoMarketingParceiros,
		ConsentimentoVersaoMarketing:    strings.TrimSpace(input.ConsentimentoVersaoMarketing),
		Status:                          domain.StatusNovo,
		CriadoEm:             agora,
		AtualizadoEm:         agora,
	}

	if err := u.repository.Salvar(ctx, l); err != nil {
		return nil, fmt.Errorf("capturar lead salvar: %w", err)
	}

	u.publicarEvento(ctx, l)

	return mapearLeadOutput(l), nil
}

func (u *UseCase) publicarEvento(ctx context.Context, l *domain.Lead) {
	payload := map[string]interface{}{
		"origem":         l.Origem,
		"lead_id":        l.Id,
		"origem_id":      l.LandingPageId,
		"landingpage_id": l.LandingPageId,
		"negocio_id":     l.NegocioId,
		"nome":           l.Nome,
		"email":          l.Email,
		"telefone":       l.Telefone,
		"whatsapp":       l.Whatsapp,
		"mensagem":       l.Mensagem,
		"consentimento":  l.Consentimento,
		"criado_em":      l.CriadoEm,
	}
	if l.PoliticaVersao != "" {
		payload["politica_versao"] = l.PoliticaVersao
	}
	if l.Referrer != "" {
		payload["referrer"] = l.Referrer
	}
	if !l.UTM.Vazio() {
		utm := map[string]string{}
		if l.UTM.Source != "" {
			utm["utm_source"] = l.UTM.Source
		}
		if l.UTM.Medium != "" {
			utm["utm_medium"] = l.UTM.Medium
		}
		if l.UTM.Campaign != "" {
			utm["utm_campaign"] = l.UTM.Campaign
		}
		if l.UTM.Content != "" {
			utm["utm_content"] = l.UTM.Content
		}
		if l.UTM.Term != "" {
			utm["utm_term"] = l.UTM.Term
		}
		payload["utm"] = utm
	}
	if len(l.CamposPersonalizados) > 0 {
		payload["campos_personalizados"] = l.CamposPersonalizados
	}

	if err := u.publisher.Publicar(ctx, evento.Evento{
		EventId:        uuid.New().String(),
		EventType:      "lead.capturado",
		SchemaVersion:  1,
		Source:         "nous",
		OccurredAt:     l.CriadoEm,
		IdempotencyKey: "nous:lead:" + l.Id,
		Payload:        payload,
	}); err != nil {
		log.Printf("publicar evento lead.capturado: %v", err)
	}
}

func emailValido(email string) bool {
	endereco, err := mail.ParseAddress(email)
	return err == nil && endereco.Address == email
}

func normalizarTelefone(valor string) string {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(valor))
	if strings.HasPrefix(valor, "+") {
		b.WriteByte('+')
	}
	for _, r := range valor {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func utmDeInput(in *dto.UTMInput) domain.UTM {
	if in == nil {
		return domain.UTM{}
	}
	return domain.UTM{
		Source:   strings.TrimSpace(in.Source),
		Medium:   strings.TrimSpace(in.Medium),
		Campaign: strings.TrimSpace(in.Campaign),
		Content:  strings.TrimSpace(in.Content),
		Term:     strings.TrimSpace(in.Term),
	}
}

func (u *UseCase) ObterLead(ctx context.Context, id string) (*dto.LeadOutput, error) {
	l, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter lead: %w", err)
	}
	return mapearLeadOutput(l), nil
}

func (u *UseCase) AtualizarStatusLead(ctx context.Context, id string, input dto.AtualizarStatusLeadInput) error {
	status := domain.Status(strings.TrimSpace(input.Status))
	if !domain.StatusValido(status) {
		return domain.ErrStatusInvalido
	}
	motivo := strings.TrimSpace(input.Motivo)
	if status == domain.StatusPerdido && motivo == "" {
		return domain.ErrMotivoObrigatorio
	}
	if status != domain.StatusPerdido {
		motivo = ""
	}
	if err := u.repository.AtualizarStatus(ctx, id, domain.AtualizacaoStatus{Status: status, Motivo: motivo}, time.Now().UTC()); err != nil {
		return fmt.Errorf("atualizar status lead: %w", err)
	}
	return nil
}

func (u *UseCase) AnonimizarLead(ctx context.Context, id string) error {
	if err := u.repository.Anonimizar(ctx, id, time.Now().UTC()); err != nil {
		return fmt.Errorf("anonimizar lead: %w", err)
	}
	return nil
}

func (u *UseCase) AnonimizarExpirados(ctx context.Context, retencao time.Duration) (int, error) {
	if retencao <= 0 {
		retencao = retencaoLeadPadrao
	}
	antesDe := time.Now().UTC().Add(-retencao)
	total, err := u.repository.AnonimizarExpirados(ctx, antesDe)
	if err != nil {
		return 0, fmt.Errorf("anonimizar expirados: %w", err)
	}
	return total, nil
}

func (u *UseCase) ListarLeads(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaLeadsOutput, error) {
	leads, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar leads: %w", err)
	}

	saida := make([]*dto.LeadOutput, 0, len(leads))
	for _, l := range leads {
		saida = append(saida, mapearLeadOutput(l))
	}

	return &dto.ListaLeadsOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func mapearLeadOutput(l *domain.Lead) *dto.LeadOutput {
	out := &dto.LeadOutput{
		Id:                   l.Id,
		LandingPageId:        l.LandingPageId,
		NegocioId:            l.NegocioId,
		Nome:                 l.Nome,
		Email:                l.Email,
		Telefone:             l.Telefone,
		Whatsapp:             l.Whatsapp,
		Mensagem:             l.Mensagem,
		CamposPersonalizados: l.CamposPersonalizados,
		Origem:               l.Origem,
		Status:               string(l.Status),
		PerdidoMotivo:        l.PerdidoMotivo,
		Referrer:             l.Referrer,
		Consentimento:                   l.Consentimento,
		ConsentimentoEm:                 l.ConsentimentoEm,
		PoliticaVersao:                  l.PoliticaVersao,
		ConsentimentoMarketingParceiros: l.ConsentimentoMarketingParceiros,
		ConsentimentoVersaoMarketing:    l.ConsentimentoVersaoMarketing,
		AnonimizadoEm:                   l.AnonimizadoEm,
		CriadoEm:             l.CriadoEm,
	}
	if !l.UTM.Vazio() {
		out.UTM = &dto.UTMInput{
			Source:   l.UTM.Source,
			Medium:   l.UTM.Medium,
			Campaign: l.UTM.Campaign,
			Content:  l.UTM.Content,
			Term:     l.UTM.Term,
		}
	}
	return out
}
