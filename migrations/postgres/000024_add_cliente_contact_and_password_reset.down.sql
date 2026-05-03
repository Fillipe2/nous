DROP TABLE IF EXISTS usuario_reset_senha_token;

ALTER TABLE cliente
    DROP COLUMN IF EXISTS whatsapp,
    DROP COLUMN IF EXISTS telefone;
