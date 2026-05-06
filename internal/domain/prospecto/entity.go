package prospecto

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("prospecto nao encontrado")
	ErrEmailDuplicado = errors.New("prospecto com este email ja existe")
	ErrStatusInvalido = errors.New("status de funil invalido")
	ErrSpamDetectado  = errors.New("submissao identificada como spam")
)

type Porte string

const (
	PorteMicro   Porte = "micro"
	PortePequeno Porte = "pequeno"
	PorteMedio   Porte = "medio"
	PorteGrande  Porte = "grande"
)

func PorteValido(p Porte) bool {
	switch p {
	case PorteMicro, PortePequeno, PorteMedio, PorteGrande:
		return true
	}
	return p == ""
}

type StatusFunil string

const (
	StatusFunilNovo        StatusFunil = "novo"
	StatusFunilContatado   StatusFunil = "contatado"
	StatusFunilQualificado StatusFunil = "qualificado"
	StatusFunilNegociacao  StatusFunil = "negociacao"
	StatusFunilConvertido  StatusFunil = "convertido"
	StatusFunilPerdido     StatusFunil = "perdido"
)

func StatusFunilValido(s StatusFunil) bool {
	switch s {
	case StatusFunilNovo, StatusFunilContatado, StatusFunilQualificado,
		StatusFunilNegociacao, StatusFunilConvertido, StatusFunilPerdido:
		return true
	}
	return false
}

type Prospecto struct {
	Id                       string
	Nome                     string
	Email                    string
	Telefone                 string
	CategoriaPretendidaId    string
	CategoriaPretendidaLivre string
	Cidade                   string
	Estado                   string
	Porte                    Porte
	Interesses               json.RawMessage
	Mensagem                 string
	Origem                   string
	UTMSource                string
	UTMMedium                string
	UTMCampaign              string
	UTMContent               string
	UTMTerm                  string
	StatusFunil              StatusFunil
	ConvertidoEmUsuarioId    string
	IP                       string
	UserAgent                string
	CriadoEm                 time.Time
	AtualizadoEm             time.Time
}
