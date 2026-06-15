package lead

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNaoEncontrado            = errors.New("lead nao encontrado")
	ErrLandingPageNaoPublicada  = errors.New("landingpage nao publicada")
	ErrConsentimentoObrigatorio = errors.New("consentimento e obrigatorio")
	ErrLeadDuplicado            = errors.New("lead duplicado em janela curta")
	ErrStatusInvalido           = errors.New("status de lead invalido")
	ErrMotivoObrigatorio        = errors.New("motivo obrigatorio para status perdido")
	ErrSpamDetectado            = errors.New("submissao identificada como spam")
	ErrPayloadGrande            = errors.New("payload acima do limite permitido")
	ErrJaAnonimizado            = errors.New("lead ja anonimizado")
)

type Status string

const (
	StatusNovo             Status = "novo"
	StatusContatado        Status = "contatado"
	StatusQualificado      Status = "qualificado"
	StatusNegociacao       Status = "negociacao"
	StatusPropostaEnviada  Status = "proposta_enviada"
	StatusConvertido       Status = "convertido"
	StatusPerdido          Status = "perdido"
)

func StatusValido(s Status) bool {
	switch s {
	case StatusNovo, StatusContatado, StatusQualificado,
		StatusNegociacao, StatusPropostaEnviada, StatusConvertido, StatusPerdido:
		return true
	}
	return false
}

type UTM struct {
	Source   string
	Medium   string
	Campaign string
	Content  string
	Term     string
}

func (u UTM) Vazio() bool {
	return u.Source == "" && u.Medium == "" && u.Campaign == "" && u.Content == "" && u.Term == ""
}

type Lead struct {
	Id                   string
	LandingPageId        string
	NegocioId            string
	Nome                 string
	Email                string
	Telefone             string
	Whatsapp             string
	Mensagem             string
	CamposPersonalizados json.RawMessage
	Origem               string
	IP                   string
	UserAgent            string
	Referrer             string
	UTM                  UTM
	Consentimento                    bool
	ConsentimentoEm                  *time.Time
	PoliticaVersao                   string
	ConsentimentoMarketingParceiros  bool
	ConsentimentoVersaoMarketing     string
	Status                           Status
	PerdidoMotivo        string
	AnonimizadoEm        *time.Time
	CriadoEm             time.Time
	AtualizadoEm         time.Time
}
