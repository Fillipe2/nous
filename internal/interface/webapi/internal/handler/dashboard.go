package handler

import (
	"errors"
	"net/http"

	ucdashboard "buildnix-backend/internal/app/usecase/dashboard"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/dashboard"
	"buildnix-backend/internal/interface/webapi/internal/middleware"
)

type DashboardHandler struct {
	useCase *ucdashboard.UseCase
}

func NewDashboardHandler(useCase *ucdashboard.UseCase) *DashboardHandler {
	return &DashboardHandler{useCase: useCase}
}

func (h *DashboardHandler) ObterDashboardCliente(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	clienteId := r.URL.Query().Get("cliente_id")
	output, err := h.useCase.ObterDashboardCliente(r.Context(), clienteId, dashboard.FiltroPeriodo{
		DataInicio: r.URL.Query().Get("data_inicio"),
		DataFim:    r.URL.Query().Get("data_fim"),
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

func (h *DashboardHandler) ObterDashboardAdmin(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	output, err := h.useCase.ObterDashboardAdmin(r.Context(), dashboard.FiltroPeriodo{
		DataInicio: r.URL.Query().Get("data_inicio"),
		DataFim:    r.URL.Query().Get("data_fim"),
	}, claims.Role)
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}
