package ui

import "proxio/proxy"

func NewStorage() *Storage {
	return &Storage{
		messages: make(map[int]*proxy.Message),
	}
}

type Storage struct {
	messages map[int]*proxy.Message
}

func (s *Storage) Add(m *proxy.Message) {
	s.messages[m.Id] = m
}

func (s *Storage) All() map[int]*proxy.Message {
	return s.messages
}

func (s *Storage) RemoveAll() {
	s.messages = make(map[int]*proxy.Message)
}
