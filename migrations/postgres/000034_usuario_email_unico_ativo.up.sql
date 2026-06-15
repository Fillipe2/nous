-- Troca o UNIQUE global de email por um índice único parcial que ignora
-- usuários excluídos (soft delete). Assim o email de um usuário excluído pode
-- ser reutilizado num novo cadastro, mas continua único entre os ativos/inativos.
ALTER TABLE usuario DROP CONSTRAINT usuario_email_key;
CREATE UNIQUE INDEX usuario_email_unico_ativo ON usuario (email) WHERE status <> 'excluido';
