# API Buildnix

Base local padrão: `http://localhost:8080`.

OpenAPI embarcado: `GET /swagger/openapi.yaml`.

## Autenticação

- `POST /auth/login`: login tradicional.
- `POST /auth/refresh`: renova token JWT.
- `POST /auth/logout`: logout.
- `GET /auth/google/login`: inicia OAuth2 Google.
- `GET /auth/google/callback`: callback OAuth2 Google.

Rotas autenticadas usam header:

```http
Authorization: Bearer <token>
```

## Usuários

- `POST /usuarios`: cadastro público de usuário comum.
- `GET /usuarios/me`: dados do usuário autenticado.
- `GET /usuarios`: lista usuários, admin.
- `PUT /usuarios/{id}`: edita usuário, próprio usuário ou admin. Requer `version`.
- `DELETE /usuarios/{id}`: soft delete, próprio usuário ou admin.

## Clientes e Relacionamentos

- `GET /clientes/me`: lista clientes vinculados ao usuário autenticado, retornando apenas dados básicos.
- `GET /clientes`: lista clientes, admin.
- `POST /clientes`: cria cliente, admin.
- `PUT /clientes/{id}`: edita cliente, admin. Requer `version`.
- `DELETE /clientes/{id}`: soft delete cliente, admin.
- `POST /usuarios-clientes`: vincula usuário a cliente, admin.

## Dashboard

- `GET /dashboard?cliente_id={id}`: resumo do cliente.

O dashboard do cliente mostra usuários vinculados, landing pages criadas, landing pages online, exportações e landing page mais utilizada por exportações. Não mostra lista nem métricas de leads.

## Negócios

- `GET /negocios?cliente_id={id}`: lista negócios do cliente vinculado.
- `POST /negocios`: cria negócio para cliente vinculado.
- `PUT /negocios/{id}`: edita negócio do cliente vinculado. Requer `version`.
- `DELETE /negocios/{id}`: soft delete negócio.

## Landing Pages

- `GET /landingpages`: lista landing pages. Filtros: `cliente_id`, `negocio_id`, `categoria_id`, `status`.
- `POST /landingpages`: cria landing page.
- `GET /landingpages/{id}`: detalhe da landing page.
- `PUT /landingpages/{id}`: edita landing page. Requer `version`.
- `PATCH /landingpages/{id}/autosave`: salva rascunho. Requer `version` e retorna nova `version`.
- `PATCH /landingpages/{id}/lock`: bloqueia edição concorrente.
- `PATCH /landingpages/{id}/unlock`: desbloqueia edição.
- `DELETE /landingpages/{id}`: soft delete.

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
- `GET /templates`: lista templates.
- `POST /templates`: upload de template, admin.
- `DELETE /templates/{id}`: soft delete template, admin.

## Leads

- `POST /leads`: captura pública via formulário da landing page.
- `GET /leads`: lista leads capturados, admin/Sognix.
- `GET /leads/{id}`: detalhe do lead, admin/Sognix.
- `PATCH /leads/{id}/status`: atualiza status, admin/Sognix.

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
