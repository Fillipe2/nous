CREATE TABLE template (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    categoria_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
    nome VARCHAR(255) NOT NULL,
    arquivo_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ativo',
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    atualizado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);
