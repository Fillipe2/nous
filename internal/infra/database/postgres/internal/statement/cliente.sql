-- name: SalvarCliente
INSERT INTO cliente (id, nome, email, telefone, whatsapp, status, criado_em, atualizado_em, version)
VALUES (@id, @nome, @email, @telefone, @whatsapp, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarClientePorID
SELECT id, nome, email, telefone, whatsapp, status, criado_em, atualizado_em, version
FROM cliente WHERE id=@id AND status != 'excluido';

-- name: AtualizarCliente
UPDATE cliente
SET nome=@nome, email=@email, telefone=@telefone, whatsapp=@whatsapp, status=@status, atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteCliente
UPDATE cliente SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarClientes
SELECT COUNT(*) FROM cliente WHERE status != 'excluido';

-- name: ListarClientes
SELECT id, nome, email, telefone, whatsapp, status, criado_em, atualizado_em, version
FROM cliente WHERE status != 'excluido'
ORDER BY criado_em DESC LIMIT @limit OFFSET @offset;

-- name: VincularUsuarioCliente
INSERT INTO usuario_cliente (usuario_id, cliente_id, criado_em)
VALUES (@usuario_id, @cliente_id, @criado_em)
ON CONFLICT DO NOTHING;

-- name: ListarClientesPorUsuario
SELECT c.id, c.nome, c.email, c.telefone, c.whatsapp, c.status, c.criado_em, c.atualizado_em, c.version
FROM cliente c
INNER JOIN usuario_cliente uc ON uc.cliente_id = c.id
WHERE uc.usuario_id=@usuario_id AND c.status != 'excluido'
ORDER BY c.nome;

-- name: VerificarAcessoCliente
SELECT EXISTS (
		SELECT 1
		FROM usuario_cliente uc
		INNER JOIN cliente c ON c.id = uc.cliente_id
		WHERE uc.cliente_id=@cliente_id
			AND uc.usuario_id=@usuario_id
			AND c.status != 'excluido'
) AS acesso;
