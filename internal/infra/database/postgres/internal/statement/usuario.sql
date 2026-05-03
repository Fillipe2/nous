-- name: SalvarUsuario
INSERT INTO usuario (id, nome, email, senha_hash, role, status, criado_em, atualizado_em, version)
VALUES (@id, @nome, @email, @senha_hash, @role, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarUsuarioPorID
SELECT id, nome, email, senha_hash, role, status, criado_em, atualizado_em, version
FROM usuario WHERE id = @id AND status != 'excluido';

-- name: BuscarUsuarioPorEmail
SELECT id, nome, email, senha_hash, role, status, criado_em, atualizado_em, version
FROM usuario WHERE email = @email;

-- name: AtualizarUsuario
UPDATE usuario
SET nome=@nome, email=@email, senha_hash=@senha_hash, role=@role, status=@status,
    atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteUsuario
UPDATE usuario SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarUsuarios
SELECT COUNT(*) FROM usuario WHERE status != 'excluido';

-- name: ListarUsuarios
SELECT id, nome, email, senha_hash, role, status, criado_em, atualizado_em, version
FROM usuario WHERE status != 'excluido'
ORDER BY criado_em DESC LIMIT @limit OFFSET @offset;

-- name: SalvarUsuarioAuth
INSERT INTO usuario_auth (id, usuario_id, tipo, provider_id, provider_email, criado_em)
VALUES (@id, @usuario_id, @tipo, @provider_id, @provider_email, @criado_em)
ON CONFLICT (tipo, provider_id) DO UPDATE SET provider_email = EXCLUDED.provider_email;

-- name: BuscarUsuarioAuthPorProvider
SELECT id, usuario_id, tipo, provider_id, provider_email, criado_em
FROM usuario_auth WHERE tipo=@tipo AND provider_id=@provider_id;

-- name: SalvarUsuarioRefreshToken
INSERT INTO usuario_refresh_token (id, usuario_id, token_hash, expires_at, criado_em)
VALUES (@id, @usuario_id, @token_hash, @expires_at, @criado_em);

-- name: BuscarUsuarioRefreshTokenPorHash
SELECT id, usuario_id, token_hash, expires_at, revogado_em, criado_em
FROM usuario_refresh_token
WHERE token_hash=@token_hash;

-- name: RevogarUsuarioRefreshToken
UPDATE usuario_refresh_token
SET revogado_em=@revogado_em
WHERE token_hash=@token_hash AND revogado_em IS NULL;

-- name: RevogarUsuarioRefreshTokensDoUsuario
UPDATE usuario_refresh_token
SET revogado_em=@revogado_em
WHERE usuario_id=@usuario_id AND revogado_em IS NULL;

-- name: ExpurgarUsuarioRefreshTokens
DELETE FROM usuario_refresh_token
WHERE expires_at < @antes_de OR (revogado_em IS NOT NULL AND revogado_em < @antes_de);

-- name: SalvarUsuarioResetSenhaToken
INSERT INTO usuario_reset_senha_token (id, usuario_id, token_hash, expires_at, criado_em)
VALUES (@id, @usuario_id, @token_hash, @expires_at, @criado_em);

-- name: BuscarUsuarioResetSenhaTokenPorHash
SELECT id, usuario_id, token_hash, expires_at, usado_em, criado_em
FROM usuario_reset_senha_token
WHERE token_hash=@token_hash;

-- name: MarcarUsuarioResetSenhaTokenUsado
UPDATE usuario_reset_senha_token
SET usado_em=@usado_em
WHERE token_hash=@token_hash AND usado_em IS NULL;
