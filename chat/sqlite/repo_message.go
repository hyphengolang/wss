package repo

import (
	"context"
	"database/sql"

	"com.adoublef.wss/chat"
)

type msgRepo struct {
	db *sql.DB
}

type MsgRepo Repo[*chat.Message]

func NewMsgRepo(db *sql.DB) MsgRepo {
	r := &msgRepo{db: db}

	return r
}

func (r *msgRepo) Create(ctx context.Context, msg *chat.Message) error {
	q := `INSERT INTO "messages" (chat_id, content) VALUES (?, ?)`

	_, err := r.db.ExecContext(ctx, q, "msg.ChatID", msg.Content)
	return err
}

func (r *msgRepo) Delete(ctx context.Context, key any) error {
	q := `DELETE FROM "messages" WHERE id = ?`

	_, err := r.db.ExecContext(ctx, q, key)
	return err
}

func (r *msgRepo) Find(ctx context.Context, key any) (*chat.Message, error) {
	q := `SELECT content FROM "messages" WHERE id = ?`
	var msg Message

	err := r.db.QueryRowContext(ctx, q, key).Scan(&msg.Content)
	if err != nil {
		return nil, err
	}

	return &chat.Message{
		ID:      msg.ID,
		Content: msg.Content,
	}, nil
}

// NOTE - should be pulling messages from a unique chat
func (r *msgRepo) FindMany(ctx context.Context) ([]*chat.Message, error) {
	q := `SELECT content FROM "messages"`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	var msgs []*chat.Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Content); err != nil {
			return nil, err
		}

		msgs = append(msgs, &chat.Message{Content: msg.Content})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return msgs, rows.Close()
}
