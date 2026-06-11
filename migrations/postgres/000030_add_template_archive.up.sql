-- Armazena o pacote ZIP de templates white label diretamente no banco (BYTEA),
-- espelhando o modelo já usado em exportacao. A coluna conteudo (TEXT) continua
-- sendo usada para templates baseados em HTML/texto.
ALTER TABLE template
    ADD COLUMN arquivo BYTEA NOT NULL DEFAULT ''::bytea,
    ADD COLUMN nome_arquivo VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN content_type VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN tamanho_bytes BIGINT NOT NULL DEFAULT 0;
