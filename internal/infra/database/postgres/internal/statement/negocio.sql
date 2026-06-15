-- name: SalvarNegocio
INSERT INTO negocio (id, nome, email, telefone, whatsapp, documento, status, criado_em, atualizado_em, version)
VALUES (@id, @nome, @email, @telefone, @whatsapp, @documento, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarNegocioPorID
SELECT id, nome, email, telefone, whatsapp, documento, status, criado_em, atualizado_em, version
FROM negocio WHERE id=@id AND status != 'excluido';

-- name: AtualizarNegocio
UPDATE negocio
SET nome=@nome, email=@email, telefone=@telefone, whatsapp=@whatsapp, documento=@documento, status=@status, atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteNegocio
UPDATE negocio SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarNegocios
SELECT COUNT(*) FROM negocio WHERE status != 'excluido';

-- name: ListarNegocios
SELECT id, nome, email, telefone, whatsapp, documento, status, criado_em, atualizado_em, version
FROM negocio WHERE status != 'excluido'
ORDER BY criado_em DESC LIMIT @limit OFFSET @offset;

-- name: VincularUsuarioNegocio
INSERT INTO usuario_negocio (usuario_id, negocio_id, criado_em)
VALUES (@usuario_id, @negocio_id, @criado_em)
ON CONFLICT DO NOTHING;

-- name: ListarNegociosPorUsuario
SELECT c.id, c.nome, c.email, c.telefone, c.whatsapp, c.documento, c.status, c.criado_em, c.atualizado_em, c.version
FROM negocio c
INNER JOIN usuario_negocio uc ON uc.negocio_id = c.id
WHERE uc.usuario_id=@usuario_id AND c.status != 'excluido'
ORDER BY c.nome;

-- name: VerificarAcessoNegocio
SELECT EXISTS (
		SELECT 1
		FROM usuario_negocio uc
		INNER JOIN negocio c ON c.id = uc.negocio_id
		WHERE uc.negocio_id=@negocio_id
			AND uc.usuario_id=@usuario_id
			AND c.status != 'excluido'
) AS acesso;
