package proxy

import "time"

type MessageContent struct {
	Id       int
	Time     *Time
	Request  *Request
	Response *Response
}

type Time struct {
	StartedAt  string
	FinishedAt string
	TimeTaken  float64
}

type Request struct {
	Method  string
	URI     string
	Headers map[string][]string
}
type Response struct {
	Code    int
	Body    string
	Headers map[string][]string
}

func BuildContent(m *Message) *MessageContent {
	req := m.Request
	res := m.Response

	c := &MessageContent{}
	c.Id = m.Id
	c.Time = &Time{
		StartedAt: m.StartedAt.Format(time.RFC3339),
	}
	c.Request = &Request{
		req.Method,
		req.RequestURI,
		req.Header,
	}
	if res != nil {
		c.Time.FinishedAt = m.FinishedAt.Format(time.RFC3339)
		c.Time.TimeTaken = m.FinishedAt.Sub(m.StartedAt).Seconds()

		c.Response = &Response{
			res.StatusCode,
			string(m.ResponseBody),
			res.Header,
		}
	}

	return c
}
