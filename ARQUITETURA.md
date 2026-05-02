# Arquitetura Buildnix Backend

## Visão Geral

Buildnix é o backend de criação, edição, exportação e gestão de landing pages whitelabel da Sognix. Ele deve subir sozinho como produto comercializável e ativar integrações externas somente por configuração.

## Camadas

```text
Interface -> App -> Domain <- Infra
```

- `internal/domain`: entidades, erros sentinela e interfaces de repositório. Não importa app, infra, interface, banco ou HTTP.
- `internal/app`: DTOs e casos de uso. Orquestra domínio por interfaces.
- `internal/infra`: implementações de PostgreSQL, JWT e eventos. Implementa interfaces do domínio.
- `internal/interface/webapi`: servidor HTTP, handlers, middlewares, router e OpenAPI.
- `cmd/internal/bootstrap`: montagem explícita de configuração, infraestrutura, repositórios e casos de uso.
- `cmd/webapi`: entrypoint da API REST.

## API HTTP

A API principal fica em `internal/interface/webapi`, iniciada por `cmd/webapi/main.go`.

Não existe entrypoint `webhook`, `worker` ou `cli` enquanto o domínio não exigir esse fluxo. O Buildnix recebe leads por rota pública HTTP e envia eventos para o Observer de forma opcional.

OpenAPI:

- Arquivo embarcado: `internal/interface/webapi/internal/router/openapi.yaml`
- Endpoint: `GET /swagger/openapi.yaml`

## Banco de Dados

PostgreSQL com `pgxpool`.

SQL segue o padrão:

- Arquivos `.sql` em `internal/infra/database/postgres/internal/statement/`.
- Statements nomeados com `-- name: NomeQuery`.
- Parâmetros nomeados com `@param_name`.
- Parse em runtime via `github.com/phenpessoa/sql2go` e `//go:embed`.
- Scanners manuais em `internal/infra/database/postgres/internal/scanner/`.
- Repositórios em `internal/infra/database/postgres/*_repository.go`.

Não usar `sqlc` neste projeto.

## Migrations

As migrations ficam em `migrations/postgres/`, separadas por recurso/tabela, com arquivos `.up.sql` e `.down.sql`.

Entidades transacionais usam, quando aplicável:

- `id`
- `status`
- `criado_em`
- `atualizado_em`
- `version`

## Concorrência

O serviço é stateless. A proteção de concorrência fica no PostgreSQL, não em mutex local, para funcionar com múltiplas instâncias.

Controle otimista:

- Entidades editáveis usam `version`.
- Updates usam `WHERE id=@id AND version=@version_anterior` ou equivalente.
- Se nenhuma linha for afetada, a API retorna HTTP 409.
- O frontend deve recarregar os dados ao receber 409.

Entidades protegidas por `version`:

- `usuario`
- `cliente`
- `negocio`
- `categoria`
- `landingpage`
- autosave de `landingpage`

Controle pessimista:

- Landing pages têm `locked_by` e `locked_at`.
- Lock expira em 5 minutos.
- Aquisição do lock é atômica no banco.

## Alto Fluxo

Proteções atuais:

- `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` e `MaxHeaderBytes` no servidor HTTP.
- Limite de corpo na borda HTTP: 1 MiB para JSON e 60 MiB para multipart.
- `pgxpool` para conexões PostgreSQL.
- Outbox com `idempotency_key` único para eventos.
- Auditoria não quebra o fluxo se falhar ao salvar.

Recomendações para produção:

- Ajustar `pool_max_conns`, `pool_min_conns` e ciclo de vida de conexão na `DATABASE_URL` conforme capacidade do banco.
- Usar rate limiting distribuído no gateway/load balancer ou middleware com backend compartilhado.
- Rodar teste de carga em `POST /leads`, login, autosave, exportação e listagens.
- Evoluir auditoria para fila/outbox dedicada se o volume de escrita por request crescer.
- Monitorar latência HTTP, saturação do pool, locks, erros 409, filas de outbox e tempo de queries.

## Eventos e Observer

Eventos enviados ao Observer usam envelope versionado:

```json
{
  "event_id": "uuid",
  "event_type": "lead.capturado",
  "schema_version": 1,
  "source": "buildnix",
  "occurred_at": "2026-05-01T12:00:00Z",
  "idempotency_key": "buildnix:lead:uuid",
  "payload": {}
}
```

Sem `OBSERVER_URL`, o produto opera standalone. Com `OBSERVER_URL`, os eventos são persistidos no outbox local e publicados no Observer.

## Leads

Leads são dados comerciais da Sognix/admin. Cliente comum não recebe lista nominal nem métricas de leads no dashboard.

Compartilhamento, venda ou repasse de lead deve virar fluxo próprio com consentimento/base legal, contrato e trilha de auditoria.

## Exportação

Exportações de landing page são ZIPs salvos em PostgreSQL na tabela `exportacao`, incluindo:

- nome do arquivo
- content type
- tamanho em bytes
- arquivo em `BYTEA`
- hash SHA-256

## Auditoria, Cobrança e BI

- `auditoria`: trilha local para segurança, LGPD e investigação.
- `uso_cliente_evento`: eventos append-only de uso faturável.
- `faturamento_mensal`: fechamento mensal por competência.
- `bi_metrica_diaria`: agregações para visão Sognix/admin, sem pesar tabelas transacionais.

Auditoria não deve ser usada sozinha como base de cobrança.
