package handler

import (
	"errors"
	"net/http"

	"nous/internal/app/dto"
	ucunidade "nous/internal/app/usecase/unidade"
	"nous/internal/domain/negocio"
	"nous/internal/domain/unidade"
	"nous/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type UnidadeHandler struct {
	useCase *ucunidade.UseCase
}

func NewUnidadeHandler(useCase *ucunidade.UseCase) *UnidadeHandler {
	return &UnidadeHandler{useCase: useCase}
}

func (h *UnidadeHandler) CriarUnidade(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	var input dto.CriarUnidadeInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarUnidade(r.Context(), input, claims.UsuarioId, claims.Role)
	if errors.Is(err, negocio.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao negocio")
		return
	}
	if errors.Is(err, unidade.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "unidade foi alterado por outro usuario, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *UnidadeHandler) ListarUnidades(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.useCase.ListarUnidades(r.Context(), unidade.FiltroListagem{
		NegocioId:  r.URL.Query().Get("negocio_id"),
		Nome:       r.URL.Query().Get("nome"),
		Status:     r.URL.Query().Get("status"),
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, negocio.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao negocio")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, output)
}

func (h *UnidadeHandler) AtualizarUnidade(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	var input dto.AtualizarUnidadeInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarUnidade(r.Context(), id, input, claims.UsuarioId, claims.Role)
	if errors.Is(err, unidade.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "unidade nao encontrado")
		return
	}
	if errors.Is(err, negocio.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado ao negocio")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *UnidadeHandler) ExcluirUnidade(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	if err := h.useCase.ExcluirUnidade(r.Context(), id, claims.UsuarioId, claims.Role); err != nil {
		if errors.Is(err, negocio.ErrNaoAutorizado) {
			responderErro(w, http.StatusForbidden, "acesso negado ao negocio")
			return
		}
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
