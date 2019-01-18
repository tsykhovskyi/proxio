package ui

import (
	"proxio/proxy"
)

func NewConnectionPool() *Pool {
	return &Pool{
		Connections: []*Connection{},
	}
}

type Pool struct {
	Connections []*Connection
}

func (p *Pool) NewConnection(conn *Connection) {
	p.Connections = append(p.Connections, conn)
}

func (p *Pool) RemoveConnection(i int) {
	p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
}

func (p *Pool) BroadcastMessage(message *proxy.MessageContent) {
	for i, conn := range p.Connections {
		if err := conn.Send(message); err != nil {
			println("error sending message")
			p.RemoveConnection(i)
		}
	}
}
