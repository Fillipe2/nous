package handler

import (
	"errors"
	"io"
	"net/http"

	"buildnix-backend/internal/app/dto"
	uclp "buildnix-backend/internal/app/usecase/landingpage"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type LandingPageHandler struct {
	useCase *uclp.UseCase
}

func NewLandingPageHandler(useCase *uclp.UseCase) *LandingPageHandler {
	return &LandingPageHandler{useCase: useCase}
}

func (h *LandingPageHandler) CriarLandingPage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	var input dto.CriarLandingPageInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.CriarLandingPage(r.Context(), input, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrNaoAutorizado) || errors.Is(err, cliente.ErrNaoAutorizado) || errors.Is(err, landingpage.ErrPlanoNaoPermite) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if errors.Is(err, landingpage.ErrTemplateObrigatorio) || errors.Is(err, landingpage.ErrTemplateIncompativel) || errors.Is(err, landingpage.ErrTipoInvalido) {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *LandingPageHandler) ListarLandingPages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacao(r)
	filtro := landingpage.FiltroListagem{
		ClienteId:   r.URL.Query().Get("cliente_id"),
		NegocioId:   r.URL.Query().Get("negocio_id"),
		CategoriaId: r.URL.Query().Get("categoria_id"),
		TemplateId:  r.URL.Query().Get("template_id"),
		Tipo:        r.URL.Query().Get("tipo"),
		Status:      r.URL.Query().Get("status"),
		Nome:        r.URL.Query().Get("nome"),
		IsDraft:     r.URL.Query().Get("is_draft"),
		DataInicio:  r.URL.Query().Get("data_inicio"),
		DataFim:     r.URL.Query().Get("data_fim"),
		Pagina:      pagina,
		PorPagina:   porPagina,
		OrdenarPor:  r.URL.Query().Get("ordenar_por"),
		Ordem:       r.URL.Query().Get("ordem"),
	}

	output, err := h.useCase.ListarLandingPages(r.Context(), filtro, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrNaoAutorizado) || errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LandingPageHandler) ObterLandingPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	output, err := h.useCase.ObterLandingPage(r.Context(), id, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrNaoEncontrada) {
		responderErro(w, http.StatusNotFound, "landingpage nao encontrada")
		return
	}
	if errors.Is(err, landingpage.ErrNaoAutorizado) || errors.Is(err, landingpage.ErrPlanoNaoPermite) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if errors.Is(err, landingpage.ErrTemplateObrigatorio) || errors.Is(err, landingpage.ErrTemplateIncompativel) || errors.Is(err, landingpage.ErrTipoInvalido) {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LandingPageHandler) AtualizarLandingPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	var input dto.AtualizarLandingPageInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.AtualizarLandingPage(r.Context(), id, input, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrNaoEncontrada) {
		responderErro(w, http.StatusNotFound, "landingpage nao encontrada")
		return
	}
	if errors.Is(err, landingpage.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if errors.Is(err, landingpage.ErrVersaoConflito) {
		responderErro(w, http.StatusConflict, "conflito de versao")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LandingPageHandler) ExcluirLandingPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())
	if err := h.useCase.ExcluirLandingPage(r.Context(), id, claims.UsuarioId, claims.Role); err != nil {
		if errors.Is(err, landingpage.ErrNaoAutorizado) || errors.Is(err, landingpage.ErrPlanoNaoPermite) {
			responderErro(w, http.StatusForbidden, "acesso negado")
			return
		}
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LandingPageHandler) Autosave(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())

	var input dto.AutosaveInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.Autosave(r.Context(), id, claims.UsuarioId, claims.Role, input)
	if err != nil {
		if errors.Is(err, landingpage.ErrNaoAutorizado) {
			responderErro(w, http.StatusForbidden, "acesso negado")
			return
		}
		if errors.Is(err, landingpage.ErrVersaoConflito) {
			responderErro(w, http.StatusConflict, "rascunho foi alterado em outra sessao, recarregue os dados e tente novamente")
			return
		}
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LandingPageHandler) BloquearEdicao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())

	output, err := h.useCase.BloquearEdicao(r.Context(), id, claims.UsuarioId, claims.Role)
	if errors.Is(err, landingpage.ErrBloqueada) {
		responderErro(w, http.StatusConflict, "landingpage bloqueada por outro usuario")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LandingPageHandler) DesbloquearEdicao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())

	if err := h.useCase.DesbloquearEdicao(r.Context(), id, claims.UsuarioId, claims.Role); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "edicao desbloqueada"})
}

func (h *LandingPageHandler) ExportarLandingPage(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	_, err := io.ReadAll(r.Body)
	if err != nil {
		responderErro(w, http.StatusBadRequest, "erro ao ler dados")
		return
	}
	responderJSON(w, http.StatusAccepted, map[string]string{"mensagem": "exportacao recebida - use POST /landingpages/{id}/exportar com multipart"})
}
