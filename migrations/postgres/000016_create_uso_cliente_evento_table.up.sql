CREATE TABLE uso_cliente_evento (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente(id) ON DELETE CASCADE,
    usuario_id UUID REFERENCES usuario(id) ON DELETE SET NULL,
    entidade VARCHAR(100),
    entidade_id UUID,
    tipo VARCHAR(100) NOT NULL,
    quantidade NUMERIC(14,4) NOT NULL DEFAULT 1,
    unidade VARCHAR(50) NOT NULL DEFAULT 'unidade',
    valor_unitario_centavos INT,
    metadados JSONB NOT NULL DEFAULT '{}'::jsonb,
    ocorrido_em TIMESTAMP NOT NULL DEFAULT NOW(),
    criado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_uso_cliente_evento_cliente ON uso_cliente_evento(cliente_id, ocorrido_em);
CREATE INDEX idx_uso_cliente_evento_tipo ON uso_cliente_evento(tipo, ocorrido_em);
CREATE INDEX idx_uso_cliente_evento_entidade ON uso_cliente_evento(entidade, entidade_id);