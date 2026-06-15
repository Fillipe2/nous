package handler

import (
	"errors"
	"net/http"

	"nous/internal/app/dto"
	ucnegocio "nous/internal/app/usecase/negocio"
	"nous/internal/domain/negocio"
	"nous/internal/interface/webapi/internal/middleware"

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
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	var input dto.CriarNegocioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarNegocio(r.Context(), input, claims.UsuarioId)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *NegocioHandler) ListarNegocios(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.useCase.ListarNegocios(r.Context(), negocio.FiltroListagem{
		UsuarioId:  r.URL.Query().Get("usuario_id"),
		Nome:       r.URL.Query().Get("nome"),
		Email:      r.URL.Query().Get("email"),
		Status:     r.URL.Query().Get("status"),
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, output)
}

func (h *NegocioHandler) ListarNegociosDoUsuario(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.useCase.ListarNegociosDoUsuario(r.Context(), negocio.FiltroListagem{
		UsuarioId:  claims.UsuarioId,
		Nome:       r.URL.Query().Get("nome"),
		Email:      r.URL.Query().Get("email"),
		Status:     r.URL.Query().Get("status"),
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, output)
}

func (h *NegocioHandler) AtualizarNegocio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarNegocioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarNegocio(r.Context(), id, input)
	if errors.Is(err, negocio.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "negocio nao encontrado")
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

	responderJSON(w, http.StatusOK, output)
}

func (h *NegocioHandler) ExcluirNegocio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.useCase.ExcluirNegocio(r.Context(), id); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NegocioHandler) VincularUsuario(w http.ResponseWriter, r *http.Request) {
	var input dto.VincularUsuarioNegocioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	if err := h.useCase.VincularUsuario(r.Context(), input); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "usuario vinculado com sucesso"})
}
