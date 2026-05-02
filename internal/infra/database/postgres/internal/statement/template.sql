-- name: SalvarTemplate
INSERT INTO template (id, cliente_id, categoria_id, tipo, origem, nome, arquivo_url, conteudo, status, criado_em, atualizado_em, version)
VALUES (@id, @cliente_id, @categoria_id, @tipo, @origem, @nome, @arquivo_url, @conteudo, @status, @criado_em, @atualizado_em, @version);

-- name: BuscarTemplatePorID
SELECT id, cliente_id, categoria_id, tipo, origem, nome, arquivo_url, conteudo, status, criado_em, atualizado_em, version
FROM template WHERE id=@id AND status != 'excluido';

-- name: SoftDeleteTemplate
UPDATE template SET status='excluido', atualizado_em=@atualizado_em WHERE id=@id;

-- name: ContarTemplates
SELECT COUNT(*) FROM template WHERE status != 'excluido';

-- name: ListarTemplates
SELECT id, cliente_id, categoria_id, tipo, origem, nome, arquivo_url, conteudo, status, criado_em, atualizado_em, version
FROM template WHERE status != 'excluido'
ORDER BY nome LIMIT @limit OFFSET @offset;

-- name: ContarTemplatesPorCategoria
SELECT COUNT(*) FROM template WHERE categoria_id=@categoria_id AND status != 'excluido';

-- name: ListarTemplatesPorCategoria
SELECT id, cliente_id, categoria_id, tipo, origem, nome, arquivo_url, conteudo, status, criado_em, atualizado_em, version
FROM template WHERE categoria_id=@categoria_id AND status != 'excluido'
ORDER BY nome LIMIT @limit OFFSET @offset;
