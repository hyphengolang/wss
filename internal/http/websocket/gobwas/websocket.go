package websocket

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func read(conn *connHander, cli *Client) {
	defer func() {
		cli.d <- conn
		conn.rwc.Close()
	}()

	conn.setReadDeadLine(pongWait)
	// handle Pong message

	for {
		// TODO -- handle parsing the message here
		// cli.handleMessage(msg *Message) error
		msg, err := conn.readRaw()
		if err != nil {
			// handle error
			break
		}

		cli.bc <- msg
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait   = 10 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func (c *connHander) setWriteDeadLine(d time.Duration) error {
	return c.rwc.SetWriteDeadline(time.Now().Add(d))
}

func (c *connHander) setReadDeadLine(d time.Duration) error {
	return c.rwc.SetReadDeadline(time.Now().Add(d))
}

func write(conn *connHander) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.rwc.Close()
	}()

	for {
		select {
		case msg, ok := <-conn.send:
			conn.setWriteDeadLine(writeWait)
			if !ok {
				conn.log("write close")
				// write close
				conn.writeRaw(&wsutil.Message{OpCode: ws.OpClose, Payload: []byte{}})
				return
			}

			if err := conn.writeRaw(msg); err != nil {
				// handle error
				// err: write tcp [::1]:8080->[::1]:33210: i/o timeout
				conn.logf("err: %v\n", err)
				return
			}
		// TODO case ticker and pong message
		case <-ticker.C:
			conn.log("ticker")
			conn.setWriteDeadLine(writeWait)
			err := conn.writeRaw(&wsutil.Message{OpCode: ws.OpPing, Payload: nil})
			if err != nil {
				conn.logf("err: %v\n", err)
				return
			}
		}
	}
}

type Client struct {
	r, d chan *connHander
	bc   chan *wsutil.Message
	cs   map[*connHander]bool
	u    *ws.HTTPUpgrader
}

func NewClient() *Client {
	cli := &Client{
		r:  make(chan *connHander),
		d:  make(chan *connHander),
		bc: make(chan *wsutil.Message),
		cs: make(map[*connHander]bool),
		u:  &ws.HTTPUpgrader{},
	}

	go cli.listen()
	return cli
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
			// NOTE -- handle what to do with the message here
			// cli.handleMessage(msg *Message) error
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
	rwc, _, _, err := cli.u.Upgrade(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE -- for testing purposes only
	l := log.Default()

	conn := &connHander{
		rwc:  rwc,
		send: make(chan *wsutil.Message),
		log:  l.Println,
		logf: l.Printf,
	}

	cli.r <- conn

	go write(conn)
	go read(conn, cli)
}

type connHander struct {
	rwc net.Conn

	send chan *wsutil.Message

	logf func(format string, v ...any)
	log  func(v ...any)
}

func (c *connHander) controlHandler(h ws.Header, r io.Reader) error {
	switch op := h.OpCode; op {
	case ws.OpPing:
		return c.handlePing(h)
	case ws.OpPong:
		return c.handlePong(h)
	case ws.OpClose:
		return c.handleClose(h)
	}

	return wsutil.ErrNotControlFrame
}

func (c *connHander) handlePing(h ws.Header) error {
	c.log("ping")
	return nil
}

func (c *connHander) handlePong(h ws.Header) error {
	c.log("pong")
	return c.setReadDeadLine(pongWait)
}

func (c *connHander) handleClose(h ws.Header) error {
	c.log("close")
	return nil
}

func (c *connHander) readRaw() (*wsutil.Message, error) {
	r := wsutil.NewReader(c.rwc, ws.StateServerSide)

	for {
		h, err := r.NextFrame()
		if err != nil {
			return nil, err
		}

		if h.OpCode.IsControl() {
			if err := c.controlHandler(h, r); err != nil {
				return nil, err
			}
			continue
		}

		// where want = ws.OpText|ws.OpBinary
		// NOTE -- eq: h.OpCode != 0 && h.OpCode != want
		if want := (ws.OpText | ws.OpBinary); h.OpCode&want == 0 {
			if err := r.Discard(); err != nil {
				return nil, err
			}
			continue
		}

		p, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return &wsutil.Message{OpCode: h.OpCode, Payload: p}, nil
	}
}

func (c *connHander) writeRaw(msg *wsutil.Message) error {
	// This is server-side
	frame := ws.NewFrame(msg.OpCode, true, msg.Payload)
	return ws.WriteFrame(c.rwc, frame)
}

// Recommended to use connHander.readRaw instead
// func (c *connHander) read() (*wsutil.Message, error) {
// 	p, op, err := wsutil.ReadClientData(c.rwc)
// 	if err != nil {
// 		return nil, err
// 	}
// 	c.logf("message: %v\n", op)

// 	return &wsutil.Message{OpCode: op, Payload: p}, nil
// }

// Recommended to use connHander.writeRaw instead
// func (c *connHander) write(msg *wsutil.Message) error {
// 	return wsutil.WriteServerMessage(c.rwc, msg.OpCode, msg.Payload)
// }
