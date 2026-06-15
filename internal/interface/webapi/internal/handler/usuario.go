package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"nous/internal/app/dto"
	ucusuario "nous/internal/app/usecase/usuario"
	"nous/internal/domain/usuario"
	"nous/internal/interface/webapi/internal/middleware"

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
	if errors.Is(err, usuario.ErrEmailDuplicado) {
		responderErro(w, http.StatusConflict, "email ja cadastrado")
		return
	}
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
	dataInicio, inicio, err := dataQuery(r, "data_inicio")
	if err != nil {
		responderErro(w, http.StatusBadRequest, "data_inicio deve estar no formato YYYY-MM-DD")
		return
	}
	dataFim, fim, err := dataQuery(r, "data_fim")
	if err != nil {
		responderErro(w, http.StatusBadRequest, "data_fim deve estar no formato YYYY-MM-DD")
		return
	}
	if !inicio.IsZero() && !fim.IsZero() && fim.Before(inicio) {
		responderErro(w, http.StatusBadRequest, "data_fim deve ser maior ou igual a data_inicio")
		return
	}

	output, err := h.useCase.ListarUsuarios(r.Context(), usuario.FiltroListagem{
		Nome:       r.URL.Query().Get("nome"),
		Email:      r.URL.Query().Get("email"),
		Role:       r.URL.Query().Get("role"),
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

func (h *UsuarioHandler) AlterarSenha(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	if claims == nil {
		responderErro(w, http.StatusUnauthorized, "nao autenticado")
		return
	}

	var input dto.AlterarSenhaInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}
	if err := h.useCase.AlterarSenha(r.Context(), claims.UsuarioId, input); err != nil {
		if errors.Is(err, usuario.ErrNaoAutorizado) {
			responderErro(w, http.StatusUnauthorized, "senha atual invalida")
			return
		}
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "senha alterada com sucesso"})
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
	if errors.Is(err, usuario.ErrEmailDuplicado) {
		responderErro(w, http.StatusConflict, "email ja cadastrado")
		return
	}
	if errors.Is(err, usuario.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "usuario foi alterado em outra sessao, recarregue os dados e tente novamente")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *UsuarioHandler) AlterarStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AlterarStatusInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AlterarStatus(r.Context(), id, input.Status)
	if errors.Is(err, usuario.ErrStatusInvalido) {
		responderErro(w, http.StatusBadRequest, "status invalido")
		return
	}
	if errors.Is(err, usuario.ErrUltimoAdmin) {
		responderErro(w, http.StatusConflict, "nao e possivel inativar o ultimo administrador")
		return
	}
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

func (h *UsuarioHandler) AlterarRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AlterarRoleInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AlterarRole(r.Context(), id, input.Role)
	if errors.Is(err, usuario.ErrRoleInvalida) {
		responderErro(w, http.StatusBadRequest, "role invalida")
		return
	}
	if errors.Is(err, usuario.ErrUltimoAdmin) {
		responderErro(w, http.StatusConflict, "nao e possivel remover o ultimo administrador")
		return
	}
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

func dataQuery(r *http.Request, nome string) (string, time.Time, error) {
	valor := r.URL.Query().Get(nome)
	if valor == "" {
		return "", time.Time{}, nil
	}
	data, err := time.Parse("2006-01-02", valor)
	if err != nil {
		return "", time.Time{}, err
	}
	return valor, data, nil
}

func periodoQuery(r *http.Request) (string, string, error) {
	dataInicio, inicio, err := dataQuery(r, "data_inicio")
	if err != nil {
		return "", "", fmt.Errorf("data_inicio deve estar no formato YYYY-MM-DD")
	}
	dataFim, fim, err := dataQuery(r, "data_fim")
	if err != nil {
		return "", "", fmt.Errorf("data_fim deve estar no formato YYYY-MM-DD")
	}
	if !inicio.IsZero() && !fim.IsZero() && fim.Before(inicio) {
		return "", "", fmt.Errorf("data_fim deve ser maior ou igual a data_inicio")
	}
	return dataInicio, dataFim, nil
}
