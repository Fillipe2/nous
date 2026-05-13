ALTER TABLE landingpage ADD COLUMN slug TEXT;
ALTER TABLE landingpage ADD COLUMN url_publica TEXT;
ALTER TABLE landingpage ADD COLUMN deploy_id TEXT;
ALTER TABLE landingpage ADD COLUMN publicada_em TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_landingpage_slug ON landingpage(slug) WHERE slug IS NOT NULL;
