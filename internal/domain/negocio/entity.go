package negocio

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado     = errors.New("negocio nao encontrado")
	ErrNaoAutorizado     = errors.New("acesso nao autorizado ao negocio")
	ErrVersaoConflito    = errors.New("conflito de versao do negocio")
	ErrDocumentoInvalido = errors.New("documento (CPF/CNPJ) invalido")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Negocio struct {
	Id           string
	Nome         string
	Email        string
	Telefone     string
	Whatsapp     string
	Documento    string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}

type UsuarioNegocio struct {
	UsuarioId string
	NegocioId string
	CriadoEm  time.Time
}
