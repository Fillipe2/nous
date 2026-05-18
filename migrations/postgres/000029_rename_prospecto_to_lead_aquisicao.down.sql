ALTER TABLE lead_aquisicao RENAME TO prospecto;

ALTER INDEX idx_lead_aquisicao_email RENAME TO idx_prospecto_email;
ALTER INDEX idx_lead_aquisicao_status_funil RENAME TO idx_prospecto_status_funil;
ALTER INDEX idx_lead_aquisicao_criado_em RENAME TO idx_prospecto_criado_em;
ALTER INDEX idx_lead_aquisicao_utm_campaign RENAME TO idx_prospecto_utm_campaign;
