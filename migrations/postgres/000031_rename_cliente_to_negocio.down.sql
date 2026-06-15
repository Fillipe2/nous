-- Reverte a renomeação cliente->negocio / negocio->unidade, na ordem inversa.

-- 8. Reverte CHECK e valor de template.origem.
ALTER TABLE template DROP CONSTRAINT chk_template_origem_negocio;
UPDATE template SET origem = 'cliente' WHERE origem = 'negocio';

-- 7. Reverte nomes de índices.
ALTER INDEX idx_lead_negocio RENAME TO idx_lead_cliente;
ALTER INDEX idx_bi_metrica_diaria_negocio RENAME TO idx_bi_metrica_diaria_cliente;
ALTER INDEX idx_uso_negocio_evento_entidade RENAME TO idx_uso_cliente_evento_entidade;
ALTER INDEX idx_uso_negocio_evento_tipo RENAME TO idx_uso_cliente_evento_tipo;
ALTER INDEX idx_uso_negocio_evento_negocio RENAME TO idx_uso_cliente_evento_cliente;
ALTER INDEX idx_assinatura_negocio_status RENAME TO idx_assinatura_cliente_status;
ALTER INDEX idx_assinatura_negocio_negocio RENAME TO idx_assinatura_cliente_cliente;
ALTER INDEX idx_template_negocio RENAME TO idx_template_cliente;
ALTER INDEX idx_landingpage_negocio RENAME TO idx_landingpage_cliente;

-- 6. Reverte tabelas relacionadas.
ALTER TABLE uso_negocio_evento RENAME TO uso_cliente_evento;
ALTER TABLE uso_cliente_evento RENAME COLUMN negocio_id TO cliente_id;

ALTER TABLE assinatura_negocio RENAME TO assinatura_cliente;
ALTER TABLE assinatura_cliente RENAME COLUMN negocio_id TO cliente_id;

ALTER TABLE usuario_negocio RENAME TO usuario_cliente;
ALTER TABLE usuario_cliente RENAME COLUMN negocio_id TO cliente_id;

-- 5. Reverte colunas negocio_id -> cliente_id.
ALTER TABLE lead RENAME COLUMN negocio_id TO cliente_id;
ALTER TABLE bi_metrica_diaria RENAME COLUMN negocio_id TO cliente_id;
ALTER TABLE faturamento_mensal RENAME COLUMN negocio_id TO cliente_id;
ALTER TABLE template RENAME COLUMN negocio_id TO cliente_id;
ALTER TABLE landingpage RENAME COLUMN negocio_id TO cliente_id;

-- 4. Reverte entidade contratante: negocio -> cliente.
ALTER TABLE negocio RENAME TO cliente;

-- 3. Re-adiciona o CHECK antigo de template.
ALTER TABLE template ADD CONSTRAINT chk_template_origem_cliente
    CHECK ((origem = 'sognix' AND cliente_id IS NULL)
        OR (origem = 'cliente' AND cliente_id IS NOT NULL));

-- 2. Reverte landingpage.unidade_id -> negocio_id.
ALTER TABLE landingpage RENAME COLUMN unidade_id TO negocio_id;

-- 1. Reverte sub-unidade: unidade -> negocio.
ALTER INDEX idx_unidade_negocio RENAME TO idx_negocio_cliente;
ALTER TABLE unidade RENAME TO negocio;
ALTER TABLE negocio RENAME COLUMN negocio_id TO cliente_id;
