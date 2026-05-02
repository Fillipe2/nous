package template

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado      = errors.New("template nao encontrado")
	ErrOrigemInvalida     = errors.New("origem do template invalida")
	ErrClienteObrigatorio = errors.New("cliente_id obrigatorio para template do cliente")
	ErrNaoAutorizado      = errors.New("acesso nao autorizado ao template")
	ErrPlanoNaoPermite    = errors.New("plano nao permite este tipo de template")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Template struct {
	Id           string
	ClienteId    string
	CategoriaId  string
	Tipo         Tipo
	Origem       Origem
	Nome         string
	ArquivoURL   string
	Conteudo     string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}

type Origem string

const (
	OrigemSognix  Origem = "sognix"
	OrigemCliente Origem = "cliente"
)

func NormalizarOrigem(valor string) (Origem, error) {
	switch Origem(valor) {
	case "", OrigemSognix:
		return OrigemSognix, nil
	case OrigemCliente:
		return OrigemCliente, nil
	default:
		return "", ErrOrigemInvalida
	}
}

type Tipo string

const (
	TipoWhiteLabel       Tipo = "white_label"
	TipoBaseCustomizavel Tipo = "base_customizavel"
)

func NormalizarTipo(valor string) Tipo {
	if Tipo(valor) == TipoBaseCustomizavel {
		return TipoBaseCustomizavel
	}
	return TipoWhiteLabel
}
