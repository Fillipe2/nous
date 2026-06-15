-- name: SalvarUnidade
INSERT INTO unidade (id, negocio_id, nome, status, criado_em, atualizado_em, version)
VALUES (@id, @negocio_id, @nome, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarUnidadePorID
SELECT id, negocio_id, nome, status, criado_em, atualizado_em, version
FROM unidade WHERE id=@id AND status != 'excluido';

-- name: AtualizarUnidade
UPDATE unidade
SET nome=@nome, status=@status, atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteUnidade
UPDATE unidade SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarUnidadesPorNegocio
SELECT COUNT(*) FROM unidade WHERE negocio_id=@negocio_id AND status != 'excluido';

-- name: ListarUnidadesPorNegocio
SELECT id, negocio_id, nome, status, criado_em, atualizado_em, version
FROM unidade WHERE negocio_id=@negocio_id AND status != 'excluido'
ORDER BY criado_em DESC LIMIT @limit OFFSET @offset;
