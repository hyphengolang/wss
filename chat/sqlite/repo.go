package repo

import (
	"context"
	"database/sql"

	"com.adoublef.wss/chat"
)

type Repo interface {
	Create(ctx context.Context, chat *chat.Chat) error
	Find(ctx context.Context, key any) (*chat.Chat, error)
	FindMany(ctx context.Context) ([]*chat.Chat, error)
}

var _ Repo = (*repo)(nil)

type repo struct {
	db *sql.DB
}

func (r *repo) Create(ctx context.Context, chat *chat.Chat) error {
	q := `INSERT INTO "chats" (id) VALUES (?)`

	_, err := r.db.ExecContext(ctx, q, chat.ID)
	return err
}

func (r *repo) Find(ctx context.Context, key any) (*chat.Chat, error) {
	q := `SELECT id FROM "chats" WHERE id = ?`
	var chat chat.Chat

	err := r.db.QueryRowContext(ctx, q, key).Scan(&chat.ID)
	return &chat, err
}

func (r *repo) FindMany(ctx context.Context) ([]*chat.Chat, error) {
	q := `SELECT id FROM "chats"`

	row, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	cs := []*chat.Chat{}
	for row.Next() {
		var c chat.Chat

		if err := row.Scan(&c.ID); err != nil {
			return nil, err
		}

		cs = append(cs, &c)
	}

	return cs, row.Err()
}

func NewRepo(conn *sql.DB) Repo {
	r := repo{
		db: conn,
	}

	return &r
}
