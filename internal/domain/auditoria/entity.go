package auditoria

import (
	"encoding/json"
	"time"
)

type Auditoria struct {
	Id              string
	UsuarioId       *string
	Acao            string
	Entidade        string
	EntidadeId      *string
	DadosAnteriores json.RawMessage
	DadosNovos      json.RawMessage
	IP              string
	UserAgent       string
	StatusCode      int
	TempoRespostaMs int
	RequestId       string
	CriadoEm        time.Time
}
