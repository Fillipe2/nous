-- name: SalvarCategoria
INSERT INTO categoria (id, nome, status, criado_em, atualizado_em, version)
VALUES (@id, @nome, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarCategoriaPorID
SELECT id, nome, status, criado_em, atualizado_em, version
FROM categoria WHERE id=@id AND status != 'excluido';

-- name: AtualizarCategoria
UPDATE categoria
SET nome=@nome, status=@status, atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteCategoria
UPDATE categoria SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarCategorias
SELECT COUNT(*) FROM categoria WHERE status != 'excluido';

-- name: ListarCategorias
SELECT id, nome, status, criado_em, atualizado_em, version
FROM categoria WHERE status != 'excluido'
ORDER BY nome LIMIT @limit OFFSET @offset;
