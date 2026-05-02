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

Instale o Goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Comandos:

```bash
goose -dir migrations/postgres postgres "$DATABASE_URL" up
goose -dir migrations/postgres postgres "$DATABASE_URL" status
goose -dir migrations/postgres postgres "$DATABASE_URL" down
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
- A API aplica headers básicos de segurança e rate limit em endpoints públicos sensíveis.
- Tráfego em produção deve usar HTTPS no gateway/load balancer.

## Alto Fluxo

Checklist antes de produção:

- Ajustar `pool_max_conns` conforme capacidade do PostgreSQL.
- Ativar rate limit distribuído no gateway/load balancer.
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
