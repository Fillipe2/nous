package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	uccliente "buildnix-backend/internal/app/usecase/cliente"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type ClienteHandler struct {
	useCase *uccliente.UseCase
}

func NewClienteHandler(useCase *uccliente.UseCase) *ClienteHandler {
	return &ClienteHandler{useCase: useCase}
}

func (h *ClienteHandler) CriarCliente(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "usuario nao autenticado")
		return
	}

	var input dto.CriarClienteInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarCliente(r.Context(), input, claims.UsuarioId)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *ClienteHandler) ListarClientes(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.useCase.ListarClientes(r.Context(), cliente.FiltroListagem{
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

func (h *ClienteHandler) ListarClientesDoUsuario(w http.ResponseWriter, r *http.Request) {
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
	output, err := h.useCase.ListarClientesDoUsuario(r.Context(), cliente.FiltroListagem{
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

func (h *ClienteHandler) AtualizarCliente(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarClienteInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarCliente(r.Context(), id, input)
	if errors.Is(err, cliente.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "cliente nao encontrado")
		return
	}
	if errors.Is(err, cliente.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "cliente foi alterado por outro usuario, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *ClienteHandler) ExcluirCliente(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.useCase.ExcluirCliente(r.Context(), id); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ClienteHandler) VincularUsuario(w http.ResponseWriter, r *http.Request) {
	var input dto.VincularUsuarioClienteInput
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
