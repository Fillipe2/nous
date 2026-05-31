package dashboard

import (
	"context"
	"fmt"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	domain "nous/internal/domain/dashboard"
)

type UseCase struct {
	repository        domain.Repository
	clienteRepository cliente.Repository
}

func NewUseCase(repository domain.Repository, clienteRepository cliente.Repository) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository}
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

func (u *UseCase) ObterDashboardCliente(ctx context.Context, clienteId string, filtro domain.FiltroPeriodo, usuarioId, role string) (*dto.DashboardClienteOutput, error) {
	if clienteId == "" {
		return nil, domain.ErrClienteObrigatorio
	}
	if role != "admin" {
		if err := u.clienteRepository.VerificarAcessoUsuario(ctx, clienteId, usuarioId); err != nil {
			return nil, fmt.Errorf("dashboard acesso cliente: %w", err)
		}
	}

	resumo, err := u.repository.ObterResumoCliente(ctx, clienteId, filtro)
	if err != nil {
		return nil, fmt.Errorf("obter dashboard cliente: %w", err)
	}

	return mapearDashboardClienteOutput(resumo), nil
}

func (u *UseCase) ObterDashboardAdmin(ctx context.Context, filtro domain.FiltroPeriodo, role string) (*dto.DashboardAdminOutput, error) {
	if role != "admin" {
		return nil, cliente.ErrNaoAutorizado
	}
	resumo, err := u.repository.ObterResumoAdmin(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("obter dashboard admin: %w", err)
	}
	return &dto.DashboardAdminOutput{
		ClientesAtivos:         resumo.ClientesAtivos,
		UsuariosAtivos:         resumo.UsuariosAtivos,
		TemplatesSognix:        resumo.TemplatesSognix,
		TemplatesClientes:      resumo.TemplatesClientes,
		LandingPagesCriadas:    resumo.LandingPagesCriadas,
		LandingPagesOnline:     resumo.LandingPagesOnline,
		LeadsCapturados:        resumo.LeadsCapturados,
		ExportacoesGeradas:     resumo.ExportacoesGeradas,
		ClientesComPlanoCustom: resumo.ClientesComPlanoCustom,
	}, nil
}

func mapearDashboardClienteOutput(resumo *domain.ResumoCliente) *dto.DashboardClienteOutput {
	output := &dto.DashboardClienteOutput{
		ClienteId:           resumo.ClienteId,
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
		ClientesVinculados:  resumo.ClientesVinculados,
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
