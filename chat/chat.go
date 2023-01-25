package chat

import "github.com/google/uuid"

type Chat struct {
	ID       uuid.UUID  `json:"id"`
	Messages []*Message `json:"messages"`
}

func NewChat() *Chat {
	chat := Chat{ID: uuid.New()}

	return &chat
}

func (c *Chat) String() string {
	return "chat no: " + c.ID.String()
}

type Message struct {
	ID      int    `json:"id,omitempty"`
	Content string `json:"content"`
	Chat    *Chat  `json:"-"`
}
