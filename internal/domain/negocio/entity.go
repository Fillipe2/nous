package negocio

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("negocio nao encontrado")
	ErrVersaoConflito = errors.New("conflito de versao do negocio")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Negocio struct {
	Id           string
	ClienteId    string
	Nome         string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}
