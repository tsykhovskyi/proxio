package proxy

import (
	"encoding/base64"
	"net/http"
	"time"
)

type MessageContent struct {
	Id       int
	Status   int
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
	Body    string
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
	c.Status = m.Status
	c.Time = &Time{
		StartedAt: m.StartedAt.Format(time.RFC3339),
	}
	c.Request = &Request{
		req.Method,
		req.RequestURI,
		getBodyForHeaders(m.RequestBody, req.Header),
		req.Header,
	}
	if res != nil {
		c.Time.FinishedAt = m.FinishedAt.Format(time.RFC3339)
		c.Time.TimeTaken = m.FinishedAt.Sub(m.StartedAt).Seconds()

		c.Response = &Response{
			res.StatusCode,
			getBodyForHeaders(m.ResponseBody, res.Header),
			res.Header,
		}
	}

	return c
}

func getBodyForHeaders(body []byte, h http.Header) string {
	if h.Get("Content-Type") == "image/jpeg" || h.Get("Content-Type") == "image/png" {
		return base64.StdEncoding.EncodeToString(body)
	}
	return string(body)
}
