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
	ErrSemArquivo         = errors.New("template nao possui arquivo para download")
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
	Arquivo      []byte
	NomeArquivo  string
	ContentType  string
	TamanhoBytes int64
	Status       Status
	CriadoEm     time.Time
	AtualizadoEm time.Time
	Version      int
}

// TemArquivo indica se o template possui um pacote binário (ex.: ZIP white label)
// armazenado, em vez de apenas conteúdo textual.
func (t *Template) TemArquivo() bool {
	return t.TamanhoBytes > 0 || len(t.Arquivo) > 0
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
