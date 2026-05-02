package exportacao

import (
	"errors"
	"time"
)

var (
	ErrNaoEncontrada = errors.New("exportacao nao encontrada")
	ErrNaoAutorizada = errors.New("acesso nao autorizado a exportacao")
)

type Exportacao struct {
	Id            string
	LandingPageId string
	UsuarioId     string
	NomeArquivo   string
	ContentType   string
	TamanhoBytes  int64
	Arquivo       []byte
	Hash          string
	CriadoEm      time.Time
}
