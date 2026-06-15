package handler

import (
	"net/http"

	"nous/internal/app/dto"
	"nous/internal/domain/auditoria"

	"github.com/go-chi/chi/v5"
)

type AuditoriaHandler struct {
	repository auditoria.Repository
}

func NewAuditoriaHandler(repository auditoria.Repository) *AuditoriaHandler {
	return &AuditoriaHandler{repository: repository}
}

func (h *AuditoriaHandler) ListarAuditoria(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	filtro := auditoria.FiltroListagem{
		UsuarioId:  r.URL.Query().Get("usuario_id"),
		Acao:       r.URL.Query().Get("acao"),
		Entidade:   r.URL.Query().Get("entidade"),
		EntidadeId: r.URL.Query().Get("entidade_id"),
		StatusCode: r.URL.Query().Get("status_code"),
		RequestId:  r.URL.Query().Get("request_id"),
		DataInicio: dataInicio,
		DataFim:    dataFim,
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	}

	auditorias, total, err := h.repository.Listar(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	dados := make([]*dto.AuditoriaOutput, 0, len(auditorias))
	for _, a := range auditorias {
		dados = append(dados, mapearAuditoriaOutput(a))
	}
	responderJSON(w, http.StatusOK, dto.ListaAuditoriaOutput{
		Dados:     dados,
		Total:     total,
		Pagina:    pagina,
		PorPagina: porPagina,
	})
}

func (h *AuditoriaHandler) ObterAuditoria(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.repository.BuscarPorID(r.Context(), id)
	if err != nil {
		responderErro(w, http.StatusNotFound, "registro nao encontrado")
		return
	}
	responderJSON(w, http.StatusOK, mapearAuditoriaOutput(a))
}

func mapearAuditoriaOutput(a *auditoria.Auditoria) *dto.AuditoriaOutput {
	return &dto.AuditoriaOutput{
		Id:              a.Id,
		UsuarioId:       a.UsuarioId,
		Acao:            a.Acao,
		Entidade:        a.Entidade,
		EntidadeId:      a.EntidadeId,
		DadosAnteriores: a.DadosAnteriores,
		DadosNovos:      a.DadosNovos,
		IP:              a.IP,
		UserAgent:       a.UserAgent,
		StatusCode:      a.StatusCode,
		TempoRespostaMs: a.TempoRespostaMs,
		RequestId:       a.RequestId,
		CriadoEm:        a.CriadoEm,
	}
}
