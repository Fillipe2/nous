package categoria

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrada  = errors.New("categoria nao encontrada")
	ErrVersaoConflito = errors.New("conflito de versao da categoria")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Categoria struct {
	Id           string
	Nome         string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}
