package lead

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/lead"
)

type UseCase struct {
	repository domain.Repository
	publisher  evento.Publisher
}

func NewUseCase(repository domain.Repository, publisher evento.Publisher) *UseCase {
	return &UseCase{repository: repository, publisher: publisher}
}

func (u *UseCase) CapturarLead(ctx context.Context, input dto.CapturarLeadInput, ip, userAgent string) (*dto.LeadOutput, error) {
	if input.Nome == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" || !emailValido(input.Email) {
		return nil, fmt.Errorf("email valido e obrigatorio")
	}
	if input.Telefone == "" && input.Whatsapp == "" {
		return nil, fmt.Errorf("telefone ou whatsapp e obrigatorio")
	}
	if input.LandingPageId == "" {
		return nil, fmt.Errorf("landingpage_id e obrigatorio")
	}

	l := &domain.Lead{
		LandingPageId:        input.LandingPageId,
		Nome:                 input.Nome,
		Email:                input.Email,
		Telefone:             input.Telefone,
		Whatsapp:             input.Whatsapp,
		Mensagem:             input.Mensagem,
		CamposPersonalizados: input.CamposPersonalizados,
		Origem:               "buildnix",
		IP:                   ip,
		UserAgent:            userAgent,
		Status:               domain.StatusNovo,
		CriadoEm:             time.Now(),
	}

	if err := u.repository.Salvar(ctx, l); err != nil {
		return nil, fmt.Errorf("capturar lead salvar: %w", err)
	}

	if err := u.publisher.Publicar(ctx, evento.Evento{
		EventId:        uuid.New().String(),
		EventType:      "lead.capturado",
		SchemaVersion:  1,
		Source:         "buildnix",
		OccurredAt:     l.CriadoEm,
		IdempotencyKey: "buildnix:lead:" + l.Id,
		Payload: map[string]interface{}{
			"origem":         "buildnix",
			"lead_id":        l.Id,
			"origem_id":      l.LandingPageId,
			"landingpage_id": l.LandingPageId,
			"nome":           l.Nome,
			"email":          l.Email,
			"telefone":       l.Telefone,
			"whatsapp":       l.Whatsapp,
			"mensagem":       l.Mensagem,
			"criado_em":      l.CriadoEm,
		},
	}); err != nil {
		log.Printf("publicar evento lead.capturado: %v", err)
	}

	return mapearLeadOutput(l), nil
}

func emailValido(email string) bool {
	endereco, err := mail.ParseAddress(email)
	return err == nil && endereco.Address == email
}

func (u *UseCase) ObterLead(ctx context.Context, id string) (*dto.LeadOutput, error) {
	l, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter lead: %w", err)
	}
	return mapearLeadOutput(l), nil
}

func (u *UseCase) AtualizarStatusLead(ctx context.Context, id string, input dto.AtualizarStatusLeadInput) error {
	status := domain.Status(input.Status)
	if err := u.repository.AtualizarStatus(ctx, id, status); err != nil {
		return fmt.Errorf("atualizar status lead: %w", err)
	}
	return nil
}

func (u *UseCase) ListarLeads(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaLeadsOutput, error) {
	leads, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar leads: %w", err)
	}

	var saida []*dto.LeadOutput
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
	return &dto.LeadOutput{
		Id:                   l.Id,
		LandingPageId:        l.LandingPageId,
		Nome:                 l.Nome,
		Email:                l.Email,
		Telefone:             l.Telefone,
		Whatsapp:             l.Whatsapp,
		Mensagem:             l.Mensagem,
		CamposPersonalizados: l.CamposPersonalizados,
		Origem:               l.Origem,
		Status:               string(l.Status),
		CriadoEm:             l.CriadoEm,
	}
}
