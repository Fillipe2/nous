package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nous/internal/infra/database/postgres/internal/statement"
)

// Postgres encapsula o pool de conexões e os statements SQL embarcados.
type Postgres struct {
	pool *pgxpool.Pool
}

// Connect estabelece e valida a conexão com o PostgreSQL, carregando os statements SQL.
func Connect(ctx context.Context, databaseURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsear config postgres: %w", err)
	}

	// SimpleProtocol: valores transferidos como texto, permitindo scan direto de
	// UUID → string, timestamp → time.Time sem conversores extras.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("criar pool postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := statement.Load(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("carregar statements SQL: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

// Close encerra o pool de conexões.
func (p *Postgres) Close() {
	p.pool.Close()
}

// QueryRow executa uma query que retorna no máximo uma linha.
func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

// Query executa uma query que retorna múltiplas linhas.
func (p *Postgres) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

// Exec executa uma query sem retorno de linhas.
func (p *Postgres) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}

// ExecResult executa uma query sem retorno de linhas e devolve o command tag.
func (p *Postgres) ExecResult(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}
