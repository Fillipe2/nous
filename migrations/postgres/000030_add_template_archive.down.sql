ALTER TABLE template
    DROP COLUMN IF EXISTS arquivo,
    DROP COLUMN IF EXISTS nome_arquivo,
    DROP COLUMN IF EXISTS content_type,
    DROP COLUMN IF EXISTS tamanho_bytes;
