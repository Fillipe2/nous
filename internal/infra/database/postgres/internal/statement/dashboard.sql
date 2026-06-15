-- name: ObterResumoDashboardNegocio
SELECT
    (SELECT COUNT(*) FROM usuario_negocio WHERE negocio_id=@negocio_id) AS usuarios_vinculados,
    (SELECT COUNT(*) FROM landingpage WHERE negocio_id=@negocio_id AND status != 'excluida') AS landingpages_criadas,
    (SELECT COUNT(*) FROM landingpage WHERE negocio_id=@negocio_id AND status = 'publicada') AS landingpages_online,
    (SELECT COUNT(*) FROM exportacao e INNER JOIN landingpage lp ON lp.id = e.landingpage_id WHERE lp.negocio_id=@negocio_id) AS exportacoes_geradas;

-- name: ObterLandingPageMaisUtilizadaDashboard
SELECT
    lp.id,
    lp.nome,
    (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id) AS total_exportacoes
FROM landingpage lp
WHERE lp.negocio_id=@negocio_id AND lp.status != 'excluida'
ORDER BY (SELECT COUNT(*) FROM exportacao e WHERE e.landingpage_id = lp.id) DESC, lp.criado_em DESC
LIMIT 1;

-- name: ObterResumoDashboardAdmin
SELECT
    (SELECT COUNT(*) FROM negocio WHERE status = 'ativo') AS negocios_ativos,
    (SELECT COUNT(*) FROM usuario WHERE status = 'ativo') AS usuarios_ativos,
    (SELECT COUNT(*) FROM template WHERE origem = 'sognix' AND status != 'excluido') AS templates_sognix,
    (SELECT COUNT(*) FROM template WHERE origem = 'negocio' AND status != 'excluido') AS templates_negocios,
    (SELECT COUNT(*) FROM landingpage WHERE status != 'excluida') AS landingpages_criadas,
    (SELECT COUNT(*) FROM landingpage WHERE status = 'publicada') AS landingpages_online,
    (SELECT COUNT(*) FROM lead) AS leads_capturados,
    (SELECT COUNT(*) FROM exportacao) AS exportacoes_geradas,
    (
        SELECT COUNT(DISTINCT ac.negocio_id)
        FROM assinatura_negocio ac
        JOIN plano p ON p.id = ac.plano_id
        WHERE ac.status = 'ativa'
          AND p.status = 'ativo'
          AND (ac.encerrado_em IS NULL OR ac.encerrado_em > NOW())
          AND COALESCE((p.limites->>'landingpages_customizaveis')::boolean, false) = true
    ) AS negocios_com_plano_customizavel;
