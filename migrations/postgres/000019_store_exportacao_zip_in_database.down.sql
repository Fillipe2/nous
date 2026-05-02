ALTER TABLE exportacao
    ADD COLUMN arquivo_url TEXT NOT NULL DEFAULT '';

ALTER TABLE exportacao
    DROP COLUMN arquivo,
    DROP COLUMN tamanho_bytes,
    DROP COLUMN content_type,
    DROP COLUMN nome_arquivo;
