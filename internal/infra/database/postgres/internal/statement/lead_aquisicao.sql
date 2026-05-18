-- name: SalvarLeadAquisicao
INSERT INTO lead_aquisicao (
    id, nome, email, telefone,
    categoria_pretendida_id, categoria_pretendida_livre,
    cidade, estado, porte, interesses, mensagem, origem,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    status_funil, convertido_em_usuario_id,
    ip, user_agent, criado_em, atualizado_em
) VALUES (
    @id, @nome, @email, @telefone,
    @categoria_pretendida_id, @categoria_pretendida_livre,
    @cidade, @estado, @porte, @interesses, @mensagem, @origem,
    @utm_source, @utm_medium, @utm_campaign, @utm_content, @utm_term,
    @status_funil, @convertido_em_usuario_id,
    @ip, @user_agent, @criado_em, @atualizado_em
);

-- name: BuscarLeadAquisicaoPorID
SELECT id, nome, email, telefone,
    categoria_pretendida_id, categoria_pretendida_livre,
    cidade, estado, porte, interesses, mensagem, origem,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    status_funil, convertido_em_usuario_id,
    ip, user_agent, criado_em, atualizado_em
FROM lead_aquisicao WHERE id=@id;

-- name: BuscarLeadAquisicaoPorEmail
SELECT id, nome, email, telefone,
    categoria_pretendida_id, categoria_pretendida_livre,
    cidade, estado, porte, interesses, mensagem, origem,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    status_funil, convertido_em_usuario_id,
    ip, user_agent, criado_em, atualizado_em
FROM lead_aquisicao WHERE email=@email;

-- name: AtualizarFunilLeadAquisicao
UPDATE lead_aquisicao
SET status_funil=@status_funil,
    convertido_em_usuario_id=@convertido_em_usuario_id,
    atualizado_em=@atualizado_em
WHERE id=@id;

