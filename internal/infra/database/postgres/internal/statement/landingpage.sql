-- name: SalvarLandingPage
INSERT INTO landingpage (id, cliente_id, negocio_id, categoria_id, template_id, tipo, nome, conteudo,
    status, is_draft, autosave_data, criado_em, atualizado_em, version)
VALUES (@id, @cliente_id, @negocio_id, @categoria_id, @template_id, @tipo, @nome, @conteudo,
    @status, @is_draft, @autosave_data, @criado_em, @atualizado_em, @version);

-- name: BuscarLandingPagePorID
SELECT id, cliente_id, negocio_id, categoria_id, template_id, tipo,
    nome, conteudo, status, is_draft, autosave_data,
    last_autosave_em, locked_by, locked_at, criado_em, atualizado_em, version
FROM landingpage WHERE id=@id AND status != 'excluida';

-- name: AtualizarLandingPage
UPDATE landingpage
SET negocio_id=@negocio_id, categoria_id=@categoria_id, template_id=@template_id, tipo=@tipo,
    nome=@nome, conteudo=@conteudo, status=@status, is_draft=@is_draft,
    atualizado_em=@atualizado_em, version=@version
WHERE id=@id AND version=@version_anterior;

-- name: SoftDeleteLandingPage
UPDATE landingpage SET status='excluida', atualizado_em=@atualizado_em WHERE id=@id;

-- name: AtualizarAutosaveLandingPage
UPDATE landingpage
SET autosave_data=@autosave_data, last_autosave_em=@last_autosave_em, is_draft=true, version=version + 1
WHERE id=@id AND version=@version
RETURNING version;

-- name: BloquearLandingPage
UPDATE landingpage
SET locked_by=@locked_by::uuid, locked_at=@locked_at
WHERE id=@id
    AND (locked_by IS NULL OR locked_by=@locked_by::uuid OR locked_at < @expirado_em);

-- name: DesbloquearLandingPage
UPDATE landingpage SET locked_by=NULL, locked_at=NULL WHERE id=@id;

-- name: VerificarAcessoLandingPage
SELECT EXISTS (
    SELECT 1 FROM landingpage lp
    JOIN usuario_cliente uc ON uc.cliente_id = lp.cliente_id
    WHERE lp.id = @id AND uc.usuario_id = @usuario_id AND lp.status != 'excluida'
) AS acesso;
