CREATE TABLE landingpage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente(id) ON DELETE CASCADE,
    negocio_id UUID REFERENCES negocio(id) ON DELETE SET NULL,
    categoria_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
    template_id UUID REFERENCES template(id) ON DELETE SET NULL,
    nome VARCHAR(255) NOT NULL,
    conteudo TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'rascunho',
    is_draft BOOLEAN NOT NULL DEFAULT TRUE,
    autosave_data JSONB,
    last_autosave_em TIMESTAMP,
    locked_by UUID REFERENCES usuario(id) ON DELETE SET NULL,
    locked_at TIMESTAMP,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_landingpage_cliente ON landingpage(cliente_id);
CREATE INDEX idx_landingpage_status ON landingpage(status);
