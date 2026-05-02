# Operação Buildnix Backend

## Ambiente

Crie o `.env` a partir do exemplo:

```bash
cp .env.example .env
```

Segredos obrigatórios em produção:

- `JWT_SECRET`
- `JWT_REFRESH_SECRET`
- `ADMIN_SENHA`, quando o seed de admin estiver ativo
- `GOOGLE_CLIENT_SECRET`, quando OAuth2 estiver ativo

`JWT_SECRET` e `JWT_REFRESH_SECRET` são obrigatórios, precisam ter pelo menos 32 caracteres e devem ser diferentes. O bootstrap falha se esses valores estiverem vazios ou com placeholders conhecidos.

Para maior carga, ajuste o pool na `DATABASE_URL`:

```text
postgres://buildnix:buildnix@localhost:5432/buildnix?sslmode=disable&pool_max_conns=20&pool_min_conns=2
```

## Admin Inicial

O seed de admin roda no bootstrap somente quando `ADMIN_EMAIL` e `ADMIN_SENHA` estão preenchidos.

Se o usuário já existir, o bootstrap deve promover/reativar o admin conforme a regra da aplicação.

## Migrations

As migrations usam o formato `golang-migrate` (`.up.sql` e `.down.sql`). Com Docker, nao e necessario instalar a CLI localmente.

```bash
docker compose up -d db
```

Comandos:

```bash
docker run --rm --network buildnix_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://buildnix:buildnix@db:5432/buildnix?sslmode=disable" up
docker run --rm --network buildnix_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://buildnix:buildnix@db:5432/buildnix?sslmode=disable" version
docker run --rm --network buildnix_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://buildnix:buildnix@db:5432/buildnix?sslmode=disable" down 1
```

Com `make` instalado, os atalhos equivalentes sao:

```bash
make migrate-up
make migrate-status
make migrate-down
```

## Testes e Build

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go build -o bin/buildnix-backend-webapi ./cmd/webapi
```

O detector de race exige GCC instalado.

## Docker Local

```bash
docker compose up --build
```

Depois que o PostgreSQL estiver pronto, rode as migrations.

A imagem da API usa base Alpine versionada, usuário não-root e healthcheck HTTP em `/health`. O `HEALTHCHECK` usa o `wget` do BusyBox disponível no Alpine; se a base mudar para distroless ou scratch, o comando do healthcheck deve ser substituído.

## Logs e Auditoria

Logs operacionais:

- Arquivo padrão: `./logs/buildnix-backend.log`.
- No Docker Compose, o arquivo fica em `/app/logs/buildnix-backend.log` dentro do volume `buildnix_backend_logs`.
- A aplicação também escreve no stdout quando `LOG_STDOUT=true`.
- Rotação por tamanho: `LOG_MAX_SIZE_MB`.
- Ciclo de vida/retenção: `LOG_MAX_AGE_DAYS` e `LOG_MAX_BACKUPS`.
- Compressão dos arquivos rotacionados: `LOG_COMPRESS=true`.

Variáveis disponíveis:

```dotenv
LOG_FILE_PATH=./logs/buildnix-backend.log
LOG_MAX_SIZE_MB=50
LOG_MAX_AGE_DAYS=30
LOG_MAX_BACKUPS=10
LOG_COMPRESS=true
LOG_STDOUT=true
```

Auditoria:

- Persistida em PostgreSQL na tabela `auditoria` para rotas autenticadas.
- Registra usuário, ação, IP, user agent, status HTTP, tempo de resposta e request id.
- Falha ao salvar auditoria não quebra o fluxo da requisição.
- Retenção recomendada: 5 anos.

## LGPD e Segurança

- Captura de lead exige consentimento explícito e rastreabilidade de origem.
- Lead não convertido deve ser anonimizado após 12 meses sem base legal ativa.
- Solicitação do titular deve aplicar soft delete e anonimização quando não houver obrigação legal de retenção.
- Auditoria pode manter referência mínima mascarada por 5 anos.
- Senhas são armazenadas com bcrypt.
- Autorização usa RBAC e vínculo usuário-cliente.
- Refresh tokens são armazenados como hash, rotacionados no refresh e revogados no logout quando enviados no payload.
- Refresh tokens expirados ou revogados há mais de 24 horas são expurgados por job local iniciado junto com a API.
- A API aplica headers básicos de segurança e rate limit em endpoints públicos sensíveis.
- Tráfego em produção deve usar HTTPS no gateway/load balancer.
- Sem `REDIS_URL`, o rate limit usa memória local e é adequado para desenvolvimento e instância única. Com `REDIS_URL`, o rate limit usa Redis e funciona de forma compartilhada entre múltiplas instâncias. Se o Redis estiver indisponível, a API cai temporariamente para memória local e tenta reconectar periodicamente.
- `OAUTH_STATE_SAMESITE` aceita `strict`, `lax` ou `none`. O padrão `strict` é o mais restritivo; use `lax` apenas se o provedor OAuth2 do ambiente exigir compatibilidade no redirect. O modo `none` só é aplicado quando `GOOGLE_REDIRECT_URL` usa HTTPS; caso contrário, a API preserva `strict`.

## Alto Fluxo

Checklist antes de produção:

- Ajustar `pool_max_conns` conforme capacidade do PostgreSQL.
- Ativar rate limit distribuído no gateway/load balancer.
- Rodar `make test-integration` com `TEST_DATABASE_URL` apontando para um PostgreSQL descartável antes de promover migrations/repositórios.
- Garantir que o workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) esteja obrigatório na proteção da branch.
- Configurar dashboards e alertas mínimos conforme [docs/OBSERVABILIDADE.md](docs/OBSERVABILIDADE.md).
- Rodar teste de carga em login, `POST /leads`, autosave, exportação e listagens.
- Monitorar erros 409 para conflito de edição.
- Monitorar latência, saturação do pool, locks e queries lentas.
- Planejar auditoria assíncrona se o volume de escrita por request crescer.
- Implementar worker de retry/backoff para eventos pendentes/falhos no outbox.

## Comandos Úteis

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go build -o bin/buildnix-backend-webapi ./cmd/webapi
go run ./cmd/webapi
```
