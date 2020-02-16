package ui

import (
	"proxio/client"
)

func NewConnectionPool() *Pool {
	pool := &Pool{
		Connections: []*Connection{},
		closeChan:   make(chan *Connection),
	}

	go func() {
		for conn := range pool.closeChan {
			pool.removeConnection(conn)
		}
	}()

	return pool
}

type Pool struct {
	Connections []*Connection
	closeChan   chan *Connection
}

func (p *Pool) NewConnection(conn *Connection) {
	p.Connections = append(p.Connections, conn)
}

func (p *Pool) removeConnection(conn *Connection) {
	for i, c := range p.Connections {
		if c == conn {
			p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
		}
	}
}

func (p *Pool) BroadcastMessage(message *client.MessageContent) {
	for _, conn := range p.Connections {
		if err := conn.Send(message); err != nil {
			println("error sending message", p.Connections)
		}
	}
}
