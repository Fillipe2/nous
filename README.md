# Nous — backend do Buildnix

Backend do **Buildnix**, o produto *builder/admin* da plataforma Sognix. O backend se chama **Nous** (νοῦς — intelecto divino que cria e ordena, em referência ao Nous de Plotino); o frontend que ele alimenta se chama **Buildnix**. Gerencia identidade do Operador, Unidades, templates, definição de landing pages, planos/billing, captura de leads de aquisição B2B do Sognix e ponte para o Observer.

O Nous opera em modo standalone e publica eventos no Observer quando configurado. Ele **não** serve as landing pages publicamente nem hospeda a operação vertical (agendamento, paciente, processo, etc.) — esses domínios pertencem ao novo backend de orquestração de LPs (em planejamento), que tem identidade própria validada via JWT do Nous.

## Posicionamento no ecossistema Sognix

```
┌──────────────────┐       ┌─────────────────────┐      ┌──────────┐      ┌────────┐
│  Frontend        │       │  NOUS (backend)     │      │ OBSERVER │      │ SBRAIN │
│  Buildnix        │◄─────►│  admin do builder   │─────►│   bus    │─────►│  CRM   │
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

**Quem fala com o Observer:** apenas o Nous. O novo backend reportará para o Nous por API interna e o Nous consolida e publica eventos canônicos para o Observer.

Detalhes de visão, fases e decisões pendentes em [docs/ROADMAP_ECOSSISTEMA.md](docs/ROADMAP_ECOSSISTEMA.md).

## Dois tipos de lead

- **Lead Tipo A — Aquisição Sognix (B2B).** Operador em potencial (Dr. João, Dra. Maria) que quer contratar produtos Sognix. Captura, gestão e funil comercial vivem no Nous.
- **Lead Tipo B — Captura do Operador (B2C).** Visitante da LP do Operador (Maria querendo agendar consulta na Clínica Sorriso). Vive no novo backend; o Nous só vê agregados via Observer/Sbrain. O recurso `lead` atual é Tipo B e migra para o novo backend quando ele nascer.

## Confidencialidade e propriedade

Este repositório contém código-fonte, documentação técnica, regras de negócio e artefatos internos de propriedade da Sognix. O conteúdo é confidencial, de uso restrito a pessoas autorizadas pela Sognix e não deve ser copiado, publicado, distribuído, sublicenciado ou reutilizado fora dos projetos e ambientes aprovados pela organização.

O repositório deve permanecer privado no GitHub. Credenciais, chaves, tokens, dumps de banco, dados pessoais reais, arquivos `.env` e artefatos de produção não devem ser versionados.

Consulte [NOTICE.md](NOTICE.md) para o aviso de propriedade e confidencialidade.

## Funcionalidades

- Autenticação JWT com refresh token e login social Google OAuth2.
- RBAC com perfis `admin` e `usuario`.
- Gestão de usuários, negocios, unidades, categorias, templates e landing pages.
- Dashboard por negocio sem expor leads ao usuário comum.
- Landing pages white label por templates prontos e landing pages customizáveis controladas por plano.
- Landing pages com autosave, lock pessimista e controle otimista por `version`.
- Captura pública de leads, com listagem restrita ao admin/Sognix.
- Exportação de landing pages em ZIP armazenado no PostgreSQL, com hash SHA-256.
- Auditoria local em PostgreSQL e logs operacionais rotativos em arquivo, com opção de stdout/stderr.
- Outbox local para eventos versionados enviados ao Observer quando configurado.

## Documentação

- [docs/ROADMAP_ECOSSISTEMA.md](docs/ROADMAP_ECOSSISTEMA.md): posicionamento do Nous no ecossistema Sognix, fases e decisões pendentes.
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
nous/
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
TEST_DATABASE_URL="postgres://usuario:senha@localhost:5432/nous_test?sslmode=disable" make test-integration
```

Use um banco descartável para essa URL.

No GitHub Actions, o workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) sobe um PostgreSQL descartável e executa esses testes com `TEST_DATABASE_URL` automaticamente.

## Migrations

```bash
docker compose up -d db
docker run --rm --network nous_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://nous:nous@db:5432/nous?sslmode=disable" up
docker run --rm --network nous_default -v "$PWD/migrations/postgres:/migrations" migrate/migrate -path=/migrations -database "postgres://nous:nous@db:5432/nous?sslmode=disable" version
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
go build -o bin/nous-webapi ./cmd/webapi
./bin/nous-webapi
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

## Modelo de entidades e relacionamentos

Esta é a referência rápida para o frontend entender o que cada ID significa e como as entidades se ligam. Detalhe campo a campo no Swagger (`/swagger`) — cada entidade tem um schema `*Output` com descrições.

| Entidade | O que é | Liga-se a |
| --- | --- | --- |
| **Operador** (`usuario`) | Quem faz login. Tem `role` (`usuario`/`admin`) e `status`. | N **Negócios** via `usuario_negocio` (N:N). |
| **Negócio** (`negocio`) | Entidade contratante (clínica/escritório), com CPF/CNPJ. **Antes `cliente`.** É o agregador central. | É dono de Unidades, Landing Pages, Templates próprios e Assinatura/Plano. |
| **Unidade** (`unidade`) | Sub-unidade de um Negócio (ex.: filial). **Antes `negocio`.** | `negocio_id` (dono). Opcionalmente referenciada por Landing Page. |
| **Categoria** | Vertical (odonto, jurídico…). Gerenciada por admin. | Classifica Negócios, Templates e Landing Pages. |
| **Template** | Base de Landing Page. `origem=sognix` (global) ou `origem=negocio` (do Negócio). | `negocio_id` quando `origem=negocio`; `categoria_id`. |
| **Landing Page** | Página do Negócio; autosave, lock e publicação. | `negocio_id` (obrigatório); `unidade_id`/`categoria_id`/`template_id` (opcionais). |
| **Exportação** | ZIP de uma Landing Page, com hash SHA-256. | `landingpage_id`, `usuario_id`. |
| **Lead (Tipo B)** | Visitante capturado pela LP pública. PII restrita a admin. | `landingpage_id`, `negocio_id`. |
| **Lead de Aquisição (Tipo A)** | Potencial Operador (funil B2B Sognix). | — |

```text
Operador ──(usuario_negocio, N:N)── Negócio ──1:N── Unidade
                                      │
                                      ├──1:N── LandingPage ──1:N── Exportacao
                                      │            └─ opcional → Unidade / Categoria / Template
                                      ├──1:N── Template (origem=negocio)
                                      └──1:N── Assinatura/Plano, métricas
LandingPage (pública) ──1:N── Lead (Tipo B)
```

**Regras que o frontend precisa conhecer:**

- **Qual ID enviar:** a maioria dos recursos é filtrada/criada por `negocio_id` (o Negócio dono). `unidade_id` é sempre opcional. Em Landing Page, `negocio_id` é obrigatório.
- **Self-onboarding:** ao criar um Negócio (`POST /negocios`), o vínculo `usuario_negocio` é criado automaticamente para o Operador autenticado.
- **Visibilidade:** `admin` (staff Sognix) vê tudo; `usuario` só vê os Negócios vinculados a ele.
- **Documento do Negócio:** CPF (11 dígitos) ou CNPJ (14, alfanumérico a partir de 2026), enviado sem máscara; validado por dígito verificador.
- **Concorrência:** entidades editáveis têm `version` (otimista, 409 em conflito); Landing Page tem lock pessimista (`lock`/`unlock`, 5 min).

## Endpoints Principais

Auth e conta:

- `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- `POST /auth/esqueci-senha`, `POST /auth/redefinir-senha`
- `POST /auth/solicitar-verificacao-email`, `POST /auth/verificar-email`
- `GET /auth/google/login`, `GET /auth/google/callback`
- `POST /usuarios` (cria sempre `usuario`), `GET /usuarios/me`, `PATCH /usuarios/me/senha`
- `PATCH /usuarios/{id}/role` (admin), `PATCH /usuarios/{id}/status` (admin)

Negócios e estrutura:

- `GET /negocios/me`, `GET /negocios` (admin), `POST /negocios`, `PUT /negocios/{id}`
- `GET /unidades?negocio_id={id}`, `POST /unidades`, `PUT /unidades/{id}`
- `GET /categorias`, `GET /templates`, `GET /dashboard`

Landing Pages e exportação:

- `GET/POST /landingpages`, `PUT /landingpages/{id}`
- `PATCH /landingpages/{id}/autosave|lock|unlock`
- `POST /landingpages/{id}/publicar|despublicar|exportar`
- `GET /exportacoes/{id}/download`

Leads e operação:

- `POST /leads` (público), `GET /leads` (admin/Sognix)
- `POST /lead_aquisicao` (público), `GET /lead_aquisicao` (admin)
- `GET /auditoria` (admin), `GET /swagger`, `GET /swagger/openapi.yaml`, `GET /health`

Detalhes completos em [docs/API.md](docs/API.md) e no Swagger interativo em `/swagger`.

No dashboard, admin pode usar filtros como `negocio_id`, `usuario_id`, `landingpage_id`, `unidade_id`, `categoria_id`, `template_id`, `tipo`, `data_inicio` e `data_fim`. Usuário comum usa a mesma rota, mas recebe apenas métricas dos negocios vinculados, sem dados pessoais de leads. A criação de negocio é self-onboarding para usuário autenticado e vincula automaticamente o negocio ao usuário criador.

## Concorrência

Entidades editáveis usam `version` para controle otimista. O frontend deve enviar a versão lida anteriormente em updates e autosave. Se a versão estiver obsoleta, a API retorna HTTP 409 e não sobrescreve os dados.

Landing pages também têm lock pessimista com `locked_by`, `locked_at` e expiração de 5 minutos para sessões de edição longas.

## Integração Sognix

Sem `OBSERVER_URL`, o Nous opera sozinho. Com `OBSERVER_URL`, eventos como `lead.capturado` são publicados em envelope versionado com `event_id`, `event_type`, `schema_version`, `source`, `occurred_at`, `idempotency_key` e `payload`.
