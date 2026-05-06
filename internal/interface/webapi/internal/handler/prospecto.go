package handler

import (
	"errors"
	"net/http"

	"buildnix-backend/internal/app/dto"
	ucprospecto "buildnix-backend/internal/app/usecase/prospecto"
	"buildnix-backend/internal/domain/prospecto"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type ProspectoHandler struct {
	useCase *ucprospecto.UseCase
}

func NewProspectoHandler(useCase *ucprospecto.UseCase) *ProspectoHandler {
	return &ProspectoHandler{useCase: useCase}
}

func (h *ProspectoHandler) RegistrarProspecto(w http.ResponseWriter, r *http.Request) {
	var input dto.RegistrarProspectoInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	ip := middleware.ClienteIP(r)
	userAgent := r.Header.Get("User-Agent")

	output, err := h.useCase.RegistrarProspecto(r.Context(), input, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, prospecto.ErrSpamDetectado):
			responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "Obrigado! Entraremos em contato em breve."})
			return
		case errors.Is(err, prospecto.ErrEmailDuplicado):
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

func (h *ProspectoHandler) ListarProspectos(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	dataInicio, dataFim, err := periodoQuery(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	filtro := prospecto.FiltroListagem{
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

	output, err := h.useCase.ListarProspectos(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *ProspectoHandler) ObterProspecto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	output, err := h.useCase.ObterProspecto(r.Context(), id)
	if errors.Is(err, prospecto.ErrNaoEncontrado) {
		responderErro(w, http.StatusNotFound, "prospecto nao encontrado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *ProspectoHandler) AtualizarFunilProspecto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input dto.AtualizarFunilProspectoInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	if err := h.useCase.AtualizarFunilProspecto(r.Context(), id, input); err != nil {
		switch {
		case errors.Is(err, prospecto.ErrNaoEncontrado):
			responderErro(w, http.StatusNotFound, "prospecto nao encontrado")
			return
		case errors.Is(err, prospecto.ErrStatusInvalido):
			responderErro(w, http.StatusBadRequest, err.Error())
			return
		default:
			responderErro(w, http.StatusInternalServerError, "erro interno")
			return
		}
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "funil atualizado"})
}
