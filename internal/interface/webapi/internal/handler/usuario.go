package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	ucusuario "buildnix-backend/internal/app/usecase/usuario"
	"buildnix-backend/internal/domain/usuario"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type UsuarioHandler struct {
	useCase *ucusuario.UseCase
}

func NewUsuarioHandler(useCase *ucusuario.UseCase) *UsuarioHandler {
	return &UsuarioHandler{useCase: useCase}
}

func (h *UsuarioHandler) CriarUsuario(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarUsuarioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarUsuario(r.Context(), input)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *UsuarioHandler) ObterUsuarioAutenticado(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	output, err := h.useCase.ObterUsuario(r.Context(), claims.UsuarioId)
	if errors.Is(err, usuario.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "usuario nao encontrado")
		return
	}
	if errors.Is(err, usuario.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "usuario foi alterado em outra sessao, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *UsuarioHandler) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	output, err := h.useCase.ListarUsuarios(r.Context(), usuario.FiltroListagem{
		Id:         r.URL.Query().Get("id"),
		ClienteId:  r.URL.Query().Get("cliente_id"),
		Nome:       r.URL.Query().Get("nome"),
		Email:      r.URL.Query().Get("email"),
		Role:       r.URL.Query().Get("role"),
		Status:     r.URL.Query().Get("status"),
		DataInicio: r.URL.Query().Get("data_inicio"),
		DataFim:    r.URL.Query().Get("data_fim"),
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

func (h *UsuarioHandler) AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	if claims == nil || (claims.Role != "admin" && claims.UsuarioId != id) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	var input dto.AtualizarUsuarioInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarUsuario(r.Context(), id, input)
	if errors.Is(err, usuario.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "usuario nao encontrado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *UsuarioHandler) ExcluirUsuario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	if claims == nil || (claims.Role != "admin" && claims.UsuarioId != id) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err := h.useCase.ExcluirUsuario(r.Context(), id); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
