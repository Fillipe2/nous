# Observabilidade Buildnix Backend

Este runbook define os sinais mínimos para acompanhar o Buildnix Backend em produção. A stack pode ser CloudWatch, Grafana/Prometheus, Datadog ou equivalente; os painéis abaixo são independentes do fornecedor.

## Dashboards mínimos

### API HTTP

- Taxa de requests por rota, método e status.
- Latência p50, p95 e p99 por rota.
- Erros 4xx e 5xx por rota.
- `429 Too Many Requests` por rota sensível.
- Tamanho de payload recusado por limite de body.

### Banco PostgreSQL

- Conexões ativas, ociosas e em espera.
- Saturação do pool de conexões da aplicação.
- Query duration p95/p99.
- Locks aguardando mais de 5 segundos.
- Deadlocks.
- Uso de CPU, memória e disco.
- Crescimento das tabelas `auditoria`, `lead`, `exportacao`, `evento_outbox` e `usuario_refresh_token`.

### Segurança e autenticação

- Falhas de login por IP e email.
- Refresh tokens revogados, expirados e expurgados.
- Tentativas de OAuth com state inválido ou ausente.
- Origens CORS rejeitadas.
- Falhas de validação de ZIP/exportação.

### Redis rate limit

- Disponibilidade do Redis.
- Erros de execução do script Lua.
- Fallback para rate limit em memória.
- Reconexões bem-sucedidas após fallback.
- Latência de comandos Redis.

### Jobs locais

- Última execução do cleanup de refresh tokens.
- Quantidade de tokens expurgados.
- Falhas do cleanup.

### Outbox e integrações

- Eventos pendentes em `evento_outbox`.
- Eventos publicados com sucesso.
- Eventos com falha por tipo.
- Tempo máximo de permanência na outbox.

## Alertas recomendados

- 5xx maior que 2% por 5 minutos.
- p95 de latência HTTP acima de 1 segundo por 10 minutos.
- Qualquer rota pública sensível com pico anormal de 401, 403 ou 429.
- Redis indisponível por mais de 2 minutos quando `REDIS_URL` estiver configurada.
- Pool PostgreSQL saturado por mais de 5 minutos.
- Deadlock PostgreSQL maior que zero.
- Locks PostgreSQL acima de 30 segundos.
- Outbox com eventos pendentes por mais de 15 minutos.
- Falha no cleanup de refresh tokens.
- Uso de disco PostgreSQL acima de 80%.

## Logs estruturados esperados

Os logs devem permitir filtrar por:

- `request_id`.
- rota e método HTTP.
- status HTTP.
- IP do cliente já sanitizado.
- usuário autenticado quando existir.
- duração da request.
- erro interno sem expor segredo ou dado pessoal sensível.

## Checklist de release

- Confirmar dashboard da API HTTP ativo.
- Confirmar dashboard PostgreSQL ativo.
- Confirmar dashboard Redis ativo quando `REDIS_URL` estiver configurada.
- Confirmar alertas de 5xx, latência, pool, Redis e outbox.
- Confirmar retenção e mascaramento de logs conforme LGPD.
- Confirmar que `.env`, tokens, dumps, logs e artefatos sensíveis não são coletados como artefatos públicos.
