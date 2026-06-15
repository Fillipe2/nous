package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nous/internal/domain/faturamento"
	"nous/internal/infra/database/postgres/internal/statement"
)

// UsoRepository implementa faturamento.UsoRepository para PostgreSQL.
type UsoRepository struct {
	connection *Postgres
}

var _ faturamento.UsoRepository = (*UsoRepository)(nil)

func NewUsoRepository(connection *Postgres) (*UsoRepository, error) {
	return &UsoRepository{connection: connection}, nil
}

func (r *UsoRepository) Registrar(ctx context.Context, uso *faturamento.UsoNegocioEvento) error {
	if uso.Id == "" {
		uso.Id = uuid.New().String()
	}
	if uso.OcorridoEm.IsZero() {
		uso.OcorridoEm = time.Now().UTC()
	}
	if uso.Unidade == "" {
		uso.Unidade = "unidade"
	}
	metadados := uso.Metadados
	if metadados == nil {
		metadados = json.RawMessage("{}")
	}
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.RegistrarUsoNegocioEvento, pgx.NamedArgs{
		"id":                      uso.Id,
		"negocio_id":              uso.NegocioId,
		"usuario_id":              nilableStr(uso.UsuarioId),
		"entidade":                nilableStr(uso.Entidade),
		"entidade_id":             nilableStr(uso.EntidadeId),
		"tipo":                    uso.Tipo,
		"quantidade":              uso.Quantidade,
		"unidade":                 uso.Unidade,
		"valor_unitario_centavos": nilableInt(uso.ValorUnitarioCentavos),
		"metadados":               metadados,
		"ocorrido_em":             uso.OcorridoEm,
	})
	if err != nil {
		return fmt.Errorf("registrar uso negocio evento: %w", err)
	}
	return nil
}

// FaturamentoRepository implementa faturamento.FaturamentoRepository para PostgreSQL.
type FaturamentoRepository struct {
	connection *Postgres
}

var _ faturamento.FaturamentoRepository = (*FaturamentoRepository)(nil)

func NewFaturamentoRepository(connection *Postgres) (*FaturamentoRepository, error) {
	return &FaturamentoRepository{connection: connection}, nil
}

func (r *FaturamentoRepository) ObterOuCriarCompetencia(ctx context.Context, negocioId string, competencia time.Time) (*faturamento.FaturamentoMensal, error) {
	comp := inicioMes(competencia)
	stmts := statement.Get()

	// Garante que o registro existe antes de consultar.
	if err := r.connection.Exec(ctx, stmts.CriarFaturamentoMensal, pgx.NamedArgs{
		"negocio_id":  negocioId,
		"competencia": comp,
	}); err != nil {
		return nil, fmt.Errorf("criar faturamento mensal: %w", err)
	}

	row := r.connection.QueryRow(ctx, stmts.ObterFaturamentoMensalPorCompetencia, pgx.NamedArgs{
		"negocio_id":  negocioId,
		"competencia": comp,
	})

	var fm faturamento.FaturamentoMensal
	var assinaturaId *string
	var fechadoEm *time.Time
	if err := row.Scan(
		&fm.Id, &fm.NegocioId, &assinaturaId, &fm.Competencia,
		&fm.ValorMensalCentavos, &fm.ValorUsoCentavos, &fm.ValorTotalCentavos,
		&fm.Status, &fechadoEm,
	); err != nil {
		return nil, fmt.Errorf("obter faturamento mensal: %w", err)
	}
	if assinaturaId != nil {
		fm.AssinaturaId = *assinaturaId
	}
	fm.FechadoEm = fechadoEm
	return &fm, nil
}

func (r *FaturamentoRepository) AdicionarUso(ctx context.Context, negocioId string, competencia time.Time, centavos int) error {
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.AdicionarUsoFaturamentoMensal, pgx.NamedArgs{
		"negocio_id":  negocioId,
		"competencia": inicioMes(competencia),
		"centavos":    centavos,
	}); err != nil {
		return fmt.Errorf("adicionar uso faturamento mensal: %w", err)
	}
	return nil
}

func (r *FaturamentoRepository) FecharCompetencia(ctx context.Context, negocioId string, competencia time.Time) error {
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.FecharFaturamentoMensal, pgx.NamedArgs{
		"negocio_id":  negocioId,
		"competencia": inicioMes(competencia),
	}); err != nil {
		return fmt.Errorf("fechar faturamento mensal: %w", err)
	}
	return nil
}

// BiRepository implementa faturamento.BiRepository para PostgreSQL.
type BiRepository struct {
	connection *Postgres
}

var _ faturamento.BiRepository = (*BiRepository)(nil)

func NewBiRepository(connection *Postgres) (*BiRepository, error) {
	return &BiRepository{connection: connection}, nil
}

func (r *BiRepository) IncrementarMetrica(ctx context.Context, negocioId string, data time.Time, metrica string, valor float64, dimensoes map[string]string) error {
	dimJSON, hash := serializarDimensoes(dimensoes)
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.IncrementarBiMetricaDiaria, pgx.NamedArgs{
		"negocio_id":    nilableStr(negocioId),
		"data":          data.Truncate(24 * time.Hour),
		"metrica":       metrica,
		"valor":         valor,
		"dimensoes":     dimJSON,
		"dimensao_hash": hash,
	}); err != nil {
		return fmt.Errorf("incrementar bi metrica diaria: %w", err)
	}
	return nil
}

func inicioMes(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func nilableInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func serializarDimensoes(dim map[string]string) (json.RawMessage, string) {
	if len(dim) == 0 {
		h := fmt.Sprintf("%x", sha256.Sum256([]byte("{}")))
		return json.RawMessage("{}"), h[:16]
	}
	keys := make([]string, 0, len(dim))
	for k := range dim {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%q:%q", k, dim[k]))
	}
	sb.WriteString("}")
	raw := sb.String()
	h := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	data, _ := json.Marshal(dim)
	return json.RawMessage(data), h[:16]
}
