package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	ucnegocio "buildnix-backend/internal/app/usecase/negocio"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/negocio"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type NegocioHandler struct {
	useCase *ucnegocio.UseCase
}

func NewNegocioHandler(useCase *ucnegocio.UseCase) *NegocioHandler {
	return &NegocioHandler{useCase: useCase}
}

func (h *NegocioHandler) CriarNegocio(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	var input dto.CriarNegocioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarNegocio(r.Context(), input, claims.UsuarioId, claims.Role)
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao cliente")
		return
	}
	if errors.Is(err, negocio.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "negocio foi alterado por outro usuario, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *NegocioHandler) ListarNegocios(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacao(r)

	output, err := h.useCase.ListarNegocios(r.Context(), negocio.FiltroListagem{
		ClienteId:  r.URL.Query().Get("cliente_id"),
		Nome:       r.URL.Query().Get("nome"),
		Status:     r.URL.Query().Get("status"),
		DataInicio: r.URL.Query().Get("data_inicio"),
		DataFim:    r.URL.Query().Get("data_fim"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao cliente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, output)
}

func (h *NegocioHandler) AtualizarNegocio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	var input dto.AtualizarNegocioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarNegocio(r.Context(), id, input, claims.UsuarioId, claims.Role)
	if errors.Is(err, negocio.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "negocio nao encontrado")
		return
	}
	if errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao cliente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *NegocioHandler) ExcluirNegocio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	if err := h.useCase.ExcluirNegocio(r.Context(), id, claims.UsuarioId, claims.Role); err != nil {
		if errors.Is(err, cliente.ErrNaoAutorizado) {
			responderErro(w, http.StatusForbidden, "acesso negado ao cliente")
			return
		}
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
