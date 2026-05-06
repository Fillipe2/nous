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
4. [x] **Captura de leads Tipo A (`prospecto`).** Tabela `prospecto` (migration
   000027) com perfil rico para alimentar vias 1 e 2.
   - Campos: `nome, email, telefone, categoria_pretendida_id (FK opcional),
     categoria_pretendida_livre, cidade, estado, porte (enum), interesses
     (jsonb), mensagem, origem, utm_*, status_funil,
     convertido_em_usuario_id`.
   - Endpoints: `POST /prospectos` (público), `GET /prospectos` (admin),
     `GET /prospectos/{id}` (admin), `PATCH /prospectos/{id}/funil` (admin).
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

## 8. Modelo de edição de LP (decisão tomada em 2026-05-05)

### Onde se edita

A LP pode ser editada em dois lugares — são duas interfaces para o mesmo dado,
não duas fontes de verdade:

- **Buildnix admin** — todas as operações: criar, editar, publicar, despublicar,
  duplicar, exportar ZIP, trocar template, configurações de plano.
- **Área logada da LP (no novo backend)** — subconjunto operacional: editar
  conteúdo (textos, imagens, horários, contato), publicar/despublicar, ver
  métricas e leads do próprio Negócio.

Funcionalidades estruturais (trocar template, mudar tipo white-label↔customizável,
limites de plano) ficam apenas no buildnix admin, onde há contexto de billing.

### Como a edição da área logada chega ao buildnix

O frontend da LP chama o **buildnix diretamente** com o JWT do Operador (emitido
pelo buildnix, válido lá). Não passa pelo novo backend como intermediário para
operações de escrita de LP — evita camada desnecessária e mantém buildnix como
única fonte de verdade.

### Como as alterações chegam ao novo backend — DECISÃO PENDENTE

**Ver seção 9 para a discussão completa e as opções em aberto.**

## 9. Decisões pendentes

- [ ] Nome do novo backend.
- [ ] Nome final do recurso Tipo A: `lead_aquisicao` × `prospecto` ×
      `inscricao` × `lead_b2b`. Sugestão de partida: `lead_aquisicao`.
- [ ] Quando criar o repositório do novo backend.
- [ ] Modelo de SSO unificado Sognix (futuro, fase 3+).
- [ ] Política de versionamento do schema de eventos publicados no Observer.
- [ ] **Mecanismo de publicação e propagação de LP ao novo backend** — ver §9.1.

### 9.1 Mecanismo de publicação de LP (decisão pendente)

#### Contexto

O buildnix já possui:
- `POST /landingpages/{id}/exportar` — gera um ZIP completo (HTML, CSS, JS,
  imagens e tudo mais) e armazena no PostgreSQL com hash SHA-256.
- `GET /exportacoes/{id}/download` — baixa o ZIP gerado.

O ZIP resolve o problema de assets (imagens, fontes, etc.) de forma natural:
tudo vai empacotado junto, sem necessidade de CDN separado ou storage externo.

A questão em aberto é: **como esse ZIP chega ao novo backend para ser servido
publicamente**, e como edições subsequentes atualizam a LP ao vivo.

#### Opção A — ZIP como mecanismo de publicação (deploy via arquivo)

O Operador edita no buildnix, clica "Publicar", o buildnix gera o ZIP e o
implanta automaticamente no hosting do novo backend.

```
Operador clica "Publicar" no buildnix
    → buildnix gera ZIP (reutiliza lógica do exportar)
    → buildnix envia ZIP ao novo backend (ou CDN/storage)
    → novo backend extrai e serve os arquivos estáticos
    → Observer recebe evento leve: { lp_id, exportacao_id, sha256, url_publica }
```

Qualquer re-edição gera um novo ZIP e repete o deploy. Edições via área logada
da LP seguem o mesmo fluxo: salva no buildnix → re-publica → novo ZIP.

**Prós:**
- Assets (imagens, fontes) vão dentro do ZIP — sem CDN ou storage extra.
- O ZIP já existe no buildnix; falta apenas o passo de deploy automático.
- SHA-256 garante integridade e permite cache agressivo.
- Simples de entender: publicar = gerar + implantar.

**Contras:**
- Re-deploy a cada edição publicada (tempo de propagação depende do tamanho
  do ZIP e da velocidade do hosting).
- O novo backend precisa de um endpoint para receber e extrair ZIPs.
- Edições frequentes via área logada geram muitos deploys.

**O que faltaria no buildnix:**
- Separar `exportar` (baixar para si) de `publicar` (deploy automático Sognix).
- `POST /landingpages/{id}/publicar` — gera ZIP e dispara o deploy.
- `POST /landingpages/{id}/despublicar` — tira do ar sem gerar ZIP.

---

#### Opção B — Evento com payload JSON (sem arquivo)

O buildnix emite um evento com o conteúdo estrutural da LP em JSON. O novo
backend mantém uma cópia ("config viva") em banco e renderiza a LP
dinamicamente.

```
Operador clica "Publicar" no buildnix
    → buildnix emite evento: { lp_id, version, conteudo (JSON), status }
    → novo backend atualiza lp_publicada em banco
    → novo backend renderiza LP a partir do JSON
```

**Prós:**
- Propagação instantânea (sem geração de arquivo).
- Re-edições são baratas — só atualiza o registro em banco.

**Contras:**
- Imagens e assets precisam de CDN ou storage externo — o JSON só referencia
  URLs, não carrega os bytes.
- Requer infraestrutura de upload de assets no buildnix (endpoint `POST
  /assets`, armazenamento de BLOBs, serving público) — **não existe hoje**.
- Novo backend vira um renderer dinâmico de LP, não um servidor de arquivos
  estáticos — superfície de código maior.
- O `conteudo` precisa ter schema versionado e estável como contrato entre
  buildnix e novo backend.

**O que faltaria no buildnix:**
- Domínio `asset` com upload, armazenamento e serving público de imagens.
- Schema formal e versionado do `conteudo` da LP.
- Evento `landingpage.publicada` com payload estruturado.

---

#### Opção C — Híbrido

ZIP para o primeiro deploy (ou deploys completos). Evento JSON apenas para
atualizações de conteúdo simples (texto, cor, telefone) sem re-geração de ZIP.

**Contras:** dois mecanismos de propagação para manter em sincronia — maior
complexidade.

---

#### Perguntas para a decisão

1. O novo backend vai servir LPs como **arquivos estáticos** ou como
   **páginas renderizadas dinamicamente**?
2. As edições via área logada da LP precisam propagar **imediatamente** ou
   toleramos o tempo de um re-deploy de ZIP?
3. Existe plano de infraestrutura para CDN/object storage (S3, R2, etc.) no
   curto prazo?
4. O volume esperado de edições por LP por dia é alto (justifica Opção B) ou
   baixo (Opção A é suficiente)?

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
