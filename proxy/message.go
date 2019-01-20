package proxy

import (
	"net/http"
	"time"
)

func newMessage(req *http.Request) *Message {
	messageCounter++
	return &Message{
		Id:        messageCounter,
		Request:   req,
		Response:  nil,
		StartedAt: time.Now(),
	}
}

type Message struct {
	Id           int
	Request      *http.Request
	RequestBody  []byte
	Response     *http.Response
	ResponseBody []byte
	StartedAt    time.Time
	FinishedAt   time.Time
}

func (m *Message) HasResponse() bool {
	return m.Response != nil
}

func (m *Message) GetContext() *MessageContent {
	return BuildContent(m)
}

func (m *Message) StopTimer() {
	m.FinishedAt = time.Now()
}
