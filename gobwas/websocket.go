package websocket

import (
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Broker struct {
	cs map[string]*Client
}

func NewBroker() *Broker {
	b := &Broker{
		cs: make(map[string]*Client),
	}

	return b
}

func (b *Broker) Find(id string) (*Client, bool) {
	cli, ok := b.cs[id]
	return cli, ok
}

func (b *Broker) Add(id string, cli *Client) {
	b.cs[id] = cli
}

func read(conn *connHander, cli *Client) {
	defer func() {
		cli.d <- conn
		conn.rwc.Close()
	}()

	for {
		msg, err := conn.read()
		if err != nil {
			// handle error
			return
		}

		cli.bc <- msg
	}
}

func write(conn *connHander) {
	defer func() {
		conn.rwc.Close()
	}()

	for {
		select {
		case msg, ok := <-conn.send:
			if !ok {
				// write close
				return
			}

			if err := conn.write(msg); err != nil {
				// handle error
				return
			}
			// TODO case ticker and pong message
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
	log.Println("gobwas: NewClient")

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

	conn := &connHander{
		rwc:  rwc,
		send: make(chan *wsutil.Message),
	}

	cli.r <- conn

	go write(conn)
	go read(conn, cli)
}

type connHander struct {
	rwc net.Conn

	send chan *wsutil.Message
}

func (c *connHander) read() (*wsutil.Message, error) {
	p, op, err := wsutil.ReadClientData(c.rwc)
	if err != nil {
		return nil, err
	}

	return &wsutil.Message{OpCode: op, Payload: p}, nil
}

func (c *connHander) write(msg *wsutil.Message) error {
	return wsutil.WriteServerMessage(c.rwc, msg.OpCode, msg.Payload)
}
