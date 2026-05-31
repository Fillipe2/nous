package postgres

import (
	"context"
	"errors"
	"fmt"

	"nous/internal/domain/dashboard"

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

func (r *DashboardRepository) ObterResumoUsuario(ctx context.Context, usuarioId string, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoUsuario, error) {
	resumo := &dashboard.ResumoUsuario{UsuarioId: usuarioId}
	args := argumentosDashboard(pgx.NamedArgs{"usuario_id": usuarioId}, filtro)
	landingPageFiltro := condicaoLandingPage("lp", filtro)
	landingPagePeriodo := condicaoPeriodo("lp.criado_em", filtro)
	exportacaoPeriodo := condicaoPeriodo("e.criado_em", filtro)

	query := fmt.Sprintf(`SELECT
    (SELECT COUNT(*) FROM usuario_cliente WHERE usuario_id=@usuario_id) AS clientes_vinculados,
    (SELECT COUNT(*) FROM landingpage lp INNER JOIN usuario_cliente uc ON uc.cliente_id=lp.cliente_id WHERE uc.usuario_id=@usuario_id AND lp.status != 'excluida' %s %s) AS landingpages_criadas,
    (SELECT COUNT(*) FROM landingpage lp INNER JOIN usuario_cliente uc ON uc.cliente_id=lp.cliente_id WHERE uc.usuario_id=@usuario_id AND lp.status = 'publicada' %s %s) AS landingpages_online,
    (SELECT COUNT(*) FROM exportacao e INNER JOIN landingpage lp ON lp.id=e.landingpage_id INNER JOIN usuario_cliente uc ON uc.cliente_id=lp.cliente_id WHERE uc.usuario_id=@usuario_id %s %s) AS exportacoes_geradas`, landingPageFiltro, landingPagePeriodo, landingPageFiltro, landingPagePeriodo, landingPageFiltro, exportacaoPeriodo)

	if err := r.connection.QueryRow(ctx, query, args).Scan(
		&resumo.ClientesVinculados,
		&resumo.LandingPagesCriadas,
		&resumo.LandingPagesOnline,
		&resumo.ExportacoesGeradas,
	); err != nil {
		return nil, fmt.Errorf("obter resumo dashboard usuario: %w", err)
	}

	landingPageMaisUsada := &dashboard.LandingPageUtilizada{}
	queryMaisUsada := fmt.Sprintf(`SELECT
    lp.id,
    lp.nome,
    (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id %s) AS total_exportacoes
FROM landingpage lp
INNER JOIN usuario_cliente uc ON uc.cliente_id=lp.cliente_id
WHERE uc.usuario_id=@usuario_id AND lp.status != 'excluida' %s
ORDER BY (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id %s) DESC, lp.criado_em DESC
LIMIT 1`, exportacaoPeriodo, landingPageFiltro, exportacaoPeriodo)
	if err := r.connection.QueryRow(ctx, queryMaisUsada, args).Scan(
		&landingPageMaisUsada.Id,
		&landingPageMaisUsada.Nome,
		&landingPageMaisUsada.TotalExportacoes,
	); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("obter landingpage mais utilizada dashboard usuario: %w", err)
	} else if err == nil {
		resumo.LandingPageMaisUsada = landingPageMaisUsada
	}

	return resumo, nil
}

func (r *DashboardRepository) ObterResumoAdmin(ctx context.Context, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoAdmin, error) {
	resumo := &dashboard.ResumoAdmin{}
	args := argumentosDashboard(pgx.NamedArgs{}, filtro)
	clienteFiltro := condicaoClienteAdmin("c", filtro)
	usuarioFiltro := condicaoUsuarioAdmin("u", filtro)
	templateSognixFiltro := condicaoTemplateAdmin("t", filtro, false)
	templateClienteFiltro := condicaoTemplateAdmin("t", filtro, true)
	landingPageFiltro := condicaoLandingPage("lp", filtro)
	exportacaoFiltro := condicaoExportacaoAdmin("e", "lp", filtro)
	assinaturaFiltro := condicaoAssinaturaAdmin("ac", filtro)
	periodoCliente := condicaoPeriodo("c.criado_em", filtro)
	periodoUsuario := condicaoPeriodo("u.criado_em", filtro)
	periodoTemplate := condicaoPeriodo("t.criado_em", filtro)
	periodoLandingPage := condicaoPeriodo("lp.criado_em", filtro)
	periodoLead := condicaoPeriodo("l.criado_em", filtro)
	periodoExportacao := condicaoPeriodo("e.criado_em", filtro)
	periodoAssinatura := condicaoPeriodo("ac.criado_em", filtro)

	query := fmt.Sprintf(`SELECT
	(SELECT COUNT(*) FROM cliente c WHERE c.status = 'ativo' %s %s) AS clientes_ativos,
	(SELECT COUNT(*) FROM usuario u WHERE u.status = 'ativo' %s %s) AS usuarios_ativos,
	(SELECT COUNT(*) FROM template t WHERE t.origem = 'sognix' AND t.status != 'excluido' %s %s) AS templates_sognix,
	(SELECT COUNT(*) FROM template t WHERE t.origem = 'cliente' AND t.status != 'excluido' %s %s) AS templates_clientes,
	(SELECT COUNT(*) FROM landingpage lp WHERE lp.status != 'excluida' %s %s) AS landingpages_criadas,
	(SELECT COUNT(*) FROM landingpage lp WHERE lp.status = 'publicada' %s %s) AS landingpages_online,
	(SELECT COUNT(*) FROM lead l INNER JOIN landingpage lp ON lp.id=l.landingpage_id WHERE 1=1 %s %s) AS leads_capturados,
	(SELECT COUNT(*) FROM exportacao e INNER JOIN landingpage lp ON lp.id=e.landingpage_id WHERE 1=1 %s %s) AS exportacoes_geradas,
	(
		SELECT COUNT(DISTINCT ac.cliente_id)
		FROM assinatura_cliente ac
		JOIN plano p ON p.id = ac.plano_id
		WHERE ac.status = 'ativa'
		  AND p.status = 'ativo'
		  AND (ac.encerrado_em IS NULL OR ac.encerrado_em > NOW())
		  AND COALESCE((p.limites->>'landingpages_customizaveis')::boolean, false) = true
		  %s %s
	) AS clientes_com_plano_customizavel`, clienteFiltro, periodoCliente, usuarioFiltro, periodoUsuario, templateSognixFiltro, periodoTemplate, templateClienteFiltro, periodoTemplate, landingPageFiltro, periodoLandingPage, landingPageFiltro, periodoLandingPage, landingPageFiltro, periodoLead, exportacaoFiltro, periodoExportacao, assinaturaFiltro, periodoAssinatura)

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
	return argumentosDashboard(args, filtro)
}

func argumentosDashboard(args pgx.NamedArgs, filtro dashboard.FiltroPeriodo) pgx.NamedArgs {
	if filtro.DataInicio != "" {
		args["data_inicio"] = dataInicioFiltro(filtro.DataInicio)
	}
	if filtro.DataFim != "" {
		_, valor := dataFimFiltro(filtro.DataFim)
		args["data_fim"] = valor
	}
	if filtro.ClienteId != "" {
		args["cliente_id"] = filtro.ClienteId
	}
	if filtro.UsuarioId != "" {
		args["usuario_id_filtro"] = filtro.UsuarioId
	}
	if filtro.LandingPageId != "" {
		args["landingpage_id"] = filtro.LandingPageId
	}
	if filtro.NegocioId != "" {
		args["negocio_id"] = filtro.NegocioId
	}
	if filtro.CategoriaId != "" {
		args["categoria_id"] = filtro.CategoriaId
	}
	if filtro.TemplateId != "" {
		args["template_id"] = filtro.TemplateId
	}
	if filtro.Tipo != "" {
		args["tipo"] = filtro.Tipo
	}
	return args
}

func condicaoPeriodo(campo string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.DataInicio != "" {
		condicao += fmt.Sprintf(" AND %s >= @data_inicio", campo)
	}
	if filtro.DataFim != "" {
		operador, _ := dataFimFiltro(filtro.DataFim)
		condicao += fmt.Sprintf(" AND %s %s @data_fim", campo, operador)
	}
	return condicao
}

func condicaoLandingPage(alias string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.ClienteId != "" {
		condicao += fmt.Sprintf(" AND %s.cliente_id=@cliente_id", alias)
	}
	if filtro.LandingPageId != "" {
		condicao += fmt.Sprintf(" AND %s.id=@landingpage_id", alias)
	}
	if filtro.NegocioId != "" {
		condicao += fmt.Sprintf(" AND %s.negocio_id=@negocio_id", alias)
	}
	if filtro.CategoriaId != "" {
		condicao += fmt.Sprintf(" AND %s.categoria_id=@categoria_id", alias)
	}
	if filtro.TemplateId != "" {
		condicao += fmt.Sprintf(" AND %s.template_id=@template_id", alias)
	}
	if filtro.Tipo != "" {
		condicao += fmt.Sprintf(" AND %s.tipo=@tipo", alias)
	}
	if filtro.UsuarioId != "" {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM usuario_cliente uc_filtro WHERE uc_filtro.cliente_id=%s.cliente_id AND uc_filtro.usuario_id=@usuario_id_filtro)", alias)
	}
	return condicao
}

func condicaoClienteAdmin(alias string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.ClienteId != "" {
		condicao += fmt.Sprintf(" AND %s.id=@cliente_id", alias)
	}
	if filtro.UsuarioId != "" {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM usuario_cliente uc_filtro WHERE uc_filtro.cliente_id=%s.id AND uc_filtro.usuario_id=@usuario_id_filtro)", alias)
	}
	if temFiltroLandingPage(filtro) {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM landingpage lp WHERE lp.cliente_id=%s.id %s)", alias, condicaoLandingPage("lp", filtro))
	}
	return condicao
}

func condicaoUsuarioAdmin(alias string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.UsuarioId != "" {
		condicao += fmt.Sprintf(" AND %s.id=@usuario_id_filtro", alias)
	}
	if filtro.ClienteId != "" && !temFiltroLandingPage(filtro) {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM usuario_cliente uc_filtro WHERE uc_filtro.usuario_id=%s.id AND uc_filtro.cliente_id=@cliente_id)", alias)
	}
	if temFiltroLandingPage(filtro) {
		condicao += fmt.Sprintf(` AND EXISTS (
			SELECT 1
			FROM usuario_cliente uc_filtro
			WHERE uc_filtro.usuario_id=%s.id
			  AND EXISTS (SELECT 1 FROM landingpage lp WHERE lp.cliente_id=uc_filtro.cliente_id %s)
		)`, alias, condicaoLandingPage("lp", filtro))
	}
	return condicao
}

func condicaoTemplateAdmin(alias string, filtro dashboard.FiltroPeriodo, templateCliente bool) string {
	condicao := ""
	if filtro.TemplateId != "" {
		condicao += fmt.Sprintf(" AND %s.id=@template_id", alias)
	}
	if filtro.CategoriaId != "" {
		condicao += fmt.Sprintf(" AND %s.categoria_id=@categoria_id", alias)
	}
	if templateCliente {
		if filtro.ClienteId != "" {
			condicao += fmt.Sprintf(" AND %s.cliente_id=@cliente_id", alias)
		}
		if filtro.UsuarioId != "" {
			condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM usuario_cliente uc_filtro WHERE uc_filtro.cliente_id=%s.cliente_id AND uc_filtro.usuario_id=@usuario_id_filtro)", alias)
		}
	}
	if filtro.LandingPageId != "" || filtro.NegocioId != "" || filtro.Tipo != "" {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM landingpage lp WHERE lp.template_id=%s.id %s)", alias, condicaoLandingPage("lp", filtro))
	}
	return condicao
}

func condicaoAssinaturaAdmin(alias string, filtro dashboard.FiltroPeriodo) string {
	condicao := ""
	if filtro.ClienteId != "" {
		condicao += fmt.Sprintf(" AND %s.cliente_id=@cliente_id", alias)
	}
	if filtro.UsuarioId != "" {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM usuario_cliente uc_filtro WHERE uc_filtro.cliente_id=%s.cliente_id AND uc_filtro.usuario_id=@usuario_id_filtro)", alias)
	}
	if temFiltroLandingPage(filtro) {
		condicao += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM landingpage lp WHERE lp.cliente_id=%s.cliente_id %s)", alias, condicaoLandingPage("lp", filtro))
	}
	return condicao
}

func condicaoExportacaoAdmin(aliasExportacao, aliasLandingPage string, filtro dashboard.FiltroPeriodo) string {
	condicao := condicaoLandingPage(aliasLandingPage, filtro)
	if filtro.UsuarioId != "" {
		condicao += fmt.Sprintf(" AND %s.usuario_id=@usuario_id_filtro", aliasExportacao)
	}
	return condicao
}

func temFiltroLandingPage(filtro dashboard.FiltroPeriodo) bool {
	return filtro.LandingPageId != "" || filtro.NegocioId != "" || filtro.CategoriaId != "" || filtro.TemplateId != "" || filtro.Tipo != ""
}
