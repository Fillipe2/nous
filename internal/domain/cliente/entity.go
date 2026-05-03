package cliente

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("cliente nao encontrado")
	ErrNaoAutorizado  = errors.New("acesso nao autorizado ao cliente")
	ErrVersaoConflito = errors.New("conflito de versao do cliente")
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Cliente struct {
	Id           string
	Nome         string
	Email        string
	Telefone     string
	Whatsapp     string
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}

type UsuarioCliente struct {
	UsuarioId string
	ClienteId string
	CriadoEm  time.Time
}
