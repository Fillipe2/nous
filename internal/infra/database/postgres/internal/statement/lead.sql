-- name: SalvarLead
INSERT INTO lead (id, landingpage_id, nome, email, telefone, whatsapp, mensagem,
    campos_personalizados, origem, ip, user_agent, status, criado_em)
VALUES (@id, @landingpage_id, @nome, @email, @telefone, @whatsapp, @mensagem,
    @campos_personalizados, @origem, @ip, @user_agent, @status, @criado_em);

-- name: BuscarLeadPorID
SELECT id, landingpage_id, nome, email, telefone, whatsapp, mensagem,
    campos_personalizados, origem, ip, user_agent, status, criado_em
FROM lead WHERE id=@id;

-- name: AtualizarStatusLead
UPDATE lead SET status=@status WHERE id=@id;
