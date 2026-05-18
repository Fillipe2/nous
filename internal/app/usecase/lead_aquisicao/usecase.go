package lead_aquisicao

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/lead_aquisicao"
)

const limiteMensagem = 2000

type UseCase struct {
	repository domain.Repository
	publisher  evento.Publisher
}

func NewUseCase(repository domain.Repository, publisher evento.Publisher) *UseCase {
	return &UseCase{repository: repository, publisher: publisher}
}

func (u *UseCase) RegistrarLeadAquisicao(ctx context.Context, input dto.RegistrarLeadAquisicaoInput, ip, userAgent string) (*dto.LeadAquisicaoOutput, error) {
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
	if len(input.Mensagem) > limiteMensagem {
		return nil, fmt.Errorf("mensagem muito longa")
	}

	porte := domain.Porte(strings.TrimSpace(input.Porte))
	if !domain.PorteValido(porte) {
		return nil, fmt.Errorf("porte invalido: use micro, pequeno, medio ou grande")
	}

	if _, err := u.repository.BuscarPorEmail(ctx, input.Email); err == nil {
		return nil, domain.ErrEmailDuplicado
	} else if !errors.Is(err, domain.ErrNaoEncontrado) {
		return nil, fmt.Errorf("verificar email lead_aquisicao: %w", err)
	}

	agora := time.Now().UTC()
	la := &domain.LeadAquisicao{
		Nome:                     input.Nome,
		Email:                    input.Email,
		Telefone:                 strings.TrimSpace(input.Telefone),
		CategoriaPretendidaId:    strings.TrimSpace(input.CategoriaPretendidaId),
		CategoriaPretendidaLivre: strings.TrimSpace(input.CategoriaPretendidaLivre),
		Cidade:                   strings.TrimSpace(input.Cidade),
		Estado:                   strings.ToUpper(strings.TrimSpace(input.Estado)),
		Porte:                    porte,
		Interesses:               input.Interesses,
		Mensagem:                 strings.TrimSpace(input.Mensagem),
		Origem:                   strings.TrimSpace(input.Origem),
		StatusFunil:              domain.StatusFunilNovo,
		IP:                       ip,
		UserAgent:                userAgent,
		CriadoEm:                 agora,
		AtualizadoEm:             agora,
	}
	if input.UTM != nil {
		la.UTMSource = strings.TrimSpace(input.UTM.Source)
		la.UTMMedium = strings.TrimSpace(input.UTM.Medium)
		la.UTMCampaign = strings.TrimSpace(input.UTM.Campaign)
		la.UTMContent = strings.TrimSpace(input.UTM.Content)
		la.UTMTerm = strings.TrimSpace(input.UTM.Term)
	}

	if err := u.repository.Salvar(ctx, la); err != nil {
		return nil, fmt.Errorf("registrar lead_aquisicao salvar: %w", err)
	}

	_ = u.publisher.Publicar(ctx, evento.Evento{
		EventType:      "lead_aquisicao.registrado",
		Source:         "buildnix",
		SchemaVersion:  1,
		IdempotencyKey: "buildnix:lead_aquisicao:registrado:" + la.Id,
		Payload: map[string]interface{}{
			"lead_aquisicao_id": la.Id,
			"email":             la.Email,
			"porte":             string(la.Porte),
			"origem":            la.Origem,
			"status_funil":      string(la.StatusFunil),
		},
	})

	return mapearOutput(la), nil
}

func (u *UseCase) ObterLeadAquisicao(ctx context.Context, id string) (*dto.LeadAquisicaoOutput, error) {
	la, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter lead_aquisicao: %w", err)
	}
	return mapearOutput(la), nil
}

func (u *UseCase) AtualizarFunilLeadAquisicao(ctx context.Context, id string, input dto.AtualizarFunilLeadAquisicaoInput) error {
	status := domain.StatusFunil(strings.TrimSpace(input.StatusFunil))
	if !domain.StatusFunilValido(status) {
		return domain.ErrStatusInvalido
	}
	atualizacao := domain.AtualizacaoFunil{
		StatusFunil:           status,
		ConvertidoEmUsuarioId: strings.TrimSpace(input.ConvertidoEmUsuarioId),
	}
	if err := u.repository.AtualizarFunil(ctx, id, atualizacao, time.Now().UTC()); err != nil {
		return fmt.Errorf("atualizar funil lead_aquisicao: %w", err)
	}
	return nil
}

func (u *UseCase) ListarLeadsAquisicao(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaLeadsAquisicaoOutput, error) {
	leads, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar leads_aquisicao: %w", err)
	}

	saida := make([]*dto.LeadAquisicaoOutput, 0, len(leads))
	for _, la := range leads {
		saida = append(saida, mapearOutput(la))
	}

	return &dto.ListaLeadsAquisicaoOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func mapearOutput(la *domain.LeadAquisicao) *dto.LeadAquisicaoOutput {
	out := &dto.LeadAquisicaoOutput{
		Id:                       la.Id,
		Nome:                     la.Nome,
		Email:                    la.Email,
		Telefone:                 la.Telefone,
		CategoriaPretendidaId:    la.CategoriaPretendidaId,
		CategoriaPretendidaLivre: la.CategoriaPretendidaLivre,
		Cidade:                   la.Cidade,
		Estado:                   la.Estado,
		Porte:                    string(la.Porte),
		Interesses:               la.Interesses,
		Mensagem:                 la.Mensagem,
		Origem:                   la.Origem,
		StatusFunil:              string(la.StatusFunil),
		ConvertidoEmUsuarioId:    la.ConvertidoEmUsuarioId,
		CriadoEm:                 la.CriadoEm,
		AtualizadoEm:             la.AtualizadoEm,
	}
	if la.UTMSource != "" || la.UTMMedium != "" || la.UTMCampaign != "" || la.UTMContent != "" || la.UTMTerm != "" {
		out.UTM = &dto.UTMInput{
			Source:   la.UTMSource,
			Medium:   la.UTMMedium,
			Campaign: la.UTMCampaign,
			Content:  la.UTMContent,
			Term:     la.UTMTerm,
		}
	}
	return out
}

func emailValido(email string) bool {
	endereco, err := mail.ParseAddress(email)
	return err == nil && endereco.Address == email
}
