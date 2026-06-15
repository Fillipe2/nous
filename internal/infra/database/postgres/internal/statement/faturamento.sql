-- name: RegistrarUsoNegocioEvento
INSERT INTO uso_negocio_evento (
    id, negocio_id, usuario_id, entidade, entidade_id, tipo,
    quantidade, unidade, valor_unitario_centavos, metadados, ocorrido_em
) VALUES (
    @id, @negocio_id, @usuario_id, @entidade, @entidade_id, @tipo,
    @quantidade, @unidade, @valor_unitario_centavos, @metadados, @ocorrido_em
);

-- name: ObterFaturamentoMensalPorCompetencia
SELECT id, negocio_id, assinatura_id, competencia,
    valor_mensal_centavos, valor_uso_centavos, valor_total_centavos,
    status, fechado_em
FROM faturamento_mensal
WHERE negocio_id=@negocio_id AND competencia=@competencia;

-- name: CriarFaturamentoMensal
INSERT INTO faturamento_mensal (negocio_id, competencia, status)
VALUES (@negocio_id, @competencia, 'aberto')
ON CONFLICT (negocio_id, competencia) DO NOTHING;

-- name: AdicionarUsoFaturamentoMensal
UPDATE faturamento_mensal
SET valor_uso_centavos = valor_uso_centavos + @centavos,
    valor_total_centavos = valor_mensal_centavos + valor_uso_centavos + @centavos,
    atualizado_em = NOW()
WHERE negocio_id=@negocio_id AND competencia=@competencia AND status='aberto';

-- name: FecharFaturamentoMensal
UPDATE faturamento_mensal
SET status='fechado', fechado_em=NOW(), atualizado_em=NOW()
WHERE negocio_id=@negocio_id AND competencia=@competencia AND status='aberto';

-- name: IncrementarBiMetricaDiaria
INSERT INTO bi_metrica_diaria (negocio_id, data, metrica, valor, dimensoes, dimensao_hash, atualizado_em)
VALUES (@negocio_id, @data, @metrica, @valor, @dimensoes, @dimensao_hash, NOW())
ON CONFLICT (negocio_id, data, metrica, dimensao_hash)
DO UPDATE SET
    valor = bi_metrica_diaria.valor + EXCLUDED.valor,
    atualizado_em = NOW();
