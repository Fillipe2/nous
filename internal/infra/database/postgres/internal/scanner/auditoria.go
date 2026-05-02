// Package scanner contém funções puras que mapeiam rows do PostgreSQL para entidades de domínio.
package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/auditoria"
)

// Auditoria escaneia uma row do banco para a entidade de domínio Auditoria.
func Auditoria(row pgx.Row) (*auditoria.Auditoria, error) {
	var (
		id              string
		usuarioID       *string
		acao            string
		entidade        *string
		entidadeID      *string
		dadosAnteriores []byte
		dadosNovos      []byte
		ip              *string
		userAgent       *string
		statusCode      *int
		tempoRespostaMs *int
		requestID       *string
		criadoEm        time.Time
	)
	if err := row.Scan(
		&id, &usuarioID, &acao, &entidade, &entidadeID,
		&dadosAnteriores, &dadosNovos, &ip, &userAgent, &statusCode,
		&tempoRespostaMs, &requestID, &criadoEm,
	); err != nil {
		return nil, err
	}
	return &auditoria.Auditoria{
		Id:              id,
		UsuarioId:       usuarioID,
		Acao:            acao,
		Entidade:        strPtrParaStr(entidade),
		EntidadeId:      entidadeID,
		DadosAnteriores: json.RawMessage(dadosAnteriores),
		DadosNovos:      json.RawMessage(dadosNovos),
		IP:              strPtrParaStr(ip),
		UserAgent:       strPtrParaStr(userAgent),
		StatusCode:      intPtrParaInt(statusCode),
		TempoRespostaMs: intPtrParaInt(tempoRespostaMs),
		RequestId:       strPtrParaStr(requestID),
		CriadoEm:        criadoEm,
	}, nil
}

// AuditoriaFromRow escaneia campos em ordem das queries de listagem.
func AuditoriaFromRow(
	id string,
	usuarioID *string,
	acao string,
	entidade *string,
	entidadeID *string,
	dadosAnteriores []byte,
	dadosNovos []byte,
	ip *string,
	userAgent *string,
	statusCode *int,
	tempoRespostaMs *int,
	requestID *string,
	criadoEm time.Time,
) *auditoria.Auditoria {
	return &auditoria.Auditoria{
		Id:              id,
		UsuarioId:       usuarioID,
		Acao:            acao,
		Entidade:        strPtrParaStr(entidade),
		EntidadeId:      entidadeID,
		DadosAnteriores: json.RawMessage(dadosAnteriores),
		DadosNovos:      json.RawMessage(dadosNovos),
		IP:              strPtrParaStr(ip),
		UserAgent:       strPtrParaStr(userAgent),
		StatusCode:      intPtrParaInt(statusCode),
		TempoRespostaMs: intPtrParaInt(tempoRespostaMs),
		RequestId:       strPtrParaStr(requestID),
		CriadoEm:        criadoEm,
	}
}

func strPtrParaStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intPtrParaInt(n *int) int {
	if n == nil {
		return 0
	}
	return *n
}
