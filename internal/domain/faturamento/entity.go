package faturamento

import (
	"encoding/json"
	"time"
)

// UsoClienteEvento registra uma unidade de consumo mensurável por cliente.
type UsoClienteEvento struct {
	Id                   string
	ClienteId            string
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

// FaturamentoMensal representa o consolidado financeiro de um cliente em uma competência.
type FaturamentoMensal struct {
	Id                   string
	ClienteId            string
	AssinaturaId         string
	Competencia          time.Time
	ValorMensalCentavos  int
	ValorUsoCentavos     int
	ValorTotalCentavos   int
	Status               string // aberto | fechado
	FechadoEm            *time.Time
}

// BiMetricaDiaria armazena uma métrica agregada por cliente/dia.
type BiMetricaDiaria struct {
	Id           string
	ClienteId    string
	Data         time.Time
	Metrica      string
	Valor        float64
	Dimensoes    json.RawMessage
	DimensaoHash string
}
