ALTER TABLE template
    ADD COLUMN origem VARCHAR(50) NOT NULL DEFAULT 'sognix',
    ADD COLUMN cliente_id UUID REFERENCES cliente(id) ON DELETE CASCADE;

CREATE INDEX idx_template_origem ON template(origem);
CREATE INDEX idx_template_cliente ON template(cliente_id);

ALTER TABLE template
    ADD CONSTRAINT chk_template_origem_cliente
    CHECK (
        (origem = 'sognix' AND cliente_id IS NULL)
        OR (origem = 'cliente' AND cliente_id IS NOT NULL)
    );
