package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type Traffic chan *Message

type TrafficTracker struct {
	messages   chan *Message
	messageCnt int
	origin     http.Handler
}

func (tt *TrafficTracker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tt.messageCnt += 1
	message := newMessage(tt.messageCnt, r)
	if nil != r.Body {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			fmt.Println("Error reading request stream", err.Error())
		}
		message.RequestBody = bodyBytes
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	handleMessageUpdate(tt.messages, message)

	writerRecorder := httptest.NewRecorder()
	tt.origin.ServeHTTP(writerRecorder, r)

	resp := writerRecorder.Result()
	message.StopTimer()

	bodyBytes := writerRecorder.Body.Bytes()
	message.Response = resp
	message.ResponseBody = bodyBytes
	handleMessageUpdate(tt.messages, message)

	for k, v := range writerRecorder.Header() {
		w.Header()[k] = v
	}
	w.Write(bodyBytes)
}

func (tt *TrafficTracker) GetTraffic() chan *Message {
	return tt.messages
}

func NewTrafficTracker(origin http.Handler) *TrafficTracker {
	return &TrafficTracker{messages: make(chan *Message), messageCnt: 0, origin: origin}
}

func handleMessageUpdate(messages chan *Message, message *Message) {
	if len(messages) == cap(messages) && len(messages) > 0 {
		<-messages
	}
	messages <- message
}
