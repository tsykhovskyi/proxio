package traffic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Traffic chan *Message

type TrafficTracker struct {
	messages   chan *Message
	messageCnt int
	messageBuf map[*http.Request]*Message
}

func (tt *TrafficTracker) RequestStarted(r *http.Request) {
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
	tt.messageBuf[r] = message
}

func (tt *TrafficTracker) RequestFinished(r *http.Request, response *http.Response) {
	message := tt.messageBuf[r]
	if message == nil {
		panic("")
	}
	message.StopTimer()

	message.Response = response
	if nil != response.Body {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			fmt.Println("Error reading response stream", err.Error())
		}
		message.ResponseBody = bodyBytes
		response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	handleMessageUpdate(tt.messages, message)
}

func (tt *TrafficTracker) GetTraffic() chan *Message {
	return tt.messages
}

func NewTrafficTracker() *TrafficTracker {
	return &TrafficTracker{
		messages:   make(chan *Message),
		messageCnt: 0,
		messageBuf: make(map[*http.Request]*Message),
	}
}

func handleMessageUpdate(messages chan *Message, message *Message) {
	if len(messages) == cap(messages) && len(messages) > 0 {
		<-messages
	}
	messages <- message
}
