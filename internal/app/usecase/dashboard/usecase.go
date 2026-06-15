package dashboard

import (
	"context"
	"fmt"

	"nous/internal/app/dto"
	domain "nous/internal/domain/dashboard"
	"nous/internal/domain/negocio"
)

type UseCase struct {
	repository        domain.Repository
	negocioRepository negocio.Repository
}

func NewUseCase(repository domain.Repository, negocioRepository negocio.Repository) *UseCase {
	return &UseCase{repository: repository, negocioRepository: negocioRepository}
}

func (u *UseCase) ObterDashboard(ctx context.Context, filtro domain.FiltroPeriodo, usuarioId, role string) (any, error) {
	if role == "admin" {
		return u.ObterDashboardAdmin(ctx, filtro, role)
	}
	filtro.UsuarioId = ""
	return u.ObterDashboardUsuario(ctx, usuarioId, filtro)
}

func (u *UseCase) ObterDashboardUsuario(ctx context.Context, usuarioId string, filtro domain.FiltroPeriodo) (*dto.DashboardUsuarioOutput, error) {
	resumo, err := u.repository.ObterResumoUsuario(ctx, usuarioId, filtro)
	if err != nil {
		return nil, fmt.Errorf("obter dashboard usuario: %w", err)
	}
	return mapearDashboardUsuarioOutput(resumo), nil
}

func (u *UseCase) ObterDashboardNegocio(ctx context.Context, negocioId string, filtro domain.FiltroPeriodo, usuarioId, role string) (*dto.DashboardNegocioOutput, error) {
	if negocioId == "" {
		return nil, domain.ErrNegocioObrigatorio
	}
	if role != "admin" {
		if err := u.negocioRepository.VerificarAcessoUsuario(ctx, negocioId, usuarioId); err != nil {
			return nil, fmt.Errorf("dashboard acesso negocio: %w", err)
		}
	}

	resumo, err := u.repository.ObterResumoNegocio(ctx, negocioId, filtro)
	if err != nil {
		return nil, fmt.Errorf("obter dashboard negocio: %w", err)
	}

	return mapearDashboardNegocioOutput(resumo), nil
}

func (u *UseCase) ObterDashboardAdmin(ctx context.Context, filtro domain.FiltroPeriodo, role string) (*dto.DashboardAdminOutput, error) {
	if role != "admin" {
		return nil, negocio.ErrNaoAutorizado
	}
	resumo, err := u.repository.ObterResumoAdmin(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("obter dashboard admin: %w", err)
	}
	return &dto.DashboardAdminOutput{
		NegociosAtivos:         resumo.NegociosAtivos,
		UsuariosAtivos:         resumo.UsuariosAtivos,
		TemplatesSognix:        resumo.TemplatesSognix,
		TemplatesNegocios:      resumo.TemplatesNegocios,
		LandingPagesCriadas:    resumo.LandingPagesCriadas,
		LandingPagesOnline:     resumo.LandingPagesOnline,
		LeadsCapturados:        resumo.LeadsCapturados,
		ExportacoesGeradas:     resumo.ExportacoesGeradas,
		NegociosComPlanoCustom: resumo.NegociosComPlanoCustom,
	}, nil
}

func mapearDashboardNegocioOutput(resumo *domain.ResumoNegocio) *dto.DashboardNegocioOutput {
	output := &dto.DashboardNegocioOutput{
		NegocioId:           resumo.NegocioId,
		UsuariosVinculados:  resumo.UsuariosVinculados,
		LandingPagesCriadas: resumo.LandingPagesCriadas,
		LandingPagesOnline:  resumo.LandingPagesOnline,
		ExportacoesGeradas:  resumo.ExportacoesGeradas,
	}
	if resumo.LandingPageMaisUsada != nil {
		output.LandingPageMaisUsada = &dto.LandingPageUtilizadaOutput{
			Id:               resumo.LandingPageMaisUsada.Id,
			Nome:             resumo.LandingPageMaisUsada.Nome,
			TotalExportacoes: resumo.LandingPageMaisUsada.TotalExportacoes,
		}
	}
	return output
}

func mapearDashboardUsuarioOutput(resumo *domain.ResumoUsuario) *dto.DashboardUsuarioOutput {
	output := &dto.DashboardUsuarioOutput{
		UsuarioId:           resumo.UsuarioId,
		NegociosVinculados:  resumo.NegociosVinculados,
		LandingPagesCriadas: resumo.LandingPagesCriadas,
		LandingPagesOnline:  resumo.LandingPagesOnline,
		ExportacoesGeradas:  resumo.ExportacoesGeradas,
	}
	if resumo.LandingPageMaisUsada != nil {
		output.LandingPageMaisUsada = &dto.LandingPageUtilizadaOutput{
			Id:               resumo.LandingPageMaisUsada.Id,
			Nome:             resumo.LandingPageMaisUsada.Nome,
			TotalExportacoes: resumo.LandingPageMaisUsada.TotalExportacoes,
		}
	}
	return output
}
