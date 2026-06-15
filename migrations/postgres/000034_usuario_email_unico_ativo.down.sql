DROP INDEX usuario_email_unico_ativo;
ALTER TABLE usuario ADD CONSTRAINT usuario_email_key UNIQUE (email);
