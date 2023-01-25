package chat

import (
	"sync"

	"github.com/google/uuid"

	"com.adoublef.wss/internal"
	websocket "com.adoublef.wss/pkg/http/websocket/gobwas"
)

type Chat struct {
	ID       uuid.UUID  `json:"id"`
	Messages []*Message `json:"messages"`

	cli *websocket.Client
}

// NOTE this should not be empty but panic if it is
func (c *Chat) Client() *websocket.Client {
	if c.cli == nil {
		panic("client is nil")
	}

	return c.cli
}

func (c *Chat) SetClient(cli *websocket.Client) {
	if c.cli == nil {
		c.cli = cli
	}
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

type Broker internal.Broker[string, *Chat]

type chatBroker struct {
	m sync.Map
}

func NewBroker() Broker {
	b := &chatBroker{}
	return b
}

func (b *chatBroker) Delete(id string) {
	b.m.Delete(id)
}

func (b *chatBroker) Load(id string) (value *Chat, ok bool) {
	v, ok := b.m.Load(id)
	return v.(*Chat), ok
}

func (b *chatBroker) Store(id string, chat *Chat) {
	b.m.Store(id, chat)
}

func (b *chatBroker) LoadOrStore(id string, chat *Chat) (*Chat, bool) {
	actual, loaded := b.m.LoadOrStore(id, chat)
	if !loaded {
		actual = chat
		actual.(*Chat).SetClient(websocket.NewClient())
	}

	return actual.(*Chat), loaded
}
