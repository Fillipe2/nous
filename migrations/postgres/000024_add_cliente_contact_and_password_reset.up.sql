ALTER TABLE cliente
    ADD COLUMN telefone VARCHAR(30),
    ADD COLUMN whatsapp VARCHAR(30);

CREATE TABLE usuario_reset_senha_token (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    usado_em TIMESTAMPTZ,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usuario_reset_senha_token_usuario ON usuario_reset_senha_token(usuario_id);
CREATE INDEX idx_usuario_reset_senha_token_expires_at ON usuario_reset_senha_token(expires_at);
