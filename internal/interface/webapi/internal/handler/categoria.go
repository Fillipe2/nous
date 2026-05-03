package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	uccategoria "buildnix-backend/internal/app/usecase/categoria"
	"buildnix-backend/internal/domain/categoria"

	"github.com/go-chi/chi/v5"
)

type CategoriaHandler struct {
	useCase *uccategoria.UseCase
}

func NewCategoriaHandler(useCase *uccategoria.UseCase) *CategoriaHandler {
	return &CategoriaHandler{useCase: useCase}
}

func (h *CategoriaHandler) CriarCategoria(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarCategoriaInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarCategoria(r.Context(), input)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *CategoriaHandler) ListarCategorias(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.useCase.ListarCategorias(r.Context(), categoria.FiltroListagem{
		Nome:       r.URL.Query().Get("nome"),
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

func (h *CategoriaHandler) AtualizarCategoria(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarCategoriaInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarCategoria(r.Context(), id, input)
	if errors.Is(err, categoria.ErrNaoEncontrada) {
		responderErro(w, http.StatusNotFound, "categoria nao encontrada")
		return
	}
	if errors.Is(err, categoria.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "categoria foi alterada por outro usuario, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *CategoriaHandler) ExcluirCategoria(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.useCase.ExcluirCategoria(r.Context(), id); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
