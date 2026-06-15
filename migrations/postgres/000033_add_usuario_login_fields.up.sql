-- email_verified: marcado true em login social (Google) e no seed de admin.
-- ultimo_login_em: registrado a cada login bem-sucedido (auditoria/segurança).
ALTER TABLE usuario ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE usuario ADD COLUMN ultimo_login_em TIMESTAMP;
