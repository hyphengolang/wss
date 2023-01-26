package repo

import (
	"context"
	"database/sql"

	intern "com.adoublef.wss/internal/communications"
	repo "com.adoublef.wss/internal/communications/sql"
)

type ThreadRepo repo.Repo[*intern.Thread]

var _ ThreadRepo = (*chatRepo)(nil)

type chatRepo struct {
	db *sql.DB
}

func (r *chatRepo) Delete(ctx context.Context, key any) error {
	q := `DELETE FROM "chats" WHERE id = ?`

	_, err := r.db.ExecContext(ctx, q, key)
	return err
}

func (r *chatRepo) Create(ctx context.Context, chat *intern.Thread) error {
	q := `INSERT INTO "chats" (id) VALUES (?)`

	_, err := r.db.ExecContext(ctx, q, chat.ID)
	return err
}

func (r *chatRepo) Find(ctx context.Context, key any) (*intern.Thread, error) {
	q1 := `SELECT id FROM "chats" WHERE id = ?`
	var chat intern.Thread = intern.Thread{
		Messages: make([]*intern.Message, 0),
	}

	err := r.db.QueryRowContext(ctx, q1, key).Scan(&chat.ID)
	if err != nil {
		return nil, err
	}

	// populate messages
	q2 := `SELECT content FROM "messages" WHERE chat_id = ?`
	rows, err := r.db.QueryContext(ctx, q2, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg intern.Message

		if err := rows.Scan(&msg.Content); err != nil {
			return nil, err
		}

		chat.Messages = append(chat.Messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepo) FindMany(ctx context.Context) ([]*intern.Thread, error) {
	q := `SELECT id FROM "chats"`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cs := []*intern.Thread{}
	for rows.Next() {
		var c intern.Thread

		if err := rows.Scan(&c.ID); err != nil {
			return nil, err
		}

		cs = append(cs, &c)
	}

	return cs, rows.Err()
}

func NewChatRepo(conn *sql.DB) ThreadRepo {
	r := chatRepo{
		db: conn,
	}

	return &r
}
