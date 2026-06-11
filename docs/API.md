# API Buildnix

Base local padrão: `http://localhost:8080`.

OpenAPI embarcado: `GET /swagger/openapi.yaml`.

## Autenticação

- `POST /auth/login`: login tradicional.
- `POST /auth/refresh`: renova token JWT.
- `POST /auth/logout`: logout.
- `POST /auth/esqueci-senha`: solicita redefinição de senha com resposta genérica.
- `POST /auth/redefinir-senha`: redefine senha usando token de uso único e revoga refresh tokens ativos do usuário.
- `GET /auth/google/login`: inicia OAuth2 Google.
- `GET /auth/google/callback`: callback OAuth2 Google.

Rotas autenticadas usam header:

```http
Authorization: Bearer <token>
```

## Usuários

- `POST /usuarios`: cadastro público de usuário comum.
- `GET /usuarios/me`: dados do usuário autenticado.
- `PATCH /usuarios/me/senha`: troca a senha do usuário autenticado informando senha atual e nova senha.
- `GET /usuarios`: lista usuários, admin.
- `PUT /usuarios/{id}`: edita nome e email, próprio usuário ou admin.
- `DELETE /usuarios/{id}`: soft delete, próprio usuário ou admin.

Filtros de data em `GET /usuarios` usam apenas data no formato `YYYY-MM-DD`; `data_fim` é inclusiva para o dia inteiro.

## Clientes e Relacionamentos

- `GET /clientes/me`: lista clientes vinculados ao usuário autenticado, retornando apenas dados básicos.
- `GET /clientes`: lista clientes, admin.
- `POST /clientes`: cria cliente via self-onboarding autenticado com nome, email, telefone e whatsapp, vinculando automaticamente ao usuário criador.
- `PUT /clientes/{id}`: edita cliente, admin. Requer `version`.
- `DELETE /clientes/{id}`: soft delete cliente, admin.

## Dashboard

- `GET /dashboard`: dashboard do usuário autenticado. Admin recebe visão global da operação e pode recortar por `cliente_id`, `usuario_id`, `landingpage_id`, `negocio_id`, `categoria_id`, `template_id` e `tipo`. Usuário comum recebe métricas agregadas dos clientes vinculados ao próprio usuário, sem precisar enviar `cliente_id`.

Filtros de data usam `YYYY-MM-DD`, com `data_fim` inclusiva para o dia inteiro. Dados pessoais de leads continuam restritos às rotas de leads admin-only; no dashboard comum não há listagem nem dados pessoais de leads.

## Opções para combos e filtros

Use os endpoints `/opcoes/*` para popular combos, selects e filtros do frontend. Eles retornam itens leves no formato `{ id, nome, label, status, tipo, origem, cliente_id, categoria_id }`, com `total`, `pagina` e `por_pagina`. A busca textual usa `q` e a ordenação padrão é por nome crescente. O frontend deve adicionar localmente a opção visual "Todos" quando quiser representar ausência de filtro.

- `GET /opcoes/clientes`: clientes disponíveis para o usuário. Admin/Sognix vê todos; usuário comum vê apenas os vinculados.
- `GET /opcoes/usuarios`: usuários para filtros administrativos, admin.
- `GET /opcoes/negocios`: negócios disponíveis. Admin/Sognix vê todos; usuário comum vê apenas negócios dos clientes vinculados.
- `GET /opcoes/categorias`: categorias disponíveis.
- `GET /opcoes/templates`: templates disponíveis. Usuário comum deve informar `cliente_id`; o backend combina templates Sognix e templates do cliente, respeitando o plano.
- `GET /opcoes/landingpages`: landing pages disponíveis. Admin/Sognix vê todas; usuário comum vê apenas LPs dos clientes vinculados.
- `GET /opcoes/enums`: tipos e status estáticos para filtros, como tipos de LP, status de lead, origem de template e tipos de template.

Exemplo para combo de cliente:

```http
GET /opcoes/clientes?q=clinica&por_pagina=100
```

Exemplo para combo de template após selecionar cliente e categoria:

```http
GET /opcoes/templates?cliente_id={cliente_id}&categoria_id={categoria_id}&tipo=white_label
```

## Negócios

- `GET /negocios?cliente_id={id}`: lista negócios do cliente vinculado.
- `POST /negocios`: cria negócio para cliente vinculado.
- `PUT /negocios/{id}`: edita negócio do cliente vinculado. Requer `version`.
- `DELETE /negocios/{id}`: soft delete negócio.

## Landing Pages

- `GET /landingpages`: lista landing pages. Filtros: `cliente_id`, `negocio_id`, `categoria_id`, `tipo`, `status`.
- `POST /landingpages`: cria landing page.
- `GET /landingpages/{id}`: detalhe da landing page.
- `PUT /landingpages/{id}`: edita landing page. Requer `version`.
- `PATCH /landingpages/{id}/autosave`: salva rascunho. Requer `version` e retorna nova `version`.
- `PATCH /landingpages/{id}/lock`: bloqueia edição concorrente.
- `PATCH /landingpages/{id}/unlock`: desbloqueia edição.
- `POST /landingpages/{id}/exportar`: upload multipart do ZIP exportado pelo front.
- `DELETE /landingpages/{id}`: soft delete.

Tipos de landing page:

- `white_label`: exige `template_id` e usa um template pronto cadastrado no banco. O backend bloqueia autosave/layout livre para esse tipo.
- `customizavel`: permite iniciar com `template_id` e depois customizar `conteudo`/`autosave_data`, desde que o plano ativo do cliente tenha `landingpages_customizaveis=true`.

Exemplo white label:

```json
{
  "cliente_id": "uuid",
  "template_id": "uuid",
  "tipo": "white_label",
  "nome": "LP Institucional"
}
```

Exemplo customizável começando de um template:

```json
{
  "cliente_id": "uuid",
  "template_id": "uuid",
  "tipo": "customizavel",
  "nome": "LP Campanha"
}
```

O plano simples white label deve manter `landingpages_customizaveis=false` ou omitir essa chave em `plano.limites`.

Exemplo de autosave:

```json
{
  "autosave_data": {"blocos": []},
  "version": 3
}
```

Resposta:

```json
{
  "mensagem": "rascunho salvo",
  "version": 4
}
```

## Categorias e Templates

- `GET /categorias`: lista categorias.
- `POST /categorias`: cria categoria, admin.
- `PUT /categorias/{id}`: edita categoria, admin. Requer `version`.
- `DELETE /categorias/{id}`: soft delete categoria, admin.
- `GET /templates`: lista templates. Filtros: `cliente_id`, `categoria_id`, `tipo`, `origem`.
- `POST /templates`: cadastra template pronto no banco. Admin pode criar template oficial Sognix ou template de cliente; usuário comum só pode criar template do cliente ao qual tem acesso. Aceita multipart com `cliente_id`, `origem`, `nome`, `tipo`, `categoria_id`, `conteudo` e/ou `arquivo`. Para white label, envie o `.zip` em `arquivo`: ele é validado (estrutura, zip-slip, limite de descompressão) e armazenado em binário. Arquivos de texto precisam ser UTF-8 válido e são salvos em `conteudo`.
- `GET /templates/{id}/download`: baixa o pacote `.zip` do template. Templates Sognix ficam disponíveis a qualquer usuário autenticado; templates de cliente exigem acesso ao cliente dono. Retorna `404` se o template não tiver arquivo binário.
- `DELETE /templates/{id}`: soft delete template, admin.

Na tela de criação/edição de LP, use:

```http
GET /templates?cliente_id={cliente_id}
```

Essa listagem retorna templates `origem=sognix` e templates `origem=cliente` daquele cliente. Se o plano do cliente não permitir customizável, templates `base_customizavel` são omitidos.

Origens de template:

- `sognix`: template oficial criado pela Sognix, global para clientes autorizados pelo plano.
- `cliente`: template criado/salvo por um cliente específico, visível apenas para aquele cliente e admin.

Tipos de template:

- `white_label`: disponível para landing pages white label e também como ponto de partida de customizáveis.
- `base_customizavel`: base para clientes com plano customizável; não pode ser usada para criar LP `white_label`.

## Leads

- `POST /leads`: captura pública via formulário da landing page.
- `GET /leads`: lista leads capturados, somente admin/Sognix. Aceita `cliente_id` para ver leads dos clientes cadastrados.
- `GET /leads/{id}`: detalhe do lead, somente admin/Sognix.
- `PATCH /leads/{id}/status`: atualiza status, somente admin/Sognix.

Leads são dados pessoais e ativo comercial da Sognix. Cliente comum não tem acesso nominal nem métricas de leads.

## Exportações

- `POST /landingpages/{id}/exportar`: upload multipart do ZIP exportado pelo front.
- `GET /exportacoes?landingpage_id={id}`: lista exportações de uma landing page.
- `GET /exportacoes/{id}/download`: download seguro do ZIP.

Exportações respeitam o vínculo do usuário com o cliente dono da landing page.

## Auditoria

- `GET /auditoria`: lista auditoria, admin.
- `GET /auditoria/{id}`: detalhe do evento de auditoria, admin.

## Erros

Erros HTTP retornam JSON com os campos:

```json
{
  "erro": "mensagem legada compatível",
  "codigo": "conflict",
  "mensagem": "registro foi alterado por outro usuario, recarregue os dados e tente novamente",
  "request_id": "req-id"
}
```

Conflito de edição concorrente retorna HTTP 409.

## Limites

- JSON: 1 MiB.
- Multipart: 60 MiB.
- Upload de exportação: o handler aceita até 50 MiB no formulário.
