# Operação Nous

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
postgres://nous:nous@localhost:5432/nous?sslmode=disable&pool_max_conns=20&pool_min_conns=2
```

## Administradores

O modelo é de **múltiplos administradores gerenciáveis**:

- **Bootstrap:** o seed cria o primeiro admin quando `ADMIN_EMAIL` e `ADMIN_SENHA` estão preenchidos. Se o usuário já existir, é promovido a `admin`, reativado e marcado como `email_verified`.
- **Gestão em runtime:** um admin promove ou rebaixa outros usuários via `PATCH /usuarios/{id}/role` (exclusivo para admin). Há trava de segurança que **impede rebaixar o último administrador ativo** (retorna 409), evitando lockout.
- O cadastro público `POST /usuarios` cria **sempre** papel `usuario` — não é possível criar admin por essa rota.
- Admin também ativa/inativa usuários via `PATCH /usuarios/{id}/status` (mesma trava do último admin); a exclusão usa `DELETE /usuarios/{id}` (soft delete).
- `email` é único entre usuários **não excluídos** (índice parcial): o email de um usuário excluído pode ser reutilizado num novo cadastro. Email já em uso retorna 409.
- Verificação de email: `email_verified` vira true em login Google e no admin do seed. Para cadastro local há `POST /auth/solicitar-verificacao-email` e `POST /auth/verificar-email`; sem provedor de email, use `EMAIL_VERIFICATION_RETURN_TOKEN=true` em dev para receber o token na resposta.
- Cada login bem-sucedido registra `ultimo_login_em`.

## Migrations

As migrations usam o formato `golang-migrate` (`.up.sql` e `.down.sql`). Com Docker, nao e necessario instalar a CLI localmente.

```bash
docker compose up -d db
```

Comandos:

```bash
docker run --rm --network nous_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://nous:nous@db:5432/nous?sslmode=disable" up
docker run --rm --network nous_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://nous:nous@db:5432/nous?sslmode=disable" version
docker run --rm --network nous_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://nous:nous@db:5432/nous?sslmode=disable" down 1
```

Via Makefile:

```bash
make migrate-up
make migrate-status
make migrate-down
```

## Testes e Build

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go build -o bin/nous-webapi ./cmd/webapi
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

- Arquivo padrão: `./logs/nous.log`.
- No Docker Compose, o arquivo fica em `/app/logs/nous.log` dentro do volume `nous_backend_logs`.
- A aplicação também escreve no stdout quando `LOG_STDOUT=true`.
- Rotação por tamanho: `LOG_MAX_SIZE_MB`.
- Ciclo de vida/retenção: `LOG_MAX_AGE_DAYS` e `LOG_MAX_BACKUPS`.
- Compressão dos arquivos rotacionados: `LOG_COMPRESS=true`.

Variáveis disponíveis:

```dotenv
LOG_FILE_PATH=./logs/nous.log
LOG_MAX_SIZE_MB=50
LOG_MAX_AGE_DAYS=30
LOG_MAX_BACKUPS=10
LOG_COMPRESS=true
LOG_STDOUT=true
```

Auditoria:

- Persistida em PostgreSQL na tabela `auditoria` para rotas autenticadas.
- Registra o identificador do usuário autenticado, ação (método + path), IP, user agent, status HTTP, tempo de resposta e request id.
- Não registra corpos de requisição/resposta nem dados pessoais de leads/titulares; o campo `dados_novos` contém apenas método, path e request id (princípio da minimização).
- Falha ao salvar auditoria não quebra o fluxo da requisição.
- Retenção recomendada: 5 anos.

## LGPD e Segurança

- Captura de lead exige consentimento explícito e rastreabilidade de origem.
- Lead não convertido deve ser anonimizado após 12 meses sem base legal ativa.
- Solicitação do titular deve aplicar soft delete e anonimização quando não houver obrigação legal de retenção.
- A auditoria retém IP e user agent completos por até 5 anos com base no legítimo interesse de segurança da informação e prevenção a fraude/abuso (LGPD art. 7º, IX e art. 10). O registro é minimizado: contém apenas o identificador do usuário autenticado e metadados da requisição, sem dados de leads/titulares.
- Senhas são armazenadas com bcrypt e exigem no mínimo 8 caracteres (limite de 72 caracteres do bcrypt) na criação, alteração e redefinição.
- Login tem proteção contra força bruta em duas camadas: rate limit por IP no middleware e bloqueio temporário por conta após 10 falhas em 15 minutos (em memória, por instância; zera no login bem-sucedido e expira automaticamente, sem travamento permanente).
- Autorização usa RBAC e vínculo usuário-negócio.
- Refresh tokens são armazenados como hash, rotacionados no refresh e revogados no logout quando enviados no payload.
- Refresh tokens expirados ou revogados há mais de 24 horas são expurgados por job local iniciado junto com a API.
- A API aplica headers de segurança (nosniff, X-Frame-Options DENY, HSTS, Referrer-Policy, Permissions-Policy e Content-Security-Policy) e rate limit em endpoints públicos sensíveis. O CSP é restritivo (`default-src 'none'`) para respostas JSON; a página `/swagger` usa um CSP próprio mais permissivo por carregar a Swagger UI.
- Tráfego em produção deve usar HTTPS no gateway/load balancer.
- `CORS_ALLOWED_ORIGINS` deve conter apenas os domínios autorizados do frontend em produção, separados por vírgula.
- Sem `REDIS_URL`, o rate limit usa memória local e é adequado para desenvolvimento e instância única. Com `REDIS_URL`, o rate limit usa Redis e funciona de forma compartilhada entre múltiplas instâncias. Se o Redis estiver indisponível, a API cai temporariamente para memória local e tenta reconectar periodicamente.
- `OAUTH_STATE_SAMESITE` aceita `strict`, `lax` ou `none`. O padrão `strict` é o mais restritivo; use `lax` apenas se o provedor OAuth2 do ambiente exigir compatibilidade no redirect. O modo `none` só é aplicado quando `GOOGLE_REDIRECT_URL` usa HTTPS; caso contrário, a API preserva `strict`.

## Proteção da Informação no Repositório

- O repositório do Nous deve permanecer privado na organização Sognix no GitHub.
- O acesso deve ser concedido pelo menor privilégio necessário, com revisão periódica de membros e permissões.
- Branches principais devem usar proteção contra push direto, exigir pull request e validações antes do merge.
- Nunca versionar `.env`, credenciais, tokens, certificados, dumps de banco, dados pessoais reais, logs sensíveis ou artefatos de produção.
- Usar GitHub Secret Scanning, Dependabot e revisão de permissões quando disponíveis na organização.
- Revogar imediatamente qualquer credencial que tenha sido exposta em commit, issue, log, artefato ou ambiente compartilhado.
- Manter `RESET_PASSWORD_RETURN_TOKEN=false` fora de desenvolvimento/testes; quando habilitado, a API registra aviso na inicialização porque o token volta na resposta HTTP.
- Manter o aviso de propriedade e confidencialidade em [../NOTICE.md](../NOTICE.md).

## Alto Fluxo

Checklist antes de produção:

- Ajustar `pool_max_conns` conforme capacidade do PostgreSQL.
- Ativar rate limit distribuído no gateway/load balancer.
- Rodar `make test-integration` com `TEST_DATABASE_URL` apontando para um PostgreSQL descartável antes de promover migrations/repositórios.
- Garantir que o workflow [`../../.github/workflows/ci.yml`](../../.github/workflows/ci.yml) esteja obrigatório na proteção da branch.
- Configurar dashboards e alertas mínimos conforme [OBSERVABILIDADE.md](OBSERVABILIDADE.md).
- Verificar que `RESET_PASSWORD_RETURN_TOKEN=false` no ambiente promovido.
- Rodar teste de carga em login, `POST /leads`, autosave, exportação e listagens.
- Monitorar erros 409 para conflito de edição.
- Monitorar latência, saturação do pool, locks e queries lentas.
- Planejar auditoria assíncrona se o volume de escrita por request crescer.
- Implementar worker de retry/backoff para eventos pendentes/falhos no outbox.

## Comandos Úteis

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go build -o bin/nous-webapi ./cmd/webapi
go run ./cmd/webapi
```
