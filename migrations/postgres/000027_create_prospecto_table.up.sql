CREATE TABLE prospecto (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    telefone VARCHAR(50),
    categoria_pretendida_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
    categoria_pretendida_livre VARCHAR(255),
    cidade VARCHAR(100),
    estado VARCHAR(2),
    porte VARCHAR(20),
    interesses JSONB,
    mensagem TEXT,
    origem VARCHAR(100),
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    utm_content VARCHAR(255),
    utm_term VARCHAR(255),
    status_funil VARCHAR(30) NOT NULL DEFAULT 'novo',
    convertido_em_usuario_id UUID REFERENCES usuario(id) ON DELETE SET NULL,
    ip VARCHAR(45),
    user_agent TEXT,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_prospecto_email ON prospecto(email);
CREATE INDEX idx_prospecto_status_funil ON prospecto(status_funil);
CREATE INDEX idx_prospecto_criado_em ON prospecto(criado_em DESC);
CREATE INDEX idx_prospecto_utm_campaign ON prospecto(utm_campaign) WHERE utm_campaign IS NOT NULL;
