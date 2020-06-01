package repository

type Session struct {
	Id   string
	User struct {
		PubKey string
	}
	Tunnels []struct {
		Domain string
	}
}

type Sessions interface {
	Find(id string) (Session, bool)
}

type SessionsRepo struct {
	sessions []Session
}

func (s *SessionsRepo) Find(id string) (Session, bool) {
	for _, session := range s.sessions {
		if session.Id == id {
			return session, true
		}
	}
	return Session{}, false
}

func (s *SessionsRepo) Populate() {
	s.sessions = append(s.sessions, Session{Id: "777", User: struct{ PubKey string }{PubKey: "ssh-pub"}, Tunnels: []struct{ Domain string }{}})
}

func NewSessionsRepo() *SessionsRepo {
	return &SessionsRepo{sessions: make([]Session, 0)}
}
