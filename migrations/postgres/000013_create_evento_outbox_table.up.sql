CREATE TABLE evento_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    schema_version INT NOT NULL DEFAULT 1,
    source VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pendente',
    tentativas INT NOT NULL DEFAULT 0,
    mensagem_erro TEXT,
    publicado_em TIMESTAMP,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evento_outbox_status ON evento_outbox(status, criado_em);
CREATE INDEX idx_evento_outbox_event_type ON evento_outbox(event_type);
