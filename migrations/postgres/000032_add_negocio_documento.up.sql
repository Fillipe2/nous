-- Documento (CPF ou CNPJ) do negócio. Opcional e nullable.
-- VARCHAR(14): CPF tem 11 dígitos; CNPJ tem 14 posições, alfanuméricas a partir
-- de 2026. Armazenado sem separadores, em maiúsculas, preservando zeros à esquerda.
ALTER TABLE negocio ADD COLUMN documento VARCHAR(14);
