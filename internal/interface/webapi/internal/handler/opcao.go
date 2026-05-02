package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	uccategoria "buildnix-backend/internal/app/usecase/categoria"
	uccliente "buildnix-backend/internal/app/usecase/cliente"
	uclandingpage "buildnix-backend/internal/app/usecase/landingpage"
	ucnegocio "buildnix-backend/internal/app/usecase/negocio"
	uctemplate "buildnix-backend/internal/app/usecase/template"
	ucusuario "buildnix-backend/internal/app/usecase/usuario"
	"buildnix-backend/internal/domain/categoria"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/domain/negocio"
	"buildnix-backend/internal/domain/template"
	"buildnix-backend/internal/domain/usuario"
	"buildnix-backend/internal/interface/webapi/internal/middleware"
)

type OpcaoHandler struct {
	cliente     *uccliente.UseCase
	usuario     *ucusuario.UseCase
	negocio     *ucnegocio.UseCase
	categoria   *uccategoria.UseCase
	template    *uctemplate.UseCase
	landingPage *uclandingpage.UseCase
}

func NewOpcaoHandler(cliente *uccliente.UseCase, usuario *ucusuario.UseCase, negocio *ucnegocio.UseCase, categoria *uccategoria.UseCase, template *uctemplate.UseCase, landingPage *uclandingpage.UseCase) *OpcaoHandler {
	return &OpcaoHandler{cliente: cliente, usuario: usuario, negocio: negocio, categoria: categoria, template: template, landingPage: landingPage}
}

func (h *OpcaoHandler) ListarClientes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacaoOpcoes(r)
	filtro := cliente.FiltroListagem{
		UsuarioId:  r.URL.Query().Get("usuario_id"),
		Nome:       r.URL.Query().Get("q"),
		Status:     r.URL.Query().Get("status"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: "nome",
		Ordem:      "asc",
	}

	if claims.Role != "admin" {
		filtro.UsuarioId = claims.UsuarioId
		output, err := h.cliente.ListarClientesDoUsuario(r.Context(), filtro)
		if err != nil {
			responderErro(w, http.StatusInternalServerError, "erro interno")
			return
		}
		responderJSON(w, http.StatusOK, opcoesClientesBasicos(output))
		return
	}

	output, err := h.cliente.ListarClientes(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesClientes(output))
}

func (h *OpcaoHandler) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacaoOpcoes(r)
	output, err := h.usuario.ListarUsuarios(r.Context(), usuario.FiltroListagem{
		ClienteId:  r.URL.Query().Get("cliente_id"),
		Nome:       r.URL.Query().Get("q"),
		Email:      r.URL.Query().Get("email"),
		Role:       r.URL.Query().Get("role"),
		Status:     r.URL.Query().Get("status"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: "nome",
		Ordem:      "asc",
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesUsuarios(output))
}

func (h *OpcaoHandler) ListarNegocios(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacaoOpcoes(r)
	output, err := h.negocio.ListarNegocios(r.Context(), negocio.FiltroListagem{
		ClienteId:  r.URL.Query().Get("cliente_id"),
		Nome:       r.URL.Query().Get("q"),
		Status:     r.URL.Query().Get("status"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: "nome",
		Ordem:      "asc",
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesNegocios(output))
}

func (h *OpcaoHandler) ListarCategorias(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacaoOpcoes(r)
	output, err := h.categoria.ListarCategorias(r.Context(), categoria.FiltroListagem{
		Nome:       r.URL.Query().Get("q"),
		Status:     r.URL.Query().Get("status"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: "nome",
		Ordem:      "asc",
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesCategorias(output))
}

func (h *OpcaoHandler) ListarTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacaoOpcoes(r)
	output, err := h.template.ListarTemplates(r.Context(), template.FiltroListagem{
		ClienteId:   r.URL.Query().Get("cliente_id"),
		CategoriaId: r.URL.Query().Get("categoria_id"),
		Tipo:        r.URL.Query().Get("tipo"),
		Origem:      r.URL.Query().Get("origem"),
		Nome:        r.URL.Query().Get("q"),
		Status:      r.URL.Query().Get("status"),
		Pagina:      pagina,
		PorPagina:   porPagina,
		OrdenarPor:  "nome",
		Ordem:       "asc",
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, template.ErrClienteObrigatorio) {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, template.ErrNaoAutorizado) || errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesTemplates(output))
}

func (h *OpcaoHandler) ListarLandingPages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacaoOpcoes(r)
	output, err := h.landingPage.ListarLandingPages(r.Context(), landingpage.FiltroListagem{
		ClienteId:   r.URL.Query().Get("cliente_id"),
		NegocioId:   r.URL.Query().Get("negocio_id"),
		CategoriaId: r.URL.Query().Get("categoria_id"),
		TemplateId:  r.URL.Query().Get("template_id"),
		Nome:        r.URL.Query().Get("q"),
		Tipo:        r.URL.Query().Get("tipo"),
		Status:      r.URL.Query().Get("status"),
		Pagina:      pagina,
		PorPagina:   porPagina,
		OrdenarPor:  "nome",
		Ordem:       "asc",
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrNaoAutorizado) || errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, opcoesLandingPages(output))
}

func (h *OpcaoHandler) ListarEnums(w http.ResponseWriter, _ *http.Request) {
	responderJSON(w, http.StatusOK, map[string][]*dto.OpcaoOutput{
		"landingpage_tipos": {
			opcaoEstatica("white_label", "White Label"),
			opcaoEstatica("customizavel", "Customizavel"),
		},
		"landingpage_status": {
			opcaoEstatica("rascunho", "Rascunho"),
			opcaoEstatica("publicada", "Publicada"),
			opcaoEstatica("arquivada", "Arquivada"),
		},
		"template_tipos": {
			opcaoEstatica("white_label", "White Label"),
			opcaoEstatica("base_customizavel", "Base customizavel"),
		},
		"template_origens": {
			opcaoEstatica("sognix", "Sognix"),
			opcaoEstatica("cliente", "Cliente"),
		},
		"lead_status": {
			opcaoEstatica("novo", "Novo"),
			opcaoEstatica("contatado", "Contatado"),
			opcaoEstatica("convertido", "Convertido"),
			opcaoEstatica("perdido", "Perdido"),
		},
	})
}

func paginacaoOpcoes(r *http.Request) (pagina, porPagina int) {
	pagina = 1
	porPagina = 100
	if p := r.URL.Query().Get("pagina"); p != "" {
		if v := atoiSafe(p); v > 0 {
			pagina = v
		}
	}
	if pp := r.URL.Query().Get("por_pagina"); pp != "" {
		if v := atoiSafe(pp); v > 0 && v <= 100 {
			porPagina = v
		}
	}
	return pagina, porPagina
}

func opcaoEstatica(id, nome string) *dto.OpcaoOutput {
	return &dto.OpcaoOutput{Id: id, Nome: nome, Label: nome}
}

func opcoesClientes(output *dto.ListaClientesOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesClientesBasicos(output *dto.ListaClientesBasicosOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesUsuarios(output *dto.ListaUsuariosOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status, Tipo: item.Role})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesNegocios(output *dto.ListaNegociosOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status, ClienteId: item.ClienteId})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesCategorias(output *dto.ListaCategoriasOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesTemplates(output *dto.ListaTemplatesOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status, Tipo: item.Tipo, Origem: item.Origem, ClienteId: item.ClienteId, CategoriaId: item.CategoriaId})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}

func opcoesLandingPages(output *dto.ListaLandingPagesOutput) *dto.ListaOpcoesOutput {
	dados := make([]*dto.OpcaoOutput, 0, len(output.Dados))
	for _, item := range output.Dados {
		dados = append(dados, &dto.OpcaoOutput{Id: item.Id, Nome: item.Nome, Label: item.Nome, Status: item.Status, Tipo: item.Tipo, ClienteId: item.ClienteId, CategoriaId: item.CategoriaId})
	}
	return &dto.ListaOpcoesOutput{Dados: dados, Total: output.Total, Pagina: output.Pagina, PorPagina: output.PorPagina}
}
