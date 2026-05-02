CREATE TABLE bi_metrica_diaria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID REFERENCES cliente(id) ON DELETE CASCADE,
    data DATE NOT NULL,
    metrica VARCHAR(100) NOT NULL,
    valor NUMERIC(18,4) NOT NULL DEFAULT 0,
    dimensoes JSONB NOT NULL DEFAULT '{}'::jsonb,
    dimensao_hash VARCHAR(64) NOT NULL DEFAULT '',
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (cliente_id, data, metrica, dimensao_hash)
);

CREATE INDEX idx_bi_metrica_diaria_cliente ON bi_metrica_diaria(cliente_id, data);
CREATE INDEX idx_bi_metrica_diaria_metrica ON bi_metrica_diaria(metrica, data);
