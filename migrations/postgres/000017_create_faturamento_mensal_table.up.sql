CREATE TABLE faturamento_mensal (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente(id) ON DELETE CASCADE,
    assinatura_id UUID REFERENCES assinatura_cliente(id) ON DELETE SET NULL,
    competencia DATE NOT NULL,
    valor_mensal_centavos INT NOT NULL DEFAULT 0,
    valor_uso_centavos INT NOT NULL DEFAULT 0,
    valor_total_centavos INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'aberto',
    fechado_em TIMESTAMP,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (cliente_id, competencia)
);

CREATE INDEX idx_faturamento_mensal_status ON faturamento_mensal(status, competencia);