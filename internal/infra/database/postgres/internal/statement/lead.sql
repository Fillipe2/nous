-- name: SalvarLead
INSERT INTO lead (id, landingpage_id, negocio_id, nome, email, telefone, whatsapp, mensagem,
    campos_personalizados, origem, ip, user_agent, referrer,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    consentimento, consentimento_em, politica_versao,
    consentimento_marketing_parceiros, consentimento_versao_marketing,
    status, criado_em, atualizado_em)
VALUES (@id, @landingpage_id, @negocio_id, @nome, @email, @telefone, @whatsapp, @mensagem,
    @campos_personalizados, @origem, @ip, @user_agent, @referrer,
    @utm_source, @utm_medium, @utm_campaign, @utm_content, @utm_term,
    @consentimento, @consentimento_em, @politica_versao,
    @consentimento_marketing_parceiros, @consentimento_versao_marketing,
    @status, @criado_em, @atualizado_em);

-- name: BuscarLeadPorID
SELECT id, landingpage_id, negocio_id, nome, email, telefone, whatsapp, mensagem,
    campos_personalizados, origem, ip, user_agent, referrer,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    consentimento, consentimento_em, politica_versao,
    consentimento_marketing_parceiros, consentimento_versao_marketing,
    status, perdido_motivo, anonimizado_em, criado_em, atualizado_em
FROM lead WHERE id=@id;

-- name: BuscarLeadDuplicado
SELECT id, landingpage_id, negocio_id, nome, email, telefone, whatsapp, mensagem,
    campos_personalizados, origem, ip, user_agent, referrer,
    utm_source, utm_medium, utm_campaign, utm_content, utm_term,
    consentimento, consentimento_em, politica_versao,
    consentimento_marketing_parceiros, consentimento_versao_marketing,
    status, perdido_motivo, anonimizado_em, criado_em, atualizado_em
FROM lead
WHERE email=@email AND landingpage_id=@landingpage_id AND criado_em >= @desde
    AND anonimizado_em IS NULL
ORDER BY criado_em DESC
LIMIT 1;

-- name: AtualizarStatusLead
UPDATE lead SET status=@status, perdido_motivo=@perdido_motivo, atualizado_em=@atualizado_em WHERE id=@id;

-- name: AnonimizarLead
UPDATE lead
SET nome='Anonimo',
    email='anon_' || id::text || '@anonimizado.local',
    telefone=NULL,
    whatsapp=NULL,
    mensagem=NULL,
    campos_personalizados=NULL,
    ip=NULL,
    user_agent=NULL,
    referrer=NULL,
    anonimizado_em=@agora,
    atualizado_em=@agora
WHERE id=@id AND anonimizado_em IS NULL;

-- name: AnonimizarLeadsExpirados
UPDATE lead
SET nome='Anonimo',
    email='anon_' || id::text || '@anonimizado.local',
    telefone=NULL,
    whatsapp=NULL,
    mensagem=NULL,
    campos_personalizados=NULL,
    ip=NULL,
    user_agent=NULL,
    referrer=NULL,
    anonimizado_em=@agora,
    atualizado_em=@agora
WHERE anonimizado_em IS NULL
    AND criado_em < @antes_de
    AND status NOT IN ('convertido');
