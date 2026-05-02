CREATE TABLE usuario_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    tipo VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (tipo, provider_id)
);
