DROP INDEX IF EXISTS idx_template_tipo;
DROP INDEX IF EXISTS idx_landingpage_tipo;

ALTER TABLE template
    DROP COLUMN IF EXISTS conteudo,
    DROP COLUMN IF EXISTS tipo;

ALTER TABLE landingpage
    DROP COLUMN IF EXISTS tipo;
