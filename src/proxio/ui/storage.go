package ui

import "proxio/client"

func NewStorage() *Storage {
	return &Storage{
		domains: make(map[string][]*client.Message),
	}
}

type Storage struct {
	domains map[string][]*client.Message
}

func (s *Storage) Add(m *client.Message) {
	messages, ok := s.domains[m.Request.Host]
	if !ok {
		s.domains[m.Request.Host] = make([]*client.Message, 0)
		messages = s.domains[m.Request.Host]
	}

	if size := len(messages); size > 0 && messages[size-1].Id == m.Id {
		messages[size-1] = m
	} else {
		s.domains[m.Request.Host] = append(messages, m)
	}
}

func (s *Storage) All(domain string) []*client.Message {
	return s.domains[domain]
}

func (s *Storage) RemoveAll(domain string) {
	messages, ok := s.domains[domain]
	if ok {
		messages = messages[:0]
	}
}
