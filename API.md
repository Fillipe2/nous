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
- `POST /landingpages/{id}/exportar`: upload multipart do ZIP exportado pelo front.
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
- `POST /templates`: upload de template (multipart). Envie `conteudo` (HTML/texto) ou `arquivo` — para white label, um `.zip` validado e armazenado em binário.
- `GET /templates/{id}/download`: baixa o `.zip` do template. Sognix é público a usuários autenticados; template de cliente exige acesso ao cliente dono.
- `DELETE /templates/{id}`: soft delete template, admin.

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
