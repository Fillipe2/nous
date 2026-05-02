// Package input define os payloads HTTP de entrada da API REST.
package input

import "buildnix-backend/internal/app/dto"

// Aliases mantem a camada de interface com contratos explicitamente nomeados sem duplicar DTOs da aplicação.
type (
	Login                  = dto.LoginInput
	RefreshToken           = dto.RefreshTokenInput
	CriarUsuario           = dto.CriarUsuarioInput
	AtualizarUsuario       = dto.AtualizarUsuarioInput
	CriarCliente           = dto.CriarClienteInput
	AtualizarCliente       = dto.AtualizarClienteInput
	VincularUsuarioCliente = dto.VincularUsuarioClienteInput
	CriarNegocio           = dto.CriarNegocioInput
	AtualizarNegocio       = dto.AtualizarNegocioInput
	CriarCategoria         = dto.CriarCategoriaInput
	AtualizarCategoria     = dto.AtualizarCategoriaInput
	CriarLandingPage       = dto.CriarLandingPageInput
	AtualizarLandingPage   = dto.AtualizarLandingPageInput
	AutosaveLandingPage    = dto.AutosaveInput
	CapturarLead           = dto.CapturarLeadInput
	AtualizarStatusLead    = dto.AtualizarStatusLeadInput
)
