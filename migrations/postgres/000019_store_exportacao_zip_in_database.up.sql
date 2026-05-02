ALTER TABLE exportacao
    ADD COLUMN nome_arquivo VARCHAR(255) NOT NULL DEFAULT 'exportacao.zip',
    ADD COLUMN content_type VARCHAR(100) NOT NULL DEFAULT 'application/zip',
    ADD COLUMN tamanho_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN arquivo BYTEA NOT NULL DEFAULT ''::bytea;

ALTER TABLE exportacao
    DROP COLUMN arquivo_url;
