package ws

import (
	"encoding/json"
	"github.com/gobwas/ws"
	"net"
	"proxio/traffic"
)

func NewConnectionPool() *Pool {
	pool := &Pool{
		DomainListeners: make(map[string][]*Connection, 0),
		CloseChan:       make(chan *Connection),
	}

	go func() {
		for conn := range pool.CloseChan {
			pool.removeConnection(conn)
		}
	}()

	return pool
}

type Pool struct {
	DomainListeners map[string][]*Connection
	CloseChan       chan *Connection
}

func (p *Pool) NewConnection(domain string, conn *Connection) {
	if _, ok := p.DomainListeners[domain]; !ok {
		p.DomainListeners[domain] = []*Connection{}
	}
	p.DomainListeners[domain] = append(p.DomainListeners[domain], conn)
}

func (p *Pool) removeConnection(conn *Connection) {
	for domain, connections := range p.DomainListeners {
		for i, c := range connections {
			if c == conn {
				p.DomainListeners[domain] = append(p.DomainListeners[domain][:i], p.DomainListeners[domain][i+1:]...)
			}
		}
	}
}

func (p *Pool) BroadcastMessage(message *traffic.MessageContent) {
	connections, ok := p.DomainListeners[message.Domain]
	if !ok {
		return
	}
	for _, conn := range connections {
		if err := conn.Send(message); err != nil {
			println("error sending message", p.DomainListeners)
		}
	}
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{conn: conn, messages: make(chan *traffic.MessageContent)}
}

type Connection struct {
	conn     net.Conn
	messages chan *traffic.MessageContent
}

func (c *Connection) Send(m *traffic.MessageContent) error {
	payload, err := json.Marshal(m)
	if err != nil {
		panic("unable to encode frame")
	}
	frame := ws.NewTextFrame(payload)

	if err = ws.WriteFrame(c.conn, frame); err != nil {
		return err
	}
	return nil
}
