-- name: BuscarPlanoAtivoPorCliente
SELECT p.id, p.nome, p.descricao, p.valor_mensal_centavos, p.limites, p.status
FROM assinatura_cliente ac
JOIN plano p ON p.id = ac.plano_id
WHERE ac.cliente_id = @cliente_id
  AND ac.status = 'ativa'
  AND p.status = 'ativo'
  AND (ac.encerrado_em IS NULL OR ac.encerrado_em > NOW())
ORDER BY ac.iniciado_em DESC
LIMIT 1;
