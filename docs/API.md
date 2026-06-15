# API Nous

Base local padrão: `http://localhost:8080`.

OpenAPI embarcado: `GET /swagger/openapi.yaml`.

## Autenticação

- `POST /auth/login`: login tradicional.
- `POST /auth/refresh`: renova token JWT.
- `POST /auth/logout`: logout.
- `POST /auth/esqueci-senha`: solicita redefinição de senha com resposta genérica.
- `POST /auth/redefinir-senha`: redefine senha usando token de uso único e revoga refresh tokens ativos do usuário.
- `POST /auth/solicitar-verificacao-email`: gera token de verificação para usuário ativo e não verificado (resposta neutra). Sem provedor de email, o token só volta na resposta com `EMAIL_VERIFICATION_RETURN_TOKEN=true`.
- `POST /auth/verificar-email`: confirma o email com token de uso único; marca `email_verified=true`.
- `GET /auth/google/login`: inicia OAuth2 Google.
- `GET /auth/google/callback`: callback OAuth2 Google.

Rotas autenticadas usam header:

```http
Authorization: Bearer <token>
```

## Usuários

- `POST /usuarios`: cadastro público de usuário. O papel é sempre `usuario`; admin não pode ser criado por aqui. Email duplicado retorna 409.
- `GET /usuarios/me`: dados do usuário autenticado.
- `PATCH /usuarios/me/senha`: troca a senha do usuário autenticado informando senha atual e nova senha.
- `GET /usuarios`: lista usuários, admin.
- `PUT /usuarios/{id}`: edita nome e email, próprio usuário ou admin. Alterar o email zera `email_verified`. Email duplicado retorna 409.
- `PATCH /usuarios/{id}/role`: promove/rebaixa o papel entre `usuario` e `admin`, exclusivo para admin. Não permite rebaixar o último administrador ativo (409).
- `PATCH /usuarios/{id}/status`: ativa/inativa o usuário (`ativo`/`inativo`), exclusivo para admin. Não aceita `excluido` (use DELETE) e não permite inativar o último administrador ativo (409).
- `DELETE /usuarios/{id}`: soft delete, próprio usuário ou admin. O email de um usuário excluído pode ser reutilizado num novo cadastro.

O usuário expõe `email_verified` (true em login Google e no admin do seed) e `ultimo_login_em` (atualizado a cada login).

Filtros de data em `GET /usuarios` usam apenas data no formato `YYYY-MM-DD`; `data_fim` é inclusiva para o dia inteiro.

## Negocios e Relacionamentos

- `GET /negocios/me`: lista negocios vinculados ao usuário autenticado, retornando apenas dados básicos.
- `GET /negocios`: lista negocios, admin.
- `POST /negocios`: cria negocio via self-onboarding autenticado com nome, email, telefone e whatsapp, vinculando automaticamente ao usuário criador.
- `PUT /negocios/{id}`: edita negocio, admin. Requer `version`.
- `DELETE /negocios/{id}`: soft delete negocio, admin.

## Dashboard

- `GET /dashboard`: dashboard do usuário autenticado. Admin recebe visão global da operação e pode recortar por `negocio_id`, `usuario_id`, `landingpage_id`, `unidade_id`, `categoria_id`, `template_id` e `tipo`. Usuário comum recebe métricas agregadas dos negocios vinculados ao próprio usuário, sem precisar enviar `negocio_id`.

Filtros de data usam `YYYY-MM-DD`, com `data_fim` inclusiva para o dia inteiro. Dados pessoais de leads continuam restritos às rotas de leads admin-only; no dashboard comum não há listagem nem dados pessoais de leads.

## Opções para combos e filtros

Use os endpoints `/opcoes/*` para popular combos, selects e filtros do frontend. Eles retornam itens leves no formato `{ id, nome, label, status, tipo, origem, negocio_id, categoria_id }`, com `total`, `pagina` e `por_pagina`. A busca textual usa `q` e a ordenação padrão é por nome crescente. O frontend deve adicionar localmente a opção visual "Todos" quando quiser representar ausência de filtro.

- `GET /opcoes/negocios`: negocios disponíveis para o usuário. Admin/Sognix vê todos; usuário comum vê apenas os vinculados.
- `GET /opcoes/usuarios`: usuários para filtros administrativos, admin.
- `GET /opcoes/unidades`: unidades disponíveis. Admin/Sognix vê todos; usuário comum vê apenas unidades dos negocios vinculados.
- `GET /opcoes/categorias`: categorias disponíveis.
- `GET /opcoes/templates`: templates disponíveis. Usuário comum deve informar `negocio_id`; o backend combina templates Sognix e templates do negocio, respeitando o plano.
- `GET /opcoes/landingpages`: landing pages disponíveis. Admin/Sognix vê todas; usuário comum vê apenas LPs dos negocios vinculados.
- `GET /opcoes/enums`: tipos e status estáticos para filtros, como tipos de LP, status de lead, origem de template e tipos de template.

Exemplo para combo de negocio:

```http
GET /opcoes/negocios?q=clinica&por_pagina=100
```

Exemplo para combo de template após selecionar negocio e categoria:

```http
GET /opcoes/templates?negocio_id={negocio_id}&categoria_id={categoria_id}&tipo=white_label
```

## Unidades

- `GET /unidades?negocio_id={id}`: lista unidades do negocio vinculado.
- `POST /unidades`: cria unidade para negocio vinculado.
- `PUT /unidades/{id}`: edita unidade do negocio vinculado. Requer `version`.
- `DELETE /unidades/{id}`: soft delete unidade.

## Landing Pages

- `GET /landingpages`: lista landing pages. Filtros: `negocio_id`, `unidade_id`, `categoria_id`, `tipo`, `status`.
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
- `customizavel`: permite iniciar com `template_id` e depois customizar `conteudo`/`autosave_data`, desde que o plano ativo do negocio tenha `landingpages_customizaveis=true`.

Exemplo white label:

```json
{
  "negocio_id": "uuid",
  "template_id": "uuid",
  "tipo": "white_label",
  "nome": "LP Institucional"
}
```

Exemplo customizável começando de um template:

```json
{
  "negocio_id": "uuid",
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
- `GET /templates`: lista templates. Filtros: `negocio_id`, `categoria_id`, `tipo`, `origem`.
- `POST /templates`: cadastra template pronto no banco. Admin pode criar template oficial Sognix ou template de negocio; usuário comum só pode criar template do negocio ao qual tem acesso. Aceita multipart com `negocio_id`, `origem`, `nome`, `tipo`, `categoria_id`, `conteudo` e/ou `arquivo`. Para white label, envie o `.zip` em `arquivo`: ele é validado (estrutura, zip-slip, limite de descompressão) e armazenado em binário. Arquivos de texto precisam ser UTF-8 válido e são salvos em `conteudo`.
- `GET /templates/{id}/download`: baixa o pacote `.zip` do template. Templates Sognix ficam disponíveis a qualquer usuário autenticado; templates de negocio exigem acesso ao negocio dono. Retorna `404` se o template não tiver arquivo binário.
- `DELETE /templates/{id}`: soft delete template, admin.

Na tela de criação/edição de LP, use:

```http
GET /templates?negocio_id={negocio_id}
```

Essa listagem retorna templates `origem=sognix` e templates `origem=negocio` daquele negocio. Se o plano do negocio não permitir customizável, templates `base_customizavel` são omitidos.

Origens de template:

- `sognix`: template oficial criado pela Sognix, global para negocios autorizados pelo plano.
- `negocio`: template criado/salvo por um negocio específico, visível apenas para aquele negocio e admin.

Tipos de template:

- `white_label`: disponível para landing pages white label e também como ponto de partida de customizáveis.
- `base_customizavel`: base para negocios com plano customizável; não pode ser usada para criar LP `white_label`.

## Leads

- `POST /leads`: captura pública via formulário da landing page.
- `GET /leads`: lista leads capturados, somente admin/Sognix. Aceita `negocio_id` para ver leads dos negocios cadastrados.
- `GET /leads/{id}`: detalhe do lead, somente admin/Sognix.
- `PATCH /leads/{id}/status`: atualiza status, somente admin/Sognix.

Leads são dados pessoais e ativo comercial da Sognix. Usuário comum não tem acesso nominal nem métricas de leads.

## Exportações

- `POST /landingpages/{id}/exportar`: upload multipart do ZIP exportado pelo front.
- `GET /exportacoes?landingpage_id={id}`: lista exportações de uma landing page.
- `GET /exportacoes/{id}/download`: download seguro do ZIP.

Exportações respeitam o vínculo do usuário com o negocio dono da landing page.

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
