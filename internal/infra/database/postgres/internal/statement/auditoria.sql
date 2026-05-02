-- name: SalvarAuditoria
INSERT INTO auditoria (id, usuario_id, acao, entidade, entidade_id, dados_anteriores, dados_novos, ip, user_agent, status_code, tempo_resposta_ms, request_id, criado_em)
VALUES (@id, @usuario_id, @acao, @entidade, @entidade_id, @dados_anteriores, @dados_novos, @ip, @user_agent, @status_code, @tempo_resposta_ms, @request_id, @criado_em);

-- name: BuscarAuditoriaPorID
SELECT id, usuario_id, acao, entidade, entidade_id, dados_anteriores, dados_novos, ip, user_agent, status_code, tempo_resposta_ms, request_id, criado_em
FROM auditoria WHERE id=@id;
