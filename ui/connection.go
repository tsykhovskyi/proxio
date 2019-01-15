package ui

import (
	"errors"
	"math/rand"
	"proxio/proxy"
	"strconv"
)

func NewConnectionPool() *Pool {
	return &Pool{
		random:      rand.New(rand.NewSource(99)),
		Connections: make(map[int]*Connection),
	}
}

type Pool struct {
	random      *rand.Rand
	Connections map[int]*Connection
}

func (p Pool) NewConnection(RequestId string) (*Connection, error) {
	id, err := strconv.Atoi(RequestId)
	if err == nil {
		if conn, exist := p.Connections[id]; exist {
			return conn, nil
		}
		return nil, errors.New("connection not found")
	}

	conn := &Connection{
		Id:       p.random.Int(),
		Messages: make(chan *proxy.Message, 10),
	}
	p.Connections[conn.Id] = conn
	return conn, nil
}

func (p Pool) CloseConnection(conn *Connection) {
	delete(p.Connections, conn.Id)
}

func (p Pool) BroadcastMessage(message *proxy.Message) {
	for _, conn := range p.Connections {
		err := conn.PushMessage(message)
		if err != nil {
			p.CloseConnection(conn)
		}
	}
}

type Connection struct {
	Id       int
	Messages chan *proxy.Message
}

func (c *Connection) GetId() string {
	return strconv.Itoa(c.Id)
}

func (c *Connection) PushMessage(m *proxy.Message) error {
	if len(c.Messages) == cap(c.Messages) {
		return errors.New("message channel is full")
	}
	c.Messages <- m
	return nil
}

func (c *Connection) PullBufferedMessages() []*proxy.Message {
	var buffer []*proxy.Message
	for {
		select {
		case message := <-c.Messages:
			buffer = append(buffer, message)
		default:
			return buffer
		}
	}
}
