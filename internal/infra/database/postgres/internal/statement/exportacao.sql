-- name: SalvarExportacao
INSERT INTO exportacao (id, landingpage_id, usuario_id, nome_arquivo, content_type, tamanho_bytes, arquivo, hash, criado_em)
VALUES (@id, @landingpage_id, @usuario_id, @nome_arquivo, @content_type, @tamanho_bytes, @arquivo, @hash, @criado_em);

-- name: BuscarExportacaoPorID
SELECT id, landingpage_id, usuario_id, nome_arquivo, content_type, tamanho_bytes, arquivo, hash, criado_em
FROM exportacao WHERE id=@id;

-- name: ListarExportacoesPorLandingPage
SELECT id, landingpage_id, usuario_id, nome_arquivo, content_type, tamanho_bytes, ''::bytea AS arquivo, hash, criado_em
FROM exportacao WHERE landingpage_id=@landingpage_id
ORDER BY criado_em DESC;

-- name: BuscarMaisRecenteExportacaoPorLandingPage
SELECT id, landingpage_id, usuario_id, nome_arquivo, content_type, tamanho_bytes, arquivo, hash, criado_em
FROM exportacao WHERE landingpage_id=@landingpage_id
ORDER BY criado_em DESC LIMIT 1;
