ALTER TABLE template DROP CONSTRAINT IF EXISTS chk_template_origem_cliente;

DROP INDEX IF EXISTS idx_template_cliente;
DROP INDEX IF EXISTS idx_template_origem;

ALTER TABLE template
    DROP COLUMN IF EXISTS cliente_id,
    DROP COLUMN IF EXISTS origem;
