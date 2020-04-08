package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var (
	messageCounter int = 0
)

func TrafficMiddleware(messages chan *Message, origin http.Handler) http.Handler {
	messageCounter := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		messageCounter++
		message := newMessage(messageCounter, r)
		if nil != r.Body {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				fmt.Println("Error reading request stream", err.Error())
			}
			message.RequestBody = bodyBytes
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		handleMessageUpdate(messages, message)

		writerRecorder := httptest.NewRecorder()
		origin.ServeHTTP(writerRecorder, r)

		resp := writerRecorder.Result()
		message.StopTimer()

		bodyBytes := writerRecorder.Body.Bytes()
		message.Response = resp
		message.ResponseBody = bodyBytes
		handleMessageUpdate(messages, message)

		for k, v := range writerRecorder.Header() {
			w.Header()[k] = v
		}
		w.Write(bodyBytes)
	})
}

func handleMessageUpdate(messages chan *Message, message *Message) {
	if len(messages) == cap(messages) && len(messages) > 0 {
		<-messages
	}
	messages <- message
}
