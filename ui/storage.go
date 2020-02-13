package ui

import "proxio/client"

func NewStorage() *Storage {
	return &Storage{
		messages: make(map[int]*client.Message),
	}
}

type Storage struct {
	messages map[int]*client.Message
}

func (s *Storage) Add(m *client.Message) {
	s.messages[m.Id] = m
}

func (s *Storage) All() map[int]*client.Message {
	return s.messages
}

func (s *Storage) RemoveAll() {
	s.messages = make(map[int]*client.Message)
}
