package websocket

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	typ  int
	data []byte
}

// Client maintains the set of active clients and broadcasts messages to the
// clients.
//
// Text messages are guaranteed to be work. Data messages are not guaranteed.
type Client struct {
	cs   map[*connHandler]bool
	bc   chan Message
	r, d chan *connHandler
	u    *websocket.Upgrader
}

func NewClient(read, write int) *Client {
	c := &Client{
		bc: make(chan Message),
		r:  make(chan *connHandler),
		d:  make(chan *connHandler),
		cs: make(map[*connHandler]bool),
		u:  &websocket.Upgrader{ReadBufferSize: read, WriteBufferSize: write},
	}

	go c.listen()
	return c
}

func (cli *Client) listen() {
	for {
		select {
		case conn := <-cli.r:
			cli.cs[conn] = true
		case conn := <-cli.d:
			delete(cli.cs, conn)
			close(conn.send)
		case msg := <-cli.bc:
			for conn := range cli.cs {
				select {
				case conn.send <- msg:
				default:
					close(conn.send)
					delete(cli.cs, conn)
				}
			}
		}
	}
}

func (cli *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rwc, err := cli.u.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := &connHandler{
		rwc:        rwc,
		send:       make(chan Message, 256),
		writeWait:  10 * time.Second,
		pongWait:   60 * time.Second,
		pingPeriod: (60 * time.Second * 9) / 10,
		readLimit:  512,
	}
	cli.r <- conn

	go write(conn)
	go read(conn, cli)
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// connHandler is a middleman between the websocket connection and the hub.
type connHandler struct {
	rwc *websocket.Conn
	// Buffered channel of outbound messages.
	send chan Message

	writeWait, pingPeriod, pongWait time.Duration
	readLimit                       int64
}

func (c *connHandler) read() (Message, error) {
	typ, msg, err := c.rwc.ReadMessage()
	if err != nil {
		return Message{websocket.CloseMessage, nil}, err
	}
	msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
	return Message{typ: typ, data: msg}, err
}

// read pumps messages from the websocket connection to the hub.
//
// The application runs read in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func read(c *connHandler, cli *Client) {
	defer func() {
		cli.d <- c
		c.rwc.Close()
	}()

	c.rwc.SetWriteDeadline(time.Now().Add(c.pongWait))
	c.rwc.SetReadLimit(c.readLimit)
	// NOTE -- unsure if this is needed
	c.rwc.SetPongHandler(func(string) error { return c.rwc.SetWriteDeadline(time.Now().Add(c.pongWait)) })

	for {
		msg, err := c.read()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		cli.bc <- msg
	}
}

// write pumps messages from the hub to the websocket connection.
//
// A goroutine running write is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func write(c *connHandler) {
	// NOTE add as field of connHandler
	// c.pingPeriod
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.rwc.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.rwc.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				// The hub closed the channel.
				c.rwc.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// NOTE - assumed that this is of type `Text`
			w, err := c.rwc.NextWriter(msg.typ)
			if err != nil {
				return
			}
			w.Write(msg.data)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write((<-c.send).data)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.rwc.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.rwc.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
