package handler

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"buildnix-backend/internal/app/dto"
	uctemplate "buildnix-backend/internal/app/usecase/template"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/template"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type TemplateHandler struct {
	useCase *uctemplate.UseCase
}

func NewTemplateHandler(useCase *uctemplate.UseCase) *TemplateHandler {
	return &TemplateHandler{useCase: useCase}
}

func (h *TemplateHandler) UploadTemplate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		responderErro(w, http.StatusBadRequest, "falha ao processar formulario")
		return
	}

	categoriaId := r.FormValue("categoria_id")
	clienteId := r.FormValue("cliente_id")
	nome := r.FormValue("nome")
	if nome == "" {
		responderErro(w, http.StatusBadRequest, "nome e obrigatorio")
		return
	}

	claims := middleware.ObterClaims(r.Context())

	conteudo := strings.TrimSpace(r.FormValue("conteudo"))
	arquivoURL := ""
	file, header, err := r.FormFile("arquivo")
	if err == nil {
		defer file.Close()
		dados, readErr := io.ReadAll(file)
		if readErr != nil {
			responderErro(w, http.StatusInternalServerError, "erro ao ler arquivo")
			return
		}
		if conteudo == "" {
			conteudo = string(dados)
		}
		arquivoURL = "templates/" + filepath.Base(header.Filename)
	}

	output, err := h.useCase.CriarTemplate(r.Context(), dto.CriarTemplateInput{
		ClienteId:   clienteId,
		CategoriaId: categoriaId,
		Tipo:        r.FormValue("tipo"),
		Origem:      r.FormValue("origem"),
		Nome:        nome,
		ArquivoURL:  arquivoURL,
		Conteudo:    conteudo,
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, template.ErrNaoAutorizado) || errors.Is(err, template.ErrPlanoNaoPermite) || errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *TemplateHandler) ListarTemplates(w http.ResponseWriter, r *http.Request) {
	categoriaId := r.URL.Query().Get("categoria_id")
	clienteId := r.URL.Query().Get("cliente_id")
	tipo := r.URL.Query().Get("tipo")
	origem := r.URL.Query().Get("origem")
	pagina, porPagina := paginacao(r)
	claims := middleware.ObterClaims(r.Context())

	output, err := h.useCase.ListarTemplates(r.Context(), template.FiltroListagem{
		ClienteId:   clienteId,
		CategoriaId: categoriaId,
		Tipo:        tipo,
		Origem:      origem,
		Nome:        r.URL.Query().Get("nome"),
		Status:      r.URL.Query().Get("status"),
		DataInicio:  r.URL.Query().Get("data_inicio"),
		DataFim:     r.URL.Query().Get("data_fim"),
		Pagina:      pagina,
		PorPagina:   porPagina,
		OrdenarPor:  r.URL.Query().Get("ordenar_por"),
		Ordem:       r.URL.Query().Get("ordem"),
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, template.ErrClienteObrigatorio) {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, template.ErrNaoAutorizado) || errors.Is(err, cliente.ErrNaoAutorizado) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *TemplateHandler) ExcluirTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.useCase.ExcluirTemplate(r.Context(), id); err != nil {
		if errors.Is(err, template.ErrNaoEncontrado) {
			responderErro(w, http.StatusNotFound, "template nao encontrado")
			return
		}
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
