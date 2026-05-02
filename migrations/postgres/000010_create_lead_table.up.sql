CREATE TABLE lead (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    landingpage_id UUID NOT NULL REFERENCES landingpage(id) ON DELETE CASCADE,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    telefone VARCHAR(20),
    whatsapp VARCHAR(20),
    mensagem TEXT,
    campos_personalizados JSONB,
    origem VARCHAR(100) NOT NULL DEFAULT 'buildnix',
    ip VARCHAR(50),
    user_agent TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'novo',
    criado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lead_landingpage ON lead(landingpage_id);
CREATE INDEX idx_lead_email ON lead(email);
CREATE INDEX idx_lead_status ON lead(status);
CREATE INDEX idx_lead_criado_em ON lead(criado_em);
