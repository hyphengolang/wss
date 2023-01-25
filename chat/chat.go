package chat

import (
	"sync"

	"github.com/google/uuid"

	websocket "com.adoublef.wss/gobwas"
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

type Broker[K, V any] interface {
	Load(key K) (value V, ok bool)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	Store(key K, value V)
	Delete(key K)
}

type ChatBroker Broker[string, *Chat]

type chatBroker struct {
	mu sync.Mutex
	cs map[string]*Chat
}

func NewBroker() ChatBroker {
	b := &chatBroker{
		cs: make(map[string]*Chat),
	}

	return b
}

// Delete implements ChatBroker
func (b *chatBroker) Delete(key string) {
	b.mu.Lock()
	{
		delete(b.cs, key)
		// NOTE - will need to delete the client as well
	}
	b.mu.Unlock()
}

// Load implements ChatBroker
func (b *chatBroker) Load(key string) (value *Chat, ok bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	cli, ok := b.cs[key]
	return cli, ok
}

// Store implements ChatBroker
func (b *chatBroker) Store(key string, c *Chat) {
	b.mu.Lock()
	{
		b.cs[key] = c
	}
	b.mu.Unlock()
}

// LoadOrStore implements ChatBroker
func (b *chatBroker) LoadOrStore(key string, c *Chat) (actual *Chat, loaded bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, loaded = b.cs[key]; !loaded {
		b.cs[key] = c
	}

	actual = b.cs[key]
	actual.SetClient(websocket.NewClient())

	return actual, loaded
}
