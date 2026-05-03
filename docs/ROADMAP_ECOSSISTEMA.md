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

1. **Account linking Google + email já existente.** Se o email do Google bate
   com `usuario` existente e `email_verified=true`, criar só o `usuario_auth`
   vinculado, sem duplicar conta.
2. **Senha para usuário criado via Google.** Reaproveitar
   `/auth/esqueci-senha` (já funciona; só validar UX).
3. **Consentimento granular no `lead` (Tipo B atual).** Adicionar
   `consentimento_marketing_parceiros bool` + `consentimento_versao_marketing
   string`. Habilita via 3 (cross-sell) sem migration extra depois.
4. **Captura de leads Tipo A (`lead_aquisicao`).** Tabela nova com perfil
   rico para alimentar vias 1 e 2.
   - Campos: `nome, email, telefone, categoria_pretendida_id (FK opcional),
     categoria_pretendida_livre, cidade, estado, porte (enum), interesses
     (jsonb), mensagem, origem, utm_*, status_funil,
     convertido_em_usuario_id`.
   - Endpoints: `POST /prospectos` (público), `GET /prospectos` (admin),
     `PATCH /prospectos/{id}` (admin).
5. **JWKS público em buildnix** — opcional na fase 1; vira obrigatório quando
   o novo backend nascer. Endpoint `GET /.well-known/jwks.json`.

### Não fazer em buildnix (vai pro novo backend)

- Pessoa global / Contato cross-Negócio.
- Domínio operacional vertical (paciente, agendamento, processo).
- Tracking de CTA / cliques / interações.
- Multi-tenant runtime de LP.
- Marketplace de fornecedores.
- Pipeline de anonimização.
- IA/CRM do Sbrain.

## 8. Decisões pendentes

- [ ] Nome do novo backend.
- [ ] Nome final do recurso Tipo A: `lead_aquisicao` × `prospecto` ×
      `inscricao` × `lead_b2b`. Sugestão de partida: `lead_aquisicao`.
- [ ] Quando criar o repositório do novo backend.
- [ ] Modelo de SSO unificado Sognix (futuro, fase 3+).
- [ ] Política de versionamento do schema de eventos publicados no Observer.

## 9. Glossário

- **Operador** — usuário do buildnix que contrata produtos Sognix
  (`internal/domain/usuario`).
- **Negócio** — entidade contratante (CNPJ, LP, operação). Em buildnix está
  modelado como `cliente` (legado) com sub-unidade `negocio`. Renomear no
  futuro para evitar overload.
- **Contato** — pessoa registrada por um Negócio (lead/paciente/cliente final).
  No buildnix atual é `lead`; no novo backend será `contato` ligado a `pessoa`.
- **Pessoa** (futuro) — entidade global do ser humano. Vive no novo backend.
- **Lead Tipo A / Aquisição** — prospecto B2B do Sognix, capturado pelo
  buildnix.
- **Lead Tipo B / Captura** — lead B2C do Operador, capturado pelo novo
  backend.
- **Categoria** — discriminador de produto vertical (`dentista`, `advogado`,
  ...). Determina qual módulo do novo backend opera o Negócio.
- **Observer** — bus de eventos do ecossistema Sognix.
- **Sbrain** — produto de CRM/BI/IA da Sognix, consumidor do Observer.
