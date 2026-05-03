DROP INDEX IF EXISTS idx_lead_utm_campaign;
DROP INDEX IF EXISTS idx_lead_anonimizado;
DROP INDEX IF EXISTS idx_lead_status_criado;
DROP INDEX IF EXISTS idx_lead_email_landingpage_criado;
DROP INDEX IF EXISTS idx_lead_cliente;

ALTER TABLE lead
    DROP COLUMN IF EXISTS atualizado_em,
    DROP COLUMN IF EXISTS anonimizado_em,
    DROP COLUMN IF EXISTS perdido_motivo,
    DROP COLUMN IF EXISTS referrer,
    DROP COLUMN IF EXISTS utm_term,
    DROP COLUMN IF EXISTS utm_content,
    DROP COLUMN IF EXISTS utm_campaign,
    DROP COLUMN IF EXISTS utm_medium,
    DROP COLUMN IF EXISTS utm_source,
    DROP COLUMN IF EXISTS politica_versao,
    DROP COLUMN IF EXISTS consentimento_em,
    DROP COLUMN IF EXISTS consentimento,
    DROP COLUMN IF EXISTS cliente_id;
