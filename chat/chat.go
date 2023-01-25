package chat

import "github.com/google/uuid"

type Chat struct {
	ID uuid.UUID `json:"id"`
}

func NewChat() *Chat {
	chat := Chat{ID: uuid.New()}

	return &chat
}

func (c *Chat) String() string {
	return "chat no: " + c.ID.String()
}
