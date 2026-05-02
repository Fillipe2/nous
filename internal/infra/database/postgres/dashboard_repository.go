package postgres

import (
	"context"
	"errors"
	"fmt"

	"buildnix-backend/internal/domain/dashboard"

	"github.com/jackc/pgx/v5"
)

type DashboardRepository struct {
	connection *Postgres
}

var _ dashboard.Repository = (*DashboardRepository)(nil)

func NewDashboardRepository(connection *Postgres) (*DashboardRepository, error) {
	return &DashboardRepository{connection: connection}, nil
}

func (r *DashboardRepository) ObterResumoCliente(ctx context.Context, clienteId string, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoCliente, error) {
	resumo := &dashboard.ResumoCliente{ClienteId: clienteId}
	args := argumentosPeriodo(pgx.NamedArgs{"cliente_id": clienteId}, filtro)
	landingPagePeriodo := condicaoPeriodo("criado_em", filtro)
	exportacaoPeriodo := condicaoPeriodo("e.criado_em", filtro)

	query := fmt.Sprintf(`SELECT
    (SELECT COUNT(*) FROM usuario_cliente WHERE cliente_id=@cliente_id) AS usuarios_vinculados,
    (SELECT COUNT(*) FROM landingpage WHERE cliente_id=@cliente_id AND status != 'excluida' %s) AS landingpages_criadas,
    (SELECT COUNT(*) FROM landingpage WHERE cliente_id=@cliente_id AND status = 'publicada' %s) AS landingpages_online,
    (SELECT COUNT(*) FROM exportacao e INNER JOIN landingpage lp ON lp.id = e.landingpage_id WHERE lp.cliente_id=@cliente_id %s) AS exportacoes_geradas`, landingPagePeriodo, landingPagePeriodo, exportacaoPeriodo)

	err := r.connection.QueryRow(ctx, query, args).Scan(
		&resumo.UsuariosVinculados,
		&resumo.LandingPagesCriadas,
		&resumo.LandingPagesOnline,
		&resumo.ExportacoesGeradas,
	)
	if err != nil {
		return nil, fmt.Errorf("obter resumo dashboard cliente: %w", err)
	}

	landingPageMaisUsada := &dashboard.LandingPageUtilizada{}
	queryMaisUsada := fmt.Sprintf(`SELECT
    lp.id,
    lp.nome,
    (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id %s) AS total_exportacoes
FROM landingpage lp
WHERE lp.cliente_id=@cliente_id AND lp.status != 'excluida'
ORDER BY (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id %s) DESC, lp.criado_em DESC
LIMIT 1`, exportacaoPeriodo, exportacaoPeriodo)
	err = r.connection.QueryRow(ctx, queryMaisUsada, args).Scan(
		&landingPageMaisUsada.Id,
		&landingPageMaisUsada.Nome,
		&landingPageMaisUsada.TotalExportacoes,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("obter landingpage mais utilizada dashboard: %w", err)
	}
	if err == nil {
		resumo.LandingPageMaisUsada = landingPageMaisUsada
	}

	return resumo, nil
}

func (r *DashboardRepository) ObterResumoAdmin(ctx context.Context, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoAdmin, error) {
	resumo := &dashboard.ResumoAdmin{}
	args := argumentosPeriodo(pgx.NamedArgs{}, filtro)
	periodo := condicaoPeriodo("criado_em", filtro)
	periodoExportacao := condicaoPeriodo("e.criado_em", filtro)
	periodoAssinatura := condicaoPeriodo("ac.criado_em", filtro)

	query := fmt.Sprintf(`SELECT
	(SELECT COUNT(*) FROM cliente WHERE status = 'ativo' %s) AS clientes_ativos,
	(SELECT COUNT(*) FROM usuario WHERE status = 'ativo' %s) AS usuarios_ativos,
	(SELECT COUNT(*) FROM template WHERE origem = 'sognix' AND status != 'excluido' %s) AS templates_sognix,
	(SELECT COUNT(*) FROM template WHERE origem = 'cliente' AND status != 'excluido' %s) AS templates_clientes,
	(SELECT COUNT(*) FROM landingpage WHERE status != 'excluida' %s) AS landingpages_criadas,
	(SELECT COUNT(*) FROM landingpage WHERE status = 'publicada' %s) AS landingpages_online,
	(SELECT COUNT(*) FROM lead WHERE 1=1 %s) AS leads_capturados,
	(SELECT COUNT(*) FROM exportacao e WHERE 1=1 %s) AS exportacoes_geradas,
	(
		SELECT COUNT(DISTINCT ac.cliente_id)
		FROM assinatura_cliente ac
		JOIN plano p ON p.id = ac.plano_id
		WHERE ac.status = 'ativa'
		  AND p.status = 'ativo'
		  AND (ac.encerrado_em IS NULL OR ac.encerrado_em > NOW())
		  AND COALESCE((p.limites->>'landingpages_customizaveis')::boolean, false) = true
		  %s
	) AS clientes_com_plano_customizavel`, periodo, periodo, periodo, periodo, periodo, periodo, periodo, periodoExportacao, periodoAssinatura)

	err := r.connection.QueryRow(ctx, query, args).Scan(
		&resumo.ClientesAtivos,
		&resumo.UsuariosAtivos,
		&resumo.TemplatesSognix,
		&resumo.TemplatesClientes,
		&resumo.LandingPagesCriadas,
		&resumo.LandingPagesOnline,
		&resumo.LeadsCapturados,
		&resumo.ExportacoesGeradas,
		&resumo.ClientesComPlanoCustom,
	)
	if err != nil {
		return nil, fmt.Errorf("obter resumo dashboard admin: %w", err)
	}
	return resumo, nil
}

func argumentosPeriodo(args pgx.NamedArgs, filtro dashboard.FiltroPeriodo) pgx.NamedArgs {
	if filtro.DataInicio != "" {
		args["data_inicio"] = filtro.DataInicio
	}
	if filtro.DataFim != "" {
		args["data_fim"] = filtro.DataFim
	}
	return args
}

func condicaoPeriodo(campo string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.DataInicio != "" {
		condicao += fmt.Sprintf(" AND %s >= @data_inicio", campo)
	}
	if filtro.DataFim != "" {
		condicao += fmt.Sprintf(" AND %s <= @data_fim", campo)
	}
	return condicao
}
