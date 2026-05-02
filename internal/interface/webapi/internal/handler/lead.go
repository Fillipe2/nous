package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	uclead "buildnix-backend/internal/app/usecase/lead"
	"buildnix-backend/internal/domain/lead"

	"github.com/go-chi/chi/v5"
)

type LeadHandler struct {
	useCase *uclead.UseCase
}

func NewLeadHandler(useCase *uclead.UseCase) *LeadHandler {
	return &LeadHandler{useCase: useCase}
}

func (h *LeadHandler) CapturarLead(w http.ResponseWriter, r *http.Request) {
	var input dto.CapturarLeadInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	output, err := h.useCase.CapturarLead(r.Context(), input, ip, userAgent)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       output.Id,
		"mensagem": "Obrigado! Entraremos em contato em breve.",
	})
}

func (h *LeadHandler) ListarLeads(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	filtro := lead.FiltroListagem{
		ClienteId:     r.URL.Query().Get("cliente_id"),
		LandingPageId: r.URL.Query().Get("landingpage_id"),
		Nome:          r.URL.Query().Get("nome"),
		Email:         r.URL.Query().Get("email"),
		Status:        r.URL.Query().Get("status"),
		DataInicio:    r.URL.Query().Get("data_inicio"),
		DataFim:       r.URL.Query().Get("data_fim"),
		Pagina:        pagina,
		PorPagina:     porPagina,
		OrdenarPor:    r.URL.Query().Get("ordenar_por"),
		Ordem:         r.URL.Query().Get("ordem"),
	}

	output, err := h.useCase.ListarLeads(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LeadHandler) ObterLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	output, err := h.useCase.ObterLead(r.Context(), id)
	if errors.Is(err, lead.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "lead nao encontrado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LeadHandler) AtualizarStatusLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarStatusLeadInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	if err := h.useCase.AtualizarStatusLead(r.Context(), id, input); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "status atualizado"})
}
