package traffic

func NewStorage() *Storage {
	return &Storage{
		domains: make(map[string][]*Message),
	}
}

type Storage struct {
	domains map[string][]*Message
}

func (s *Storage) Add(m *Message) {
	messages, ok := s.domains[m.Request.Host]
	if !ok {
		s.domains[m.Request.Host] = make([]*Message, 0)
		messages = s.domains[m.Request.Host]
	}

	if size := len(messages); size > 0 && messages[size-1].Id == m.Id {
		messages[size-1] = m
	} else {
		s.domains[m.Request.Host] = append(messages, m)
	}
}

func (s *Storage) All(domain string) []*Message {
	return s.domains[domain]
}

func (s *Storage) RemoveAll(domain string) {
	messages, ok := s.domains[domain]
	if ok {
		messages = messages[:0]
	}
}
