CREATE TABLE usuario_verificacao_email_token (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    usado_em TIMESTAMPTZ,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_usuario_verificacao_email_token_usuario ON usuario_verificacao_email_token(usuario_id);
CREATE INDEX idx_usuario_verificacao_email_token_expires_at ON usuario_verificacao_email_token(expires_at);
