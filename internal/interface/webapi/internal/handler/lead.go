package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"nous/internal/app/dto"
	uclead "nous/internal/app/usecase/lead"
	"nous/internal/domain/landingpage"
	"nous/internal/domain/lead"
	"nous/internal/interface/webapi/internal/middleware"

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

	ip := middleware.ClienteIP(r)
	userAgent := r.Header.Get("User-Agent")
	if input.Referrer == "" {
		input.Referrer = r.Header.Get("Referer")
	}

	output, err := h.useCase.CapturarLead(r.Context(), input, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, lead.ErrSpamDetectado):
			// Resposta neutra para nao ensinar bots.
			responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "Obrigado! Entraremos em contato em breve."})
			return
		case errors.Is(err, lead.ErrConsentimentoObrigatorio),
			errors.Is(err, lead.ErrPayloadGrande):
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, landingpage.ErrNaoEncontrada),
			errors.Is(err, lead.ErrLandingPageNaoPublicada):
			responderErro(w, http.StatusNotFound, "landingpage nao disponivel")
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

func (h *LeadHandler) ListarLeads(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	filtro := lead.FiltroListagem{
		ClienteId:     r.URL.Query().Get("cliente_id"),
		LandingPageId: r.URL.Query().Get("landingpage_id"),
		Nome:          r.URL.Query().Get("nome"),
		Email:         r.URL.Query().Get("email"),
		Status:        r.URL.Query().Get("status"),
		UTMSource:     r.URL.Query().Get("utm_source"),
		UTMCampaign:   r.URL.Query().Get("utm_campaign"),
		DataInicio:    dataInicio,
		DataFim:       dataFim,
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
		switch {
		case errors.Is(err, lead.ErrStatusInvalido), errors.Is(err, lead.ErrMotivoObrigatorio):
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		default:
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "status atualizado"})
}

func (h *LeadHandler) AnonimizarLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.useCase.AnonimizarLead(r.Context(), id); err != nil {
		if errors.Is(err, lead.ErrJaAnonimizado) {
			responderErro(w, http.StatusConflict, "lead ja anonimizado")
			return
		}
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "lead anonimizado"})
}

func (h *LeadHandler) AnonimizarLeadsExpirados(w http.ResponseWriter, r *http.Request) {
	retencao := retencaoDuracao(r.URL.Query().Get("dias"))
	total, err := h.useCase.AnonimizarExpirados(r.Context(), retencao)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, map[string]int{"anonimizados": total})
}

func retencaoDuracao(dias string) time.Duration {
	if dias == "" {
		return 0
	}
	n, err := strconv.Atoi(dias)
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * 24 * time.Hour
}
