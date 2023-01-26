package repo

import (
	"context"

	comms "com.adoublef.wss/internal/communications"
	repo "com.adoublef.wss/internal/communications/sql"
	pg "com.adoublef.wss/internal/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Thread struct {
	ID uuid.UUID
}

type ThreadRepo repo.Repo[*comms.Thread]

type threadRepo struct {
	h pg.Conn[Thread]
}

func (r *threadRepo) Delete(ctx context.Context, key any) error {
	// NOTE -- key needs to be type-checked
	const q = `DELETE FROM communications.thread WHERE id = @id`

	args := pgx.NamedArgs{
		"id": key,
	}

	return r.h.ExecContext(ctx, q, args)
}

func (r *threadRepo) Create(ctx context.Context, chat *comms.Thread) error {
	const q = `INSERT INTO communications.thread (id) VALUES (@id)`

	args := pgx.NamedArgs{
		"id": chat.ID,
	}

	return r.h.ExecContext(ctx, q, args)
}

func (r *threadRepo) Find(ctx context.Context, key any) (*comms.Thread, error) {
	const q = `SELECT id FROM communications.thread WHERE id = @id`
	args := pgx.NamedArgs{"id": key}

	var thread comms.Thread
	_, err := r.h.QueryRowContext(ctx, func(row pgx.Row, thr *Thread) error {
		if err := row.Scan(&thr.ID); err != nil {
			return err
		}

		thread = comms.Thread{
			ID: thr.ID,
		}

		return nil
	}, q, args)
	return &thread, err
}

func (r *threadRepo) FindMany(ctx context.Context) ([]*comms.Thread, error) {
	const q = `SELECT id FROM communications.thread`

	var tt []*comms.Thread
	_, err := r.h.QueryContext(ctx, func(rows pgx.Rows, thr *Thread) error {
		if err := rows.Scan(&thr.ID); err != nil {
			return err
		}

		thread := comms.Thread{
			ID: thr.ID,
		}

		tt = append(tt, &thread)
		return nil
	}, q, nil)
	return tt, err
}

func NewChatRepo(conn *pgxpool.Pool) ThreadRepo {
	r := &threadRepo{h: pg.NewHandler[Thread](conn)}

	return r
}
