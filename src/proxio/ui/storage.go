package ui

import "proxio/client"

func NewStorage() *Storage {
	return &Storage{
		messages: make(map[int]*client.Message),
		domains:  make(map[string][]*client.Message),
	}
}

type Storage struct {
	messages map[int]*client.Message
	domains  map[string][]*client.Message
}

func (s *Storage) Add(m *client.Message) {
	s.messages[m.Id] = m

	messages, ok := s.domains[m.Request.Host]
	if !ok {
		s.domains[m.Request.Host] = make([]*client.Message, 0)
		messages = s.domains[m.Request.Host]
	}

	s.domains[m.Request.Host] = append(messages, m)
}

func (s *Storage) All(domain string) []*client.Message {
	return s.domains[domain]
}

func (s *Storage) RemoveAll() {
	// s.messages = make(map[int]*client.Message)
}
