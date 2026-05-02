package dashboard

import (
	"context"
	"fmt"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	domain "buildnix-backend/internal/domain/dashboard"
)

type UseCase struct {
	repository        domain.Repository
	clienteRepository cliente.Repository
}

func NewUseCase(repository domain.Repository, clienteRepository cliente.Repository) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository}
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
