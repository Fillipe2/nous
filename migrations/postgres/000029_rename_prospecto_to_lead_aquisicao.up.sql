ALTER TABLE prospecto RENAME TO lead_aquisicao;

ALTER INDEX idx_prospecto_email RENAME TO idx_lead_aquisicao_email;
ALTER INDEX idx_prospecto_status_funil RENAME TO idx_lead_aquisicao_status_funil;
ALTER INDEX idx_prospecto_criado_em RENAME TO idx_lead_aquisicao_criado_em;
ALTER INDEX idx_prospecto_utm_campaign RENAME TO idx_lead_aquisicao_utm_campaign;
