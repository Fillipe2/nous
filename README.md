# Buildnix Backend

Backend do sistema de criação e gestão de landing pages whitelabel da plataforma Sognix.

O Buildnix funciona em modo standalone e mantém integração opcional com o Observer para publicação de eventos do ecossistema Sognix.

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
- Goose para migrations
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
| `OBSERVER_URL` | URL do Observer. Vazio mantém operação standalone. |
| `SERVER_PORT` | Porta HTTP da API. |

Planos customizáveis devem definir `landingpages_customizaveis=true` em `plano.limites`. Planos white label simples devem omitir essa chave ou usar `false`.

## Migrations

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir migrations/postgres postgres "$DATABASE_URL" up
goose -dir migrations/postgres postgres "$DATABASE_URL" status
goose -dir migrations/postgres postgres "$DATABASE_URL" down
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

Depois que o PostgreSQL estiver pronto, rode as migrations com Goose usando a `DATABASE_URL` do ambiente.

## Endpoints Principais

- `POST /auth/login`
- `POST /auth/refresh`
- `GET /auth/google/login`
- `GET /auth/google/callback`
- `POST /usuarios`
- `GET /usuarios/me`
- `GET /clientes/me`
- `GET /clientes`
- `POST /clientes`
- `PUT /clientes/{id}`
- `GET /dashboard?cliente_id={id}`
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
- `GET /leads` admin/Sognix
- `GET /auditoria` admin
- `GET /swagger/openapi.yaml`

Detalhes completos em [docs/API.md](docs/API.md).

## Concorrência

Entidades editáveis usam `version` para controle otimista. O frontend deve enviar a versão lida anteriormente em updates e autosave. Se a versão estiver obsoleta, a API retorna HTTP 409 e não sobrescreve os dados.

Landing pages também têm lock pessimista com `locked_by`, `locked_at` e expiração de 5 minutos para sessões de edição longas.

## Integração Sognix

Sem `OBSERVER_URL`, o Buildnix opera sozinho. Com `OBSERVER_URL`, eventos como `lead.capturado` são publicados em envelope versionado com `event_id`, `event_type`, `schema_version`, `source`, `occurred_at`, `idempotency_key` e `payload`.
