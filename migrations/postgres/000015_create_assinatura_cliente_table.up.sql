CREATE TABLE assinatura_cliente (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente(id) ON DELETE CASCADE,
    plano_id UUID NOT NULL REFERENCES plano(id) ON DELETE RESTRICT,
    status VARCHAR(50) NOT NULL DEFAULT 'ativa',
    iniciado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    encerrado_em TIMESTAMP,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assinatura_cliente_cliente ON assinatura_cliente(cliente_id);
CREATE INDEX idx_assinatura_cliente_status ON assinatura_cliente(status);