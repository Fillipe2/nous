DROP INDEX IF EXISTS idx_auditoria_status_code;
DROP INDEX IF EXISTS idx_auditoria_request_id;

ALTER TABLE auditoria
    DROP COLUMN request_id,
    DROP COLUMN tempo_resposta_ms,
    DROP COLUMN status_code,
    DROP COLUMN user_agent;
