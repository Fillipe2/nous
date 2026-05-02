-- name: SalvarNegocio
INSERT INTO negocio (id, cliente_id, nome, status, criado_em, atualizado_em, version)
VALUES (@id, @cliente_id, @nome, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarNegocioPorID
SELECT id, cliente_id, nome, status, criado_em, atualizado_em, version
FROM negocio WHERE id=@id AND status != 'excluido';

-- name: AtualizarNegocio
UPDATE negocio
SET nome=@nome, status=@status, atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteNegocio
UPDATE negocio SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarNegociosPorCliente
SELECT COUNT(*) FROM negocio WHERE cliente_id=@cliente_id AND status != 'excluido';

-- name: ListarNegociosPorCliente
SELECT id, cliente_id, nome, status, criado_em, atualizado_em, version
FROM negocio WHERE cliente_id=@cliente_id AND status != 'excluido'
ORDER BY criado_em DESC LIMIT @limit OFFSET @offset;
