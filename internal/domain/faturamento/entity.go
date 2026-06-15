package faturamento

import (
	"encoding/json"
	"time"
)

// UsoNegocioEvento registra uma unidade de consumo mensurável por negocio.
type UsoNegocioEvento struct {
	Id                   string
	NegocioId            string
	UsuarioId            string
	Entidade             string
	EntidadeId           string
	Tipo                 string
	Quantidade           float64
	Unidade              string
	ValorUnitarioCentavos int
	Metadados            json.RawMessage
	OcorridoEm           time.Time
}

// FaturamentoMensal representa o consolidado financeiro de um negocio em uma competência.
type FaturamentoMensal struct {
	Id                   string
	NegocioId            string
	AssinaturaId         string
	Competencia          time.Time
	ValorMensalCentavos  int
	ValorUsoCentavos     int
	ValorTotalCentavos   int
	Status               string // aberto | fechado
	FechadoEm            *time.Time
}

// BiMetricaDiaria armazena uma métrica agregada por negocio/dia.
type BiMetricaDiaria struct {
	Id           string
	NegocioId    string
	Data         time.Time
	Metrica      string
	Valor        float64
	Dimensoes    json.RawMessage
	DimensaoHash string
}
