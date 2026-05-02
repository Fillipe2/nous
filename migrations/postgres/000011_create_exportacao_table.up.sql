CREATE TABLE exportacao (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    landingpage_id UUID NOT NULL REFERENCES landingpage(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    arquivo_url TEXT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exportacao_landingpage ON exportacao(landingpage_id);
