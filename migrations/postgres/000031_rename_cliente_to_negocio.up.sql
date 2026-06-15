-- Renomeia a entidade contratante `cliente` -> `negocio` e a antiga
-- sub-unidade `negocio` -> `unidade`, eliminando o overload de nomenclatura.
-- A ordem importa: primeiro libera o nome "negocio" (sub-unidade -> unidade),
-- depois promove `cliente` -> `negocio`.

-- 1. Sub-unidade: negocio -> unidade (libera o nome "negocio").
ALTER TABLE negocio RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE negocio RENAME TO unidade;
ALTER INDEX idx_negocio_cliente RENAME TO idx_unidade_negocio;

-- 2. landingpage: a FK para a sub-unidade vira unidade_id (antes de liberar negocio_id).
ALTER TABLE landingpage RENAME COLUMN negocio_id TO unidade_id;

-- 3. Remove o CHECK antigo de template antes dos renames de coluna/valor.
ALTER TABLE template DROP CONSTRAINT chk_template_origem_cliente;

-- 4. Entidade contratante: cliente -> negocio.
ALTER TABLE cliente RENAME TO negocio;

-- 5. Colunas cliente_id -> negocio_id nas tabelas que referenciam a entidade.
ALTER TABLE landingpage RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE template RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE faturamento_mensal RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE bi_metrica_diaria RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE lead RENAME COLUMN cliente_id TO negocio_id;

-- 6. Tabelas relacionadas: renomeia coluna e depois a tabela.
ALTER TABLE usuario_cliente RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE usuario_cliente RENAME TO usuario_negocio;

ALTER TABLE assinatura_cliente RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE assinatura_cliente RENAME TO assinatura_negocio;

ALTER TABLE uso_cliente_evento RENAME COLUMN cliente_id TO negocio_id;
ALTER TABLE uso_cliente_evento RENAME TO uso_negocio_evento;

-- 7. Índices que carregam o nome antigo.
ALTER INDEX idx_landingpage_cliente RENAME TO idx_landingpage_negocio;
ALTER INDEX idx_template_cliente RENAME TO idx_template_negocio;
ALTER INDEX idx_assinatura_cliente_cliente RENAME TO idx_assinatura_negocio_negocio;
ALTER INDEX idx_assinatura_cliente_status RENAME TO idx_assinatura_negocio_status;
ALTER INDEX idx_uso_cliente_evento_cliente RENAME TO idx_uso_negocio_evento_negocio;
ALTER INDEX idx_uso_cliente_evento_tipo RENAME TO idx_uso_negocio_evento_tipo;
ALTER INDEX idx_uso_cliente_evento_entidade RENAME TO idx_uso_negocio_evento_entidade;
ALTER INDEX idx_bi_metrica_diaria_cliente RENAME TO idx_bi_metrica_diaria_negocio;
ALTER INDEX idx_lead_cliente RENAME TO idx_lead_negocio;

-- 8. Valor de enum template.origem: 'cliente' -> 'negocio' e novo CHECK.
UPDATE template SET origem = 'negocio' WHERE origem = 'cliente';
ALTER TABLE template ADD CONSTRAINT chk_template_origem_negocio
    CHECK ((origem = 'sognix' AND negocio_id IS NULL)
        OR (origem = 'negocio' AND negocio_id IS NOT NULL));
