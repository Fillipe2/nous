package prospecto

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"buildnix-backend/internal/app/dto"
	domain "buildnix-backend/internal/domain/prospecto"
)

const limiteMensagemProspecto = 2000

type UseCase struct {
	repository domain.Repository
}

func NewUseCase(repository domain.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) RegistrarProspecto(ctx context.Context, input dto.RegistrarProspectoInput, ip, userAgent string) (*dto.ProspectoOutput, error) {
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
	if len(input.Mensagem) > limiteMensagemProspecto {
		return nil, fmt.Errorf("mensagem muito longa")
	}

	porte := domain.Porte(strings.TrimSpace(input.Porte))
	if !domain.PorteValido(porte) {
		return nil, fmt.Errorf("porte invalido: use micro, pequeno, medio ou grande")
	}

	if _, err := u.repository.BuscarPorEmail(ctx, input.Email); err == nil {
		return nil, domain.ErrEmailDuplicado
	} else if !errors.Is(err, domain.ErrNaoEncontrado) {
		return nil, fmt.Errorf("verificar email prospecto: %w", err)
	}

	agora := time.Now().UTC()
	p := &domain.Prospecto{
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
		p.UTMSource = strings.TrimSpace(input.UTM.Source)
		p.UTMMedium = strings.TrimSpace(input.UTM.Medium)
		p.UTMCampaign = strings.TrimSpace(input.UTM.Campaign)
		p.UTMContent = strings.TrimSpace(input.UTM.Content)
		p.UTMTerm = strings.TrimSpace(input.UTM.Term)
	}

	if err := u.repository.Salvar(ctx, p); err != nil {
		return nil, fmt.Errorf("registrar prospecto salvar: %w", err)
	}

	return mapearProspectoOutput(p), nil
}

func (u *UseCase) ObterProspecto(ctx context.Context, id string) (*dto.ProspectoOutput, error) {
	p, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter prospecto: %w", err)
	}
	return mapearProspectoOutput(p), nil
}

func (u *UseCase) AtualizarFunilProspecto(ctx context.Context, id string, input dto.AtualizarFunilProspectoInput) error {
	status := domain.StatusFunil(strings.TrimSpace(input.StatusFunil))
	if !domain.StatusFunilValido(status) {
		return domain.ErrStatusInvalido
	}
	atualizacao := domain.AtualizacaoFunil{
		StatusFunil:           status,
		ConvertidoEmUsuarioId: strings.TrimSpace(input.ConvertidoEmUsuarioId),
	}
	if err := u.repository.AtualizarFunil(ctx, id, atualizacao, time.Now().UTC()); err != nil {
		return fmt.Errorf("atualizar funil prospecto: %w", err)
	}
	return nil
}

func (u *UseCase) ListarProspectos(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaProspectosOutput, error) {
	prospectos, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar prospectos: %w", err)
	}

	saida := make([]*dto.ProspectoOutput, 0, len(prospectos))
	for _, p := range prospectos {
		saida = append(saida, mapearProspectoOutput(p))
	}

	return &dto.ListaProspectosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func mapearProspectoOutput(p *domain.Prospecto) *dto.ProspectoOutput {
	out := &dto.ProspectoOutput{
		Id:                       p.Id,
		Nome:                     p.Nome,
		Email:                    p.Email,
		Telefone:                 p.Telefone,
		CategoriaPretendidaId:    p.CategoriaPretendidaId,
		CategoriaPretendidaLivre: p.CategoriaPretendidaLivre,
		Cidade:                   p.Cidade,
		Estado:                   p.Estado,
		Porte:                    string(p.Porte),
		Interesses:               p.Interesses,
		Mensagem:                 p.Mensagem,
		Origem:                   p.Origem,
		StatusFunil:              string(p.StatusFunil),
		ConvertidoEmUsuarioId:    p.ConvertidoEmUsuarioId,
		CriadoEm:                 p.CriadoEm,
		AtualizadoEm:             p.AtualizadoEm,
	}
	if p.UTMSource != "" || p.UTMMedium != "" || p.UTMCampaign != "" || p.UTMContent != "" || p.UTMTerm != "" {
		out.UTM = &dto.UTMInput{
			Source:   p.UTMSource,
			Medium:   p.UTMMedium,
			Campaign: p.UTMCampaign,
			Content:  p.UTMContent,
			Term:     p.UTMTerm,
		}
	}
	return out
}

func emailValido(email string) bool {
	endereco, err := mail.ParseAddress(email)
	return err == nil && endereco.Address == email
}
