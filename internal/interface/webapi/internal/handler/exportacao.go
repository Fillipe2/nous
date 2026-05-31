package handler

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	ucexportacao "nous/internal/app/usecase/exportacao"
	"nous/internal/domain/exportacao"
	"nous/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type ExportacaoHandler struct {
	useCase *ucexportacao.UseCase
}

func NewExportacaoHandler(useCase *ucexportacao.UseCase) *ExportacaoHandler {
	return &ExportacaoHandler{useCase: useCase}
}

func (h *ExportacaoHandler) ExportarLandingPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		responderErro(w, http.StatusBadRequest, "falha ao processar formulario")
		return
	}

	file, header, err := r.FormFile("arquivo")
	if err != nil {
		responderErro(w, http.StatusBadRequest, "arquivo zip obrigatorio")
		return
	}
	defer file.Close()
	if !contentTypeZIP(header.Header.Get("Content-Type")) {
		responderErro(w, http.StatusBadRequest, "arquivo zip invalido")
		return
	}

	dados, err := io.ReadAll(file)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro ao ler arquivo")
		return
	}
	if err := validarZIP(dados); err != nil {
		responderErro(w, http.StatusBadRequest, "arquivo zip invalido")
		return
	}

	nomeArquivo := filepath.Base(header.Filename)
	if nomeArquivo == "" || nomeArquivo == "." {
		nomeArquivo = "exportacao_" + time.Now().Format("20060102150405") + ".zip"
	}
	if filepath.Ext(nomeArquivo) != ".zip" {
		nomeArquivo += ".zip"
	}

	contentType := "application/zip"
	output, err := h.useCase.ExportarLandingPage(r.Context(), id, claims.UsuarioId, claims.Role, nomeArquivo, contentType, dados)
	if errors.Is(err, exportacao.ErrNaoAutorizada) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro ao processar exportacao")
		return
	}

	responderJSON(w, http.StatusCreated, output)
}

func (h *ExportacaoHandler) ListarExportacoes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ObterClaims(r.Context())
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.useCase.ListarExportacoes(r.Context(), exportacao.FiltroListagem{
		ClienteId:     r.URL.Query().Get("cliente_id"),
		LandingPageId: r.URL.Query().Get("landingpage_id"),
		UsuarioId:     r.URL.Query().Get("usuario_id"),
		NomeArquivo:   r.URL.Query().Get("nome_arquivo"),
		DataInicio:    dataInicio,
		DataFim:       dataFim,
		Pagina:        pagina,
		PorPagina:     porPagina,
		OrdenarPor:    r.URL.Query().Get("ordenar_por"),
		Ordem:         r.URL.Query().Get("ordem"),
	}, claims.UsuarioId, claims.Role)
	if errors.Is(err, exportacao.ErrNaoAutorizada) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *ExportacaoHandler) BaixarExportacao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.ObterClaims(r.Context())

	dados, meta, err := h.useCase.BaixarExportacao(r.Context(), id, claims.UsuarioId, claims.Role)
	if errors.Is(err, exportacao.ErrNaoEncontrada) {
		responderErro(w, http.StatusNotFound, "exportacao nao encontrada")
		return
	}
	if errors.Is(err, exportacao.ErrNaoAutorizada) {
		responderErro(w, http.StatusForbidden, "acesso negado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro ao baixar arquivo")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(meta.NomeArquivo)+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dados)
}

func validarZIP(dados []byte) error {
	if len(dados) == 0 {
		return fmt.Errorf("zip vazio")
	}
	leitor, err := zip.NewReader(bytes.NewReader(dados), int64(len(dados)))
	if err != nil {
		return err
	}
	const limiteDescomprimido = 200 << 20
	var total uint64
	for _, arquivo := range leitor.File {
		if arquivo.FileInfo().IsDir() {
			continue
		}
		nome := filepath.Clean(arquivo.Name)
		if filepath.IsAbs(nome) || strings.HasPrefix(nome, "..") {
			return fmt.Errorf("entrada zip insegura")
		}
		total += arquivo.UncompressedSize64
		if total > limiteDescomprimido {
			return fmt.Errorf("zip excede limite descomprimido")
		}
	}
	return nil
}

func contentTypeZIP(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	return contentType == "" || contentType == "application/zip" || contentType == "application/x-zip-compressed" || contentType == "application/octet-stream"
}
