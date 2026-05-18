# Roadmap do Ecossistema Sognix

Este documento registra decisões arquiteturais sobre o posicionamento do
**buildnix-backend** dentro do ecossistema Sognix e o caminho para os próximos
produtos. É a fonte de verdade para retomar o trabalho em outras threads.

## 1. Posicionamento do buildnix

O buildnix é **um dos produtos da Sognix** — não o núcleo. Sua responsabilidade
é o lado *builder/admin*: identidade do Operador, gestão de Negócios,
templates, definição de landing pages, planos/billing, captura de leads de
aquisição (Tipo A) e ponte para o Observer.

O buildnix **não** é responsável por:

- Servir as landing pages publicamente.
- Operação vertical (paciente/agenda/pagamento/processo).
- Captura de leads B2C dos Operadores (Tipo B).
- Tracking de cliques/interações dos visitantes da LP.
- Identidade global de pessoa (dedup cross-Negócio).

Tudo isso é responsabilidade do **novo backend** (ver §3).

## 2. Mapa do ecossistema

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
                           │  (substitui steeth) │◄────►│ Frontends das LPs│
                           │  multi-tenant       │      │ (públicas + área │
                           │  núcleo + módulos   │      │ logada do dono)  │
                           │  verticais          │      └──────────────────┘
                           └─────────────────────┘
```

**Quem fala com o Observer:** apenas o buildnix. O novo backend reporta para o
buildnix por API interna; o buildnix consolida e publica eventos canônicos no
Observer. Isso garante um único ponto de saída para o ecossistema.

## 3. Novo backend (a ser criado, em outro repositório)

Cenário definido: **plataforma com núcleo + módulos verticais** (não é só
runtime de LP).

- **Nome ainda não decidido.** Sugestões: `runnix` / `flownix` / `opsnix`
  (mantendo o sufixo "nix") ou `sognix-runtime` / `sognix-platform` /
  `atlas` / `nexus`. Evitar `sognix-core` (Sognix tem múltiplos produtos).
- **Multi-tenant por Negócio.**
- **Núcleo:** Operador (espelho), Negócio (espelho), Categoria (espelho),
  LP publicada (config viva), Pessoa global (dedup cross-Negócio), Contato por
  Negócio, interações (cliques de CTA, redirects), auditoria operacional.
- **Módulos verticais** com extensões específicas:
  - `odonto/` — paciente, dentista, agendamento, prescrição, pagamento,
    certificado médico, escala de disponibilidade.
  - `juridico/` (futuro) — cliente, processo, audiência, prazo.
  - novos módulos conforme novos verticais.
- **Auth:** valida JWT emitido pelo buildnix. Buildnix exporá
  `/.well-known/jwks.json` quando o novo backend nascer.

O **steeth-backend atual** vira referência viva durante a migração. Quando o
módulo `odonto/` do novo backend estiver em paridade, o frontend troca a base
URL e o steeth é arquivado.

## 4. Identidade no ecossistema

- **Operador** (Dr. João, Dra. Maria) — pessoa que contrata produtos Sognix.
  Login único no buildnix; o novo backend valida o JWT.
- **Negócio** — entidade que tem CNPJ, LP e operação (Clínica Sorriso,
  Advocacia João). Um Operador pode ter N Negócios; cada Negócio tem uma
  `categoria` que define qual módulo vertical opera.
- **Pessoa** (no novo backend) — registro global do ser humano que aparece em
  Negócios. Identificada por email/CPF/telefone. Não é usuário, não tem login.
- **Contato** (no novo backend) — relação `(pessoa, negocio)`. Status no funil,
  origem, UTM, consentimentos por Negócio. Privacidade isolada: dois Negócios
  nunca veem o `Contato` um do outro.
- **Paciente / ClienteJuridico / etc.** — extensões verticais do `Contato`.

**Privacidade LGPD:** o Negócio é o **controlador** dos dados de Contato/
Pessoa. A Sognix é **processadora**. Isso precisa estar refletido nos
contratos com Operadores e nos Termos de Privacidade das LPs.

## 5. Os dois tipos de lead

| | **Lead Tipo A — Aquisição Sognix** | **Lead Tipo B — Captura do Operador** |
|---|---|---|
| Quem é | Dr. João, dono de clínica | Maria, paciente em potencial |
| Onde mora | **buildnix** (`lead_aquisicao`) | **novo backend** (`Contato`) |
| Origem | LP institucional Sognix, anúncios | LP da Clínica Sorriso |
| Quem vê | Time comercial Sognix | O próprio Operador |
| Visão Sognix | Direta | Apenas agregada (via Observer/Sbrain), nunca PII individual |
| Vira | Operador | Paciente/cliente do Operador |

O `lead` que existe hoje em buildnix é Tipo B e **será migrado** para o novo
backend. Até lá permanece em buildnix como contrato vivo.

## 6. Vias de monetização

Todas as cinco fazem parte da estratégia. Ordem do mais simples para o mais
ambicioso:

1. **SaaS por uso ao Operador.** Plano por volume de leads/LPs/módulos. Já
   existe esqueleto em `plano`, `assinatura_cliente`, `uso_cliente_evento`.
2. **Marketplace B2B.** Vender o Operador qualificado para fornecedores. Usa
   `lead_aquisicao` enriquecido (cidade, porte, especialidade, necessidades).
   Operador é o titular dos dados, autoriza explicitamente. Modelo: assinatura
   do fornecedor + comissão por conversão.
3. **Cross-sell ao paciente final (opt-in).** Maria marca "aceito ofertas de
   parceiros". Sognix oferece financiamento/seguro/plano dental. Comissão por
   conversão. Exige opt-in granular auditável e opt-out fácil.
4. **Sbrain como upsell de IA/CRM.** Qualificação de lead, previsão de
   no-show, sugestões de timing. Usa apenas dados do próprio Negócio.
5. **Inteligência de mercado anonimizada.** Relatórios de tendência sem PII.

## 7. Plano para finalizar o buildnix

Trabalho mínimo para fechar buildnix sem trancar nenhum caminho de receita.

### Fase 1 — Em andamento

- [x] Swagger UI redesenhado (filtros, sort, tema).
- [x] OAuth Google `SameSite=Lax` por padrão.
- [x] `.env` carregado automaticamente em `go run` via `godotenv`.
- [x] Documentação dos endpoints `/auth/google/*` esclarecendo que não dá pra
      testar via "Try it out" do Swagger.
- [x] Logs detalhados no callback Google.

### Fase 2 — Próximos passos combinados

1. [x] **Account linking Google + email já existente.** Se o email do Google bate
   com `usuario` existente e `email_verified=true`, criar só o `usuario_auth`
   vinculado, sem duplicar conta.
2. [x] **Senha para usuário criado via Google.** Reaproveitar
   `/auth/esqueci-senha` (já funciona; fluxo confirmado funcional para
   usuários sem senha local).
3. [x] **Consentimento granular no `lead` (Tipo B atual).** Adicionados
   `consentimento_marketing_parceiros bool` + `consentimento_versao_marketing
   string` (migration 000026). Habilita via 3 (cross-sell) sem migration
   extra depois.
4. [x] **Captura de leads Tipo A (`lead_aquisicao`).** Tabela `lead_aquisicao` (migration
   000027, renomeada em 000029) com perfil rico para alimentar vias 1 e 2.
   - Campos: `nome, email, telefone, categoria_pretendida_id (FK opcional),
     categoria_pretendida_livre, cidade, estado, porte (enum), interesses
     (jsonb), mensagem, origem, utm_*, status_funil,
     convertido_em_usuario_id`.
   - Endpoints: `POST /lead_aquisicao` (público), `GET /lead_aquisicao` (admin),
     `GET /lead_aquisicao/{id}` (admin), `PATCH /lead_aquisicao/{id}/funil` (admin).
5. [x] **JWKS público em buildnix.** Endpoint `GET /.well-known/jwks.json`.
   Ativo quando `JWT_RSA_PRIVATE_KEY` está configurado (RS256). Sem a chave,
   retorna 503. Aceita PKCS8 e PKCS1.

### Não fazer em buildnix (vai pro novo backend)

- Pessoa global / Contato cross-Negócio.
- Domínio operacional vertical (paciente, agendamento, processo).
- Tracking de CTA / cliques / interações.
- Multi-tenant runtime de LP.
- Marketplace de fornecedores.
- Pipeline de anonimização.
- IA/CRM do Sbrain.

## 8. Modelo de edição e publicação de LP (atualizado 2026-05-09)

### Onde se edita

Buildnix é a **única fonte de verdade** para LP. Duas interfaces, mesmos dados:

- **Buildnix admin** — todas as operações: criar, editar, publicar, despublicar,
  duplicar, exportar ZIP, trocar template, configurações de plano.
- **Área logada da LP** — subconjunto básico: editar textos e cores.
  Chama o buildnix diretamente com o JWT do Operador. Não passa pelo novo backend.

Funcionalidades estruturais (trocar template, mudar tipo white-label↔customizável,
limites de plano) ficam apenas no buildnix admin, onde há contexto de billing.

O novo backend **não participa** do fluxo de edição ou publicação de LP.

### Mecanismo de publicação (decisão tomada em 2026-05-09)

```
Operador clica "Publicar"
    → buildnix gera ZIP (reutiliza lógica do exportar)
    → buildnix faz deploy do ZIP para hosting estático externo via API
    → buildnix armazena url_publica + deploy_id + sha256 na LP
    → Observer recebe evento: { lp_id, exportacao_id, sha256, url_publica }
```

Edições via área logada seguem o mesmo fluxo: salva no buildnix → re-publica → novo ZIP.

**Plataforma de hosting:** ainda não definida (Netlify, Vercel, Render, VPS própria).
A integração fica atrás de uma interface (`hosting.Deployer`) configurada por env var,
exatamente como o Observer. Quando a plataforma for escolhida, implementa-se só o adaptador.

### O que falta implementar no buildnix

- Interface `hosting.Deployer` no domínio (ou infra) com métodos `Deploy` e `Despublicar`.
- `POST /landingpages/{id}/publicar` — gera ZIP e chama `Deployer.Deploy`.
- `POST /landingpages/{id}/despublicar` — chama `Deployer.Despublicar` e atualiza status.
- Campos na LP/exportação: `url_publica`, `deploy_id`, `publicada_em`.
- Adaptador noop para operar standalone sem hosting configurado.

## 9. Decisões pendentes

- [ ] Nome do novo backend.
- [x] Nome final do recurso Tipo A: **`lead_aquisicao`** (decidido e implementado em 2026-05-18).
- [ ] Quando criar o repositório do novo backend.
- [ ] Modelo de SSO unificado Sognix (futuro, fase 3+).
- [ ] Política de versionamento do schema de eventos publicados no Observer.
- [ ] **Plataforma de hosting estático para deploy de LP** (Netlify, Vercel, Render, VPS).

## 9. Glossário

- **Operador** — usuário do buildnix que contrata produtos Sognix
  (`internal/domain/usuario`).
- **Negócio** — entidade contratante (CNPJ, LP, operação). Em buildnix está
  modelado como `cliente` (legado) com sub-unidade `negocio`. Renomear no
  futuro para evitar overload.
- **Contato** — pessoa registrada por um Negócio (lead/paciente/cliente final).
  No buildnix atual é `lead`; no novo backend será `contato` ligado a `pessoa`.
- **Pessoa** (futuro) — entidade global do ser humano. Vive no novo backend.
- **Lead Tipo A / Aquisição** (`lead_aquisicao`) — lead B2B do Sognix, capturado pelo
  buildnix. Potencial Operador interessado em contratar produtos Sognix.
- **Lead Tipo B / Captura** — lead B2C do Operador, capturado pelo novo
  backend.
- **Categoria** — discriminador de produto vertical (`dentista`, `advogado`,
  ...). Determina qual módulo do novo backend opera o Negócio.
- **Observer** — bus de eventos do ecossistema Sognix.
- **Sbrain** — produto de CRM/BI/IA da Sognix, consumidor do Observer.
