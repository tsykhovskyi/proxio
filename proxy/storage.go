package proxy

func NewStorage() *Storage {
	return &Storage{}
}

type Storage struct {
	messages []*Message
}

func (s *Storage) Add(m *Message) {
	s.messages = append(s.messages, m)
}

func (s *Storage) All() []*Message {
	return s.messages
}
