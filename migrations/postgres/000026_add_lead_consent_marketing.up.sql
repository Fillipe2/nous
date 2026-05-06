ALTER TABLE lead
    ADD COLUMN consentimento_marketing_parceiros BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN consentimento_versao_marketing VARCHAR(50);
