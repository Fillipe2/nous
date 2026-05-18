package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	ucleadaquisicao "buildnix-backend/internal/app/usecase/lead_aquisicao"
	domain "buildnix-backend/internal/domain/lead_aquisicao"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type LeadAquisicaoHandler struct {
	useCase *ucleadaquisicao.UseCase
}

func NewLeadAquisicaoHandler(useCase *ucleadaquisicao.UseCase) *LeadAquisicaoHandler {
	return &LeadAquisicaoHandler{useCase: useCase}
}

func (h *LeadAquisicaoHandler) RegistrarLeadAquisicao(w http.ResponseWriter, r *http.Request) {
	var input dto.RegistrarLeadAquisicaoInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	ip := middleware.ClienteIP(r)
	userAgent := r.Header.Get("User-Agent")

	output, err := h.useCase.RegistrarLeadAquisicao(r.Context(), input, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSpamDetectado):
			responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "Obrigado! Entraremos em contato em breve."})
			return
		case errors.Is(err, domain.ErrEmailDuplicado):
			responderErro(w, http.StatusConflict, "email ja cadastrado")
			return
		default:
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	responderJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       output.Id,
		"mensagem": "Obrigado! Entraremos em contato em breve.",
	})
}

func (h *LeadAquisicaoHandler) ListarLeadsAquisicao(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	filtro := domain.FiltroListagem{
		Nome:        r.URL.Query().Get("nome"),
		Email:       r.URL.Query().Get("email"),
		StatusFunil: r.URL.Query().Get("status_funil"),
		Porte:       r.URL.Query().Get("porte"),
		Cidade:      r.URL.Query().Get("cidade"),
		Estado:      r.URL.Query().Get("estado"),
		UTMCampaign: r.URL.Query().Get("utm_campaign"),
		DataInicio:  dataInicio,
		DataFim:     dataFim,
		Pagina:      pagina,
		PorPagina:   porPagina,
		OrdenarPor:  r.URL.Query().Get("ordenar_por"),
		Ordem:       r.URL.Query().Get("ordem"),
	}

	output, err := h.useCase.ListarLeadsAquisicao(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LeadAquisicaoHandler) ObterLeadAquisicao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	output, err := h.useCase.ObterLeadAquisicao(r.Context(), id)
	if errors.Is(err, domain.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "lead de aquisicao nao encontrado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *LeadAquisicaoHandler) AtualizarFunilLeadAquisicao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarFunilLeadAquisicaoInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	if err := h.useCase.AtualizarFunilLeadAquisicao(r.Context(), id, input); err != nil {
		switch {
		case errors.Is(err, domain.ErrNaoEncontrado):
			responderErro(w, http.StatusNotFound, "lead de aquisicao nao encontrado")
			return
		case errors.Is(err, domain.ErrStatusInvalido):
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		default:
			responderErro(w, http.StatusInternalServerError, "erro interno")
			return
		}
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "funil atualizado"})
}
