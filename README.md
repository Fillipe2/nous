# Buildnix Backend

Backend do **buildnix**, o produto *builder/admin* da plataforma Sognix. Gerencia identidade do Operador, Negócios, templates, definição de landing pages, planos/billing, captura de leads de aquisição B2B do Sognix e ponte para o Observer.

O buildnix opera em modo standalone e publica eventos no Observer quando configurado. Ele **não** serve as landing pages publicamente nem hospeda a operação vertical (agendamento, paciente, processo, etc.) — esses domínios pertencem ao novo backend de orquestração de LPs (em planejamento), que tem identidade própria validada via JWT do buildnix.

## Posicionamento no ecossistema Sognix

```
┌──────────────────┐       ┌─────────────────────┐      ┌──────────┐      ┌────────┐
│  Frontend        │       │  BUILDNIX BACKEND   │      │ OBSERVER │      │ SBRAIN │
│  Buildnix        │◄─────►│  (admin do builder) │─────►│   bus    │─────►│  CRM   │
│  (admin builder) │       │  identidade, planos │      └──────────┘      └────────┘
└──────────────────┘       └─────────▲───────────┘
                                     │ API interna
                                     ▼
                           ┌─────────────────────┐
                           │  NOVO BACKEND       │      ┌──────────────────┐
                           │  multi-tenant       │◄────►│ Frontends das LPs│
                           │  núcleo + módulos   │      │ (públicas + área │
                           │  verticais (odonto, │      │ logada do dono)  │
                           │  jurídico, ...)     │      └──────────────────┘
                           └─────────────────────┘
```

**Quem fala com o Observer:** apenas o buildnix. O novo backend reportará para o buildnix por API interna e o buildnix consolida e publica eventos canônicos para o Observer.

Detalhes de visão, fases e decisões pendentes em [docs/ROADMAP_ECOSSISTEMA.md](docs/ROADMAP_ECOSSISTEMA.md).

## Dois tipos de lead

- **Lead Tipo A — Aquisição Sognix (B2B).** Operador em potencial (Dr. João, Dra. Maria) que quer contratar produtos Sognix. Captura, gestão e funil comercial vivem no buildnix.
- **Lead Tipo B — Captura do Operador (B2C).** Visitante da LP do Operador (Maria querendo agendar consulta na Clínica Sorriso). Vive no novo backend; o buildnix só vê agregados via Observer/Sbrain. O recurso `lead` atual é Tipo B e migra para o novo backend quando ele nascer.

## Confidencialidade e propriedade

Este repositório contém código-fonte, documentação técnica, regras de negócio e artefatos internos de propriedade da Sognix. O conteúdo é confidencial, de uso restrito a pessoas autorizadas pela Sognix e não deve ser copiado, publicado, distribuído, sublicenciado ou reutilizado fora dos projetos e ambientes aprovados pela organização.

O repositório deve permanecer privado no GitHub. Credenciais, chaves, tokens, dumps de banco, dados pessoais reais, arquivos `.env` e artefatos de produção não devem ser versionados.

Consulte [NOTICE.md](NOTICE.md) para o aviso de propriedade e confidencialidade.

## Funcionalidades

- Autenticação JWT com refresh token e login social Google OAuth2.
- RBAC com perfis `admin` e `usuario`.
- Gestão de usuários, clientes, negócios, categorias, templates e landing pages.
- Dashboard por cliente sem expor leads ao cliente comum.
- Landing pages white label por templates prontos e landing pages customizáveis controladas por plano.
- Landing pages com autosave, lock pessimista e controle otimista por `version`.
- Captura pública de leads, com listagem restrita ao admin/Sognix.
- Exportação de landing pages em ZIP armazenado no PostgreSQL, com hash SHA-256.
- Auditoria local em PostgreSQL e logs operacionais rotativos em arquivo, com opção de stdout/stderr.
- Outbox local para eventos versionados enviados ao Observer quando configurado.

## Documentação

- [docs/ROADMAP_ECOSSISTEMA.md](docs/ROADMAP_ECOSSISTEMA.md): posicionamento do buildnix no ecossistema Sognix, fases e decisões pendentes.
- [docs/ARQUITETURA.md](docs/ARQUITETURA.md): arquitetura, banco, concorrência, alto fluxo e integração.
- [docs/API.md](docs/API.md): endpoints, contratos importantes e respostas de erro.
- [docs/OPERACAO.md](docs/OPERACAO.md): ambiente, migrations, testes, build, deploy, logs e LGPD.
- [NOTICE.md](NOTICE.md): aviso de propriedade intelectual, confidencialidade e uso restrito.
- [rules/CONVENCOES.md](rules/CONVENCOES.md): regras de código Go e padrões do projeto.

As pastas `docs/` e `rules/` são a referência do projeto.

## Arquitetura

O projeto segue Clean Architecture:

```text
Interface -> App -> Domain <- Infra
```

Estrutura principal:

```text
buildnix-backend/
├── cmd/
│   ├── webapi/              # Entrypoint da API REST
│   └── internal/bootstrap/  # Montagem de config, infra, repos e use cases
├── internal/
│   ├── domain/              # Entidades e interfaces, sem dependências externas
│   ├── app/                 # DTOs e casos de uso
│   ├── infra/               # PostgreSQL, JWT e event bus
│   └── interface/webapi/    # Server HTTP, handlers, middlewares, router e OpenAPI
├── migrations/postgres/     # Migrations SQL separadas por recurso
├── bin/                     # Binários compilados
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Pré-requisitos

- Go 1.26.2
- PostgreSQL 15+
- GCC para `go test -race`
- Docker para rodar migrations com `golang-migrate`
- Docker e Docker Compose, opcionalmente

## Configuração

```bash
cp .env.example .env
```

Variáveis principais:

| Variável | Descrição |
| --- | --- |
| `DATABASE_URL` | URL PostgreSQL. Aceita parâmetros do `pgxpool`, como `pool_max_conns`. |
| `JWT_SECRET` | Segredo do access token JWT. Trocar em produção. |
| `JWT_REFRESH_SECRET` | Segredo do refresh token JWT. Trocar em produção. |
| `ADMIN_NOME` | Nome do admin inicial. |
| `ADMIN_EMAIL` | Email do admin inicial. Se vazio, não cria seed. |
| `ADMIN_SENHA` | Senha do admin inicial. Se vazia, não cria seed. |
| `GOOGLE_CLIENT_ID` | Client ID do Google OAuth2. |
| `GOOGLE_CLIENT_SECRET` | Client Secret do Google OAuth2. |
| `GOOGLE_REDIRECT_URL` | Callback OAuth2. |
| `OAUTH_STATE_SAMESITE` | Modo SameSite do cookie OAuth state: `strict`, `lax` ou `none`. `none` exige `GOOGLE_REDIRECT_URL` em HTTPS; sem HTTPS, a API usa `strict`. |
| `OBSERVER_URL` | URL do Observer. Vazio mantém operação standalone. |
| `CORS_ALLOWED_ORIGINS` | Origens autorizadas para CORS, separadas por vírgula. |
| `REDIS_URL` | Redis opcional para rate limit distribuído em produção multi-instância. |
| `RESET_PASSWORD_RETURN_TOKEN` | Se `true`, retorna token de redefinição na resposta HTTP. Use apenas em desenvolvimento/testes; a inicialização registra aviso quando habilitado. |
| `SERVER_PORT` | Porta HTTP da API. |

Planos customizáveis devem definir `landingpages_customizaveis=true` em `plano.limites`. Planos white label simples devem omitir essa chave ou usar `false`.

### Testes de integração

Os testes contra PostgreSQL real ficam protegidos por `TEST_DATABASE_URL` e criam um schema temporário para aplicar as migrations reais sem tocar nas tabelas existentes do banco informado.

```bash
TEST_DATABASE_URL="postgres://usuario:senha@localhost:5432/buildnix_test?sslmode=disable" make test-integration
```

Use um banco descartável para essa URL.

No GitHub Actions, o workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) sobe um PostgreSQL descartável e executa esses testes com `TEST_DATABASE_URL` automaticamente.

## Migrations

```bash
docker compose up -d db
docker run --rm --network buildnix_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://buildnix:buildnix@db:5432/buildnix?sslmode=disable" up
docker run --rm --network buildnix_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://buildnix:buildnix@db:5432/buildnix?sslmode=disable" version
```

Com `make` instalado, os atalhos equivalentes sao:

```bash
make migrate-up
make migrate-status
```

## Build, Testes e Execução

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go build -o bin/buildnix-backend-webapi ./cmd/webapi
./bin/buildnix-backend-webapi
```

Com Makefile:

```bash
make test
make build
make run
```

## Docker

```bash
docker compose up --build
```

Depois que o PostgreSQL estiver pronto, rode as migrations com `make migrate-up`.

## Endpoints Principais

- `POST /auth/login`
- `POST /auth/refresh`
- `GET /health`
- `GET /auth/google/login`
- `GET /auth/google/callback`
- `POST /usuarios`
- `GET /usuarios/me`
- `GET /clientes/me`
- `GET /clientes`
- `POST /clientes`
- `PUT /clientes/{id}`
- `GET /dashboard`
- `GET /negocios?cliente_id={id}`
- `GET /landingpages`
- `POST /landingpages`
- `PUT /landingpages/{id}`
- `PATCH /landingpages/{id}/autosave`
- `PATCH /landingpages/{id}/lock`
- `PATCH /landingpages/{id}/unlock`
- `POST /landingpages/{id}/exportar`
- `GET /exportacoes/{id}/download`
- `POST /leads`
- `GET /swagger`
- `GET /leads` somente admin/Sognix
- `GET /auditoria` admin
- `GET /swagger/openapi.yaml`

Detalhes completos em [docs/API.md](docs/API.md).

No dashboard, admin pode usar filtros como `cliente_id`, `usuario_id`, `landingpage_id`, `negocio_id`, `categoria_id`, `template_id`, `tipo`, `data_inicio` e `data_fim`. Usuário comum usa a mesma rota, mas recebe apenas métricas dos clientes vinculados, sem dados pessoais de leads. A criação de cliente é self-onboarding para usuário autenticado e vincula automaticamente o cliente ao usuário criador.

## Concorrência

Entidades editáveis usam `version` para controle otimista. O frontend deve enviar a versão lida anteriormente em updates e autosave. Se a versão estiver obsoleta, a API retorna HTTP 409 e não sobrescreve os dados.

Landing pages também têm lock pessimista com `locked_by`, `locked_at` e expiração de 5 minutos para sessões de edição longas.

## Integração Sognix

Sem `OBSERVER_URL`, o Buildnix opera sozinho. Com `OBSERVER_URL`, eventos como `lead.capturado` são publicados em envelope versionado com `event_id`, `event_type`, `schema_version`, `source`, `occurred_at`, `idempotency_key` e `payload`.
