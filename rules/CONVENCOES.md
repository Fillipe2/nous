# Convenções Nous

## Idioma

O código do Nous usa português nos nomes de domínio, funções, mensagens e documentação, preservando termos técnicos inevitáveis de bibliotecas e protocolo.

Evite misturar inglês e português no mesmo identificador ou mensagem.

## Estrutura e Dependências

Fluxo obrigatório:

```text
Interface -> App -> Domain <- Infra
```

Regras:

- `domain` não importa `app`, `infra`, `interface`, banco ou HTTP.
- `app` importa `domain`, mas não importa `infra` nem `interface`.
- `infra` importa `domain` para implementar interfaces, mas não importa `app` nem `interface`.
- `interface` importa `app` e `domain`, mas não acessa `infra` diretamente.
- `cmd` deve montar dependências via `cmd/internal/bootstrap` e usar apenas pacotes públicos da interface.

## Context

- `context.Context` sempre é o primeiro parâmetro.
- Nunca armazenar `context.Context` em struct.
- Propagar contexto em banco, HTTP externo, filas, cache e operações de I/O.
- Handlers usam `r.Context()`.

## Erros

- Erros de domínio ficam no pacote do domínio como sentinelas `Err...`.
- Infra pode ter erros próprios quando necessário.
- Use `fmt.Errorf("contexto da operação: %w", err)` para wrap.
- Compare erros com `errors.Is()` ou `errors.As()`.
- Não retorne valor útil e erro ao mesmo tempo.
- Não use comparação por string de erro.

## SQL e Repositórios

- SQL em arquivos `.sql` dentro de `internal/infra/database/postgres/internal/statement/`.
- Nomear queries com `-- name: NomeQuery`.
- Usar parâmetros nomeados `@param_name` e `pgx.NamedArgs`.
- Scanners ficam em `internal/infra/database/postgres/internal/scanner/`.
- Repositórios devem declarar assertiva de interface, por exemplo: `var _ negocio.Repository = (*NegocioRepository)(nil)`.
- Não usar `sqlc`.
- Não colocar SQL transacional como string hardcoded no repositório quando puder virar statement nomeado.

## Concorrência

- Entidades editáveis devem expor e receber `version`.
- Updates devem validar `version` no banco.
- Conflito de versão deve retornar erro de domínio e HTTP 409.
- Locks em memória não protegem múltiplas instâncias; use PostgreSQL para locks/condições atômicas.

## HTTP

- Handlers usam early return.
- Handlers decodificam input, chamam use case e traduzem erro de domínio para HTTP.
- Respostas de erro devem manter `erro`, `codigo`, `mensagem` e `request_id`.
- Rotas novas devem atualizar `internal/interface/webapi/internal/router/openapi.yaml`.

## Build

Binários devem ir para `bin/`:

```bash
go build -o bin/nous-webapi ./cmd/webapi
```

Não compilar com caminho de arquivo como `go build ./cmd/webapi/main.go`.

## Testes

Comandos padrão:

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
```

Testes de contrato devem cobrir OpenAPI e rotas críticas. Testes de carga devem cobrir endpoints públicos, autosave e exportação antes de produção.

## OpenAPI

- Manter `GET /swagger/openapi.yaml` funcionando.
- Atualizar OpenAPI quando endpoint, payload ou código de erro mudar.

## Proibições Técnicas

- Não criar novo entrypoint (`webhook`, `worker`, `cli`) sem necessidade real do domínio.
- Não acoplar Nous diretamente ao banco de outro produto.
- Não expor leads para usuário comum.
- Não quebrar operação standalone quando Observer estiver vazio.
- Não salvar segredos reais em arquivos versionados.
