package ui

import (
	"proxio/client"
)

func NewConnectionPool() *Pool {
	pool := &Pool{
		DomainListeners: make(map[string][]*Connection, 0),
		closeChan:       make(chan *Connection),
	}

	go func() {
		for conn := range pool.closeChan {
			pool.removeConnection(conn)
		}
	}()

	return pool
}

type Pool struct {
	DomainListeners map[string][]*Connection
	closeChan       chan *Connection
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

func (p *Pool) BroadcastMessage(message *client.MessageContent) {
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
