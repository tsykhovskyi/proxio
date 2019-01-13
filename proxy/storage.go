package proxy

func NewStorage() *Storage {
	return &Storage{
		messages: make(map[int]*Message),
	}
}

type Storage struct {
	messages map[int]*Message
}

func (s *Storage) Add(m *Message) {
	s.messages[m.Id] = m
}

func (s *Storage) All() map[int]*Message {
	return s.messages
}

func (s *Storage) RemoveAll() {
	s.messages = make(map[int]*Message)
}
