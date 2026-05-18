-- name: SalvarEventoOutbox
INSERT INTO evento_outbox (
    event_id, event_type, schema_version, source, occurred_at,
    idempotency_key, payload, status, tentativas, criado_em
)
VALUES (
    @event_id, @event_type, @schema_version, @source, @occurred_at,
    @idempotency_key, @payload, 'pendente', 0, NOW()
)
ON CONFLICT (idempotency_key) DO NOTHING;

-- name: MarcarEventoOutboxPublicado
UPDATE evento_outbox
SET status='publicado', publicado_em=NOW(), atualizado_em=NOW(), mensagem_erro=NULL
WHERE event_id=@event_id;

-- name: MarcarEventoOutboxFalha
UPDATE evento_outbox
SET status='falha', tentativas=tentativas + 1, atualizado_em=NOW(), mensagem_erro=@mensagem_erro
WHERE event_id=@event_id;

-- name: BuscarEventosOutboxPendentes
SELECT event_id, event_type, schema_version, source, occurred_at, idempotency_key, payload
FROM evento_outbox
WHERE (status='falha' AND tentativas < 5)
   OR (status='pendente' AND criado_em < NOW() - INTERVAL '2 minutes')
ORDER BY criado_em ASC
LIMIT @limite;
