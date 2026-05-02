package landingpage

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNaoEncontrada        = errors.New("landingpage nao encontrada")
	ErrBloqueada            = errors.New("landingpage bloqueada por outro usuario")
	ErrVersaoConflito       = errors.New("conflito de versao, recarregue e tente novamente")
	ErrNaoAutorizado        = errors.New("acesso nao autorizado a esta landingpage")
	ErrPlanoNaoPermite      = errors.New("plano nao permite este tipo de landingpage")
	ErrTemplateObrigatorio  = errors.New("template obrigatorio para landingpage white label")
	ErrTemplateIncompativel = errors.New("template incompativel com o tipo de landingpage")
	ErrTipoInvalido         = errors.New("tipo de landingpage invalido")
)

type Status string

const (
	StatusRascunho  Status = "rascunho"
	StatusPublicada Status = "publicada"
	StatusArquivada Status = "arquivada"
	StatusExcluida  Status = "excluida"
)

type LandingPage struct {
	Id             string
	ClienteId      string
	NegocioId      string
	CategoriaId    string
	TemplateId     string
	Tipo           Tipo
	Nome           string
	Conteudo       string
	Status         Status
	IsDraft        bool
	AutosaveData   json.RawMessage
	LastAutosaveEm *time.Time
	LockedBy       *string
	LockedAt       *time.Time
	CriadoEm       time.Time
	AtualizadoEm   time.Time
	Version        int
}

type Tipo string

const (
	TipoWhiteLabel   Tipo = "white_label"
	TipoCustomizavel Tipo = "customizavel"
)

func NormalizarTipo(valor string) (Tipo, error) {
	switch Tipo(valor) {
	case "", TipoWhiteLabel:
		return TipoWhiteLabel, nil
	case TipoCustomizavel:
		return TipoCustomizavel, nil
	default:
		return "", ErrTipoInvalido
	}
}

const DuracaoLock = 5 * time.Minute
