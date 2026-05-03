package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// ClienteRepository implementa cliente.Repository para PostgreSQL.
type ClienteRepository struct {
	connection *Postgres
}

var _ cliente.Repository = (*ClienteRepository)(nil)

// NewClienteRepository constrói o repositório de clientes.
func NewClienteRepository(connection *Postgres) (*ClienteRepository, error) {
	return &ClienteRepository{connection: connection}, nil
}

func (r *ClienteRepository) Salvar(ctx context.Context, c *cliente.Cliente) error {
	c.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarCliente, pgx.NamedArgs{
		"id":            c.Id,
		"nome":          c.Nome,
		"email":         c.Email,
		"telefone":      c.Telefone,
		"whatsapp":      c.Whatsapp,
		"status":        string(c.Status),
		"criado_em":     c.CriadoEm,
		"atualizado_em": c.AtualizadoEm,
		"version":       c.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar cliente: %w", err)
	}
	return nil
}

func (r *ClienteRepository) BuscarPorID(ctx context.Context, id string) (*cliente.Cliente, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarClientePorID, pgx.NamedArgs{"id": id})
	c, err := scanner.Cliente(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, cliente.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar cliente por id: %w", err)
	}
	return c, nil
}

func (r *ClienteRepository) Atualizar(ctx context.Context, c *cliente.Cliente) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarCliente, pgx.NamedArgs{
		"nome":             c.Nome,
		"email":            c.Email,
		"telefone":         c.Telefone,
		"whatsapp":         c.Whatsapp,
		"status":           string(c.Status),
		"atualizado_em":    c.AtualizadoEm,
		"version":          c.Version,
		"version_anterior": c.Version - 1,
		"id":               c.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar cliente: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return cliente.ErrVersaoConflito
	}
	return nil
}

func (r *ClienteRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteCliente, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete cliente: %w", err)
	}
	return nil
}

func (r *ClienteRepository) Listar(ctx context.Context, filtro cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT c.id, c.nome, c.email, c.telefone, c.whatsapp, c.status, c.criado_em, c.atualizado_em, c.version FROM cliente c`
	args := []any{}
	argIdx := 1
	if filtro.UsuarioId != "" {
		query += " INNER JOIN usuario_cliente uc ON uc.cliente_id=c.id"
	}
	query += " WHERE 1=1"
	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND uc.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}

	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND c.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Email != "" {
		query += fmt.Sprintf(" AND c.email ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Email+"%")
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND c.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND c.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND c.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar clientes: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "c.nome", "email": "c.email", "status": "c.status", "criado_em": "c.criado_em", "atualizado_em": "c.atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar clientes: %w", err)
	}
	defer rows.Close()

	var clientes []*cliente.Cliente
	for rows.Next() {
		c, err := scanner.Cliente(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear cliente: %w", err)
		}
		clientes = append(clientes, c)
	}
	return clientes, total, nil
}

func (r *ClienteRepository) VincularUsuario(ctx context.Context, uc *cliente.UsuarioCliente) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.VincularUsuarioCliente, pgx.NamedArgs{
		"usuario_id": uc.UsuarioId,
		"cliente_id": uc.ClienteId,
		"criado_em":  uc.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("vincular usuario cliente: %w", err)
	}
	return nil
}

func (r *ClienteRepository) ListarPorUsuario(ctx context.Context, usuarioId string) ([]*cliente.Cliente, error) {
	stmts := statement.Get()
	rows, err := r.connection.Query(ctx, stmts.ListarClientesPorUsuario, pgx.NamedArgs{"usuario_id": usuarioId})
	if err != nil {
		return nil, fmt.Errorf("listar clientes por usuario: %w", err)
	}
	defer rows.Close()

	var clientes []*cliente.Cliente
	for rows.Next() {
		c, err := scanner.Cliente(rows)
		if err != nil {
			return nil, fmt.Errorf("escanear cliente: %w", err)
		}
		clientes = append(clientes, c)
	}
	return clientes, nil
}

func (r *ClienteRepository) VerificarAcessoUsuario(ctx context.Context, clienteId, usuarioId string) error {
	stmts := statement.Get()
	var acesso bool
	if err := r.connection.QueryRow(ctx, stmts.VerificarAcessoCliente, pgx.NamedArgs{
		"cliente_id": clienteId,
		"usuario_id": usuarioId,
	}).Scan(&acesso); err != nil {
		return fmt.Errorf("verificar acesso cliente: %w", err)
	}
	if !acesso {
		return cliente.ErrNaoAutorizado
	}
	return nil
}
