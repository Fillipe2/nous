package handler

import (
	"errors"
	"net/http"

	ucdashboard "nous/internal/app/usecase/dashboard"
	"nous/internal/domain/cliente"
	"nous/internal/domain/dashboard"
	"nous/internal/interface/webapi/internal/middleware"
)

type DashboardHandler struct {
	useCase *ucdashboard.UseCase
}

func NewDashboardHandler(useCase *ucdashboard.UseCase) *DashboardHandler {
	return &DashboardHandler{useCase: useCase}
}

func (h *DashboardHandler) ObterDashboard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.useCase.ObterDashboard(r.Context(), dashboard.FiltroPeriodo{
		DataInicio:    dataInicio,
		DataFim:       dataFim,
		ClienteId:     r.URL.Query().Get("cliente_id"),
		UsuarioId:     r.URL.Query().Get("usuario_id"),
		LandingPageId: r.URL.Query().Get("landingpage_id"),
		NegocioId:     r.URL.Query().Get("negocio_id"),
		CategoriaId:   r.URL.Query().Get("categoria_id"),
		TemplateId:    r.URL.Query().Get("template_id"),
		Tipo:          r.URL.Query().Get("tipo"),
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, dashboard.ErrClienteObrigatorio) {
		responderErro(w, http.StatusBadRequest, "cliente_id e obrigatorio")
		return
	}
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao cliente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}
