ALTER TABLE landingpage
    ADD COLUMN tipo VARCHAR(50) NOT NULL DEFAULT 'white_label';

ALTER TABLE template
    ADD COLUMN tipo VARCHAR(50) NOT NULL DEFAULT 'white_label',
    ADD COLUMN conteudo TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_landingpage_tipo ON landingpage(tipo);
CREATE INDEX idx_template_tipo ON template(tipo);
