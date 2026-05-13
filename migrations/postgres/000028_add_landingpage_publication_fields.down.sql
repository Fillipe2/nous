DROP INDEX IF EXISTS idx_landingpage_slug;
ALTER TABLE landingpage DROP COLUMN IF EXISTS publicada_em;
ALTER TABLE landingpage DROP COLUMN IF EXISTS deploy_id;
ALTER TABLE landingpage DROP COLUMN IF EXISTS url_publica;
ALTER TABLE landingpage DROP COLUMN IF EXISTS slug;
