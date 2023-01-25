package chat

import (
	"sync"

	"github.com/google/uuid"

	websocket "com.adoublef.wss/gobwas"
	"com.adoublef.wss/internal"
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

// Material
//
// https://www.cs.cmu.edu/~410-s05/lectures/L31_LockFree.pdf
type chatBroker struct {
	m sync.Map
}

func NewBroker() Broker {
	b := &chatBroker{}

	return b
}

// Delete implements ChatBroker
func (b *chatBroker) Delete(key string) {
	b.m.Delete(key)
}

// Load implements ChatBroker
func (b *chatBroker) Load(key string) (value *Chat, ok bool) {
	v, ok := b.m.Load(key)
	return v.(*Chat), ok
}

// Store implements ChatBroker
func (b *chatBroker) Store(key string, c *Chat) {
	b.m.Store(key, c)
}

// LoadOrStore implements ChatBroker
func (b *chatBroker) LoadOrStore(key string, chat *Chat) (*Chat, bool) {
	actual, loaded := b.m.LoadOrStore(key, chat)
	// does not exist
	if !loaded {
		actual = chat
		actual.(*Chat).SetClient(websocket.NewClient())
	}

	return actual.(*Chat), loaded
}
