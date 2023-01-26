package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conn[T any] interface {
	Conn() *pgxpool.Pool
	ExecContext(ctx context.Context, query string, args pgx.NamedArgs) error
	QueryRowContext(ctx context.Context, scanner func(row pgx.Row, t *T) error, query string, args pgx.NamedArgs) (*T, error)
	QueryContext(ctx context.Context, scanner func(row pgx.Rows, t *T) error, query string, args pgx.NamedArgs) ([]*T, error)
}

type C interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type connHandler[T any] struct {
	conn *pgxpool.Pool
}

func NewHandler[T any](conn *pgxpool.Pool) Conn[T] {
	return &connHandler[T]{conn: conn}
}

func (h *connHandler[T]) Conn() *pgxpool.Pool {
	return h.conn
}

func (h *connHandler[T]) ExecContext(ctx context.Context, query string, args pgx.NamedArgs) error {
	return ExecContext(ctx, h.conn, query, args)
}

func (h *connHandler[T]) QueryRowContext(ctx context.Context, scanner func(row pgx.Row, t *T) error, query string, args pgx.NamedArgs) (*T, error) {
	return QueryRowContext(ctx, h.conn, scanner, query, args)
}

func (h *connHandler[T]) QueryContext(ctx context.Context, scanner func(row pgx.Rows, t *T) error, query string, args pgx.NamedArgs) ([]*T, error) {
	return QueryContext(ctx, h.conn, scanner, query, args)
}

func ExecContext(ctx context.Context, q *pgxpool.Pool, query string, args ...any) error {
	_, err := q.Exec(ctx, query, args...)
	return err
}

func QueryRowContext[T any](ctx context.Context, q *pgxpool.Pool, scanner func(r pgx.Row, t *T) error, query string, args ...any) (*T, error) {
	var t T
	err := scanner(q.QueryRow(ctx, query, args...), &t)
	return &t, err
}

func QueryContext[T any](ctx context.Context, q *pgxpool.Pool, scanner func(r pgx.Rows, v *T) error, query string, args ...any) ([]*T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []*T
	for rows.Next() {
		var v T
		err = scanner(rows, &v)
		if err != nil {
			return nil, err
		}
		vs = append(vs, &v)
	}
	return vs, rows.Err()
}

type ReadWriter[T any] interface {
	Reader[T]
	Writer[T]
}

type Reader[T any] interface {
	Find(ctx context.Context, key any) (T, error)
	FindMany(ctx context.Context) ([]T, error)
}

type Writer[T any] interface {
	Create(ctx context.Context, t T) error
}
