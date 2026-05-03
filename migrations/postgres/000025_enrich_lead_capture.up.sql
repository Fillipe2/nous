ALTER TABLE lead
    ADD COLUMN cliente_id UUID REFERENCES cliente(id) ON DELETE SET NULL,
    ADD COLUMN consentimento BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN consentimento_em TIMESTAMPTZ,
    ADD COLUMN politica_versao VARCHAR(50),
    ADD COLUMN utm_source VARCHAR(255),
    ADD COLUMN utm_medium VARCHAR(255),
    ADD COLUMN utm_campaign VARCHAR(255),
    ADD COLUMN utm_content VARCHAR(255),
    ADD COLUMN utm_term VARCHAR(255),
    ADD COLUMN referrer TEXT,
    ADD COLUMN perdido_motivo VARCHAR(255),
    ADD COLUMN anonimizado_em TIMESTAMPTZ,
    ADD COLUMN atualizado_em TIMESTAMPTZ;

UPDATE lead l
SET cliente_id = lp.cliente_id,
    atualizado_em = l.criado_em
FROM landingpage lp
WHERE lp.id = l.landingpage_id;

CREATE INDEX idx_lead_cliente ON lead(cliente_id);
CREATE INDEX idx_lead_email_landingpage_criado ON lead(email, landingpage_id, criado_em DESC);
CREATE INDEX idx_lead_status_criado ON lead(status, criado_em);
CREATE INDEX idx_lead_anonimizado ON lead(anonimizado_em) WHERE anonimizado_em IS NULL;
CREATE INDEX idx_lead_utm_campaign ON lead(utm_campaign) WHERE utm_campaign IS NOT NULL;
