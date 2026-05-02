ALTER TABLE auditoria
    ADD COLUMN user_agent TEXT,
    ADD COLUMN status_code INT,
    ADD COLUMN tempo_resposta_ms INT,
    ADD COLUMN request_id VARCHAR(100);

CREATE INDEX idx_auditoria_request_id ON auditoria(request_id);
CREATE INDEX idx_auditoria_status_code ON auditoria(status_code, criado_em);
