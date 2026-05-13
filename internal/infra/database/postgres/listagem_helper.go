package postgres

import (
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, constraint)
}

func colunaOrdenacao(valor, padrao string, permitidas map[string]string) string {
	valor = strings.ToLower(strings.TrimSpace(valor))
	if coluna, ok := permitidas[valor]; ok {
		return coluna
	}
	return permitidas[padrao]
}

func direcaoOrdenacao(valor string) string {
	if strings.EqualFold(strings.TrimSpace(valor), "asc") {
		return "ASC"
	}
	return "DESC"
}

func dataInicioFiltro(valor string) any {
	if data, ok := parseDataISO(valor); ok {
		return data
	}
	return valor
}

func dataFimFiltro(valor string) (string, any) {
	if data, ok := parseDataISO(valor); ok {
		return "<", data.AddDate(0, 0, 1)
	}
	return "<=", valor
}

func parseDataISO(valor string) (time.Time, bool) {
	valor = strings.TrimSpace(valor)
	data, err := time.Parse("2006-01-02", valor)
	return data, err == nil
}
