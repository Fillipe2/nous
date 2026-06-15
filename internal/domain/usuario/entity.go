package usuario

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrado  = errors.New("usuario nao encontrado")
	ErrEmailDuplicado = errors.New("email ja cadastrado")
	ErrSenhaInvalida  = errors.New("senha invalida")
	ErrSenhaFraca     = errors.New("senha deve ter no minimo 8 caracteres")
	ErrNaoAutorizado  = errors.New("nao autorizado")
	ErrVersaoConflito = errors.New("conflito de versao do usuario")
	ErrRoleInvalida   = errors.New("role invalida")
	ErrUltimoAdmin    = errors.New("nao e possivel remover o ultimo administrador")
	ErrStatusInvalido = errors.New("status invalido")
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleUsuario Role = "usuario"
)

// RoleValida informa se o valor corresponde a um papel conhecido.
func RoleValida(r Role) bool {
	return r == RoleAdmin || r == RoleUsuario
}

type Status string

const (
	StatusAtivo    Status = "ativo"
	StatusInativo  Status = "inativo"
	StatusExcluido Status = "excluido"
)

// StatusGerenciavelValido informa se o status pode ser definido manualmente por
// um admin. A exclusão usa o soft delete dedicado, não esta via.
func StatusGerenciavelValido(s Status) bool {
	return s == StatusAtivo || s == StatusInativo
}

type Usuario struct {
	Id            string
	Nome          string
	Email         string
	SenhaHash     string
	Role          Role
	Status        Status
	EmailVerified bool
	UltimoLoginEm *time.Time
	CriadoEm      time.Time
	AtualizadoEm  time.Time
	Version       int
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

type ResetSenhaToken struct {
	Id        string
	UsuarioId string
	TokenHash string
	ExpiresAt time.Time
	UsadoEm   *time.Time
	CriadoEm  time.Time
}

type VerificacaoEmailToken struct {
	Id        string
	UsuarioId string
	TokenHash string
	ExpiresAt time.Time
	UsadoEm   *time.Time
	CriadoEm  time.Time
}
