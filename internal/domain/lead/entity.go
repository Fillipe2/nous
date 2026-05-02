package lead

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNaoEncontrado = errors.New("lead nao encontrado")
)

type Status string

const (
	StatusNovo       Status = "novo"
	StatusContatado  Status = "contatado"
	StatusConvertido Status = "convertido"
	StatusPerdido    Status = "perdido"
)

type Lead struct {
	Id                   string
	LandingPageId        string
	Nome                 string
	Email                string
	Telefone             string
	Whatsapp             string
	Mensagem             string
	CamposPersonalizados json.RawMessage
	Origem               string
	IP                   string
	UserAgent            string
	Status               Status
	CriadoEm             time.Time
}
