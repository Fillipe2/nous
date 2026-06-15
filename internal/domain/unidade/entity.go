package unidade

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("unidade nao encontrado")
	ErrVersaoConflito = errors.New("conflito de versao do unidade")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Unidade struct {
	Id           string
	NegocioId    string
	Nome         string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}
