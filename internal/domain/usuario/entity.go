package usuario

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("usuario nao encontrado")
	ErrEmailDuplicado = errors.New("email ja cadastrado")
	ErrSenhaInvalida  = errors.New("senha invalida")
	ErrNaoAutorizado  = errors.New("nao autorizado")
	ErrVersaoConflito = errors.New("conflito de versao do usuario")
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleUsuario Role = "usuario"
)

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

type Usuario struct {
	Id           string
	Nome         string
	Email        string
	SenhaHash    string
	Role         Role
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}

type UsuarioAuth struct {
	Id            string
	UsuarioId     string
	Tipo          string
	ProviderId    string
	ProviderEmail string
	CriadoEm      time.Time
}

type RefreshToken struct {
	Id         string
	UsuarioId  string
	TokenHash  string
	ExpiresAt  time.Time
	RevogadoEm *time.Time
	CriadoEm   time.Time
}
