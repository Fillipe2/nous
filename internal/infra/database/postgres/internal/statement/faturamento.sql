-- name: RegistrarUsoClienteEvento
INSERT INTO uso_cliente_evento (
    id, cliente_id, usuario_id, entidade, entidade_id, tipo,
    quantidade, unidade, valor_unitario_centavos, metadados, ocorrido_em
) VALUES (
    @id, @cliente_id, @usuario_id, @entidade, @entidade_id, @tipo,
    @quantidade, @unidade, @valor_unitario_centavos, @metadados, @ocorrido_em
);

-- name: ObterFaturamentoMensalPorCompetencia
SELECT id, cliente_id, assinatura_id, competencia,
    valor_mensal_centavos, valor_uso_centavos, valor_total_centavos,
    status, fechado_em
FROM faturamento_mensal
WHERE cliente_id=@cliente_id AND competencia=@competencia;

-- name: CriarFaturamentoMensal
INSERT INTO faturamento_mensal (cliente_id, competencia, status)
VALUES (@cliente_id, @competencia, 'aberto')
ON CONFLICT (cliente_id, competencia) DO NOTHING;

-- name: AdicionarUsoFaturamentoMensal
UPDATE faturamento_mensal
SET valor_uso_centavos = valor_uso_centavos + @centavos,
    valor_total_centavos = valor_mensal_centavos + valor_uso_centavos + @centavos,
    atualizado_em = NOW()
WHERE cliente_id=@cliente_id AND competencia=@competencia AND status='aberto';

-- name: FecharFaturamentoMensal
UPDATE faturamento_mensal
SET status='fechado', fechado_em=NOW(), atualizado_em=NOW()
WHERE cliente_id=@cliente_id AND competencia=@competencia AND status='aberto';

-- name: IncrementarBiMetricaDiaria
INSERT INTO bi_metrica_diaria (cliente_id, data, metrica, valor, dimensoes, dimensao_hash, atualizado_em)
VALUES (@cliente_id, @data, @metrica, @valor, @dimensoes, @dimensao_hash, NOW())
ON CONFLICT (cliente_id, data, metrica, dimensao_hash)
DO UPDATE SET
    valor = bi_metrica_diaria.valor + EXCLUDED.valor,
    atualizado_em = NOW();
