package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proxio/proxy"
	"regexp"
)

func NewController() *Controller {
	return &Controller{
		Storage:        NewStorage(),
		ConnectionPool: NewConnectionPool(),
	}
}

type Controller struct {
	Storage        *Storage
	ConnectionPool *Pool
}

func (c *Controller) listenMessages(messagesChan chan *proxy.Message) {
	go func() {
		for m := range messagesChan {
			c.Storage.Add(m)
			c.ConnectionPool.BroadcastMessage(m)
		}
	}()
}

func (c *Controller) static(w http.ResponseWriter, r *http.Request) {
	fileName := filterURI(r.RequestURI)

	if fileName == "/" {
		fileName = "/index.html"
	}

	http.ServeFile(w, r, "ui/web"+fileName)
}

func filterURI(uri string) string {
	re := regexp.MustCompile("[?].*$")
	uri = re.ReplaceAllString(uri, "")

	return uri
}

func (c *Controller) allMessages(w http.ResponseWriter, r *http.Request) {
	response := make([]*proxy.MessageContent, 0)

	for _, m := range c.Storage.All() {
		response = append(response, m.GetContext())
	}

	payload, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on message reading: %s", err), 500)
	}

	w.Write(payload)
}

func (c *Controller) check(w http.ResponseWriter, r *http.Request) {
	const RequestIdHeader = "Requests-Identifier"
	requestId := r.Header.Get(RequestIdHeader)
	connection, err := c.ConnectionPool.NewConnection(requestId)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Your connection was closed or not found"))
		return
	}

	messages := connection.PullBufferedMessages()
	if messages == nil {
		ctx := r.Context()
		select {
		case m := <-connection.Messages:
			messages = append(messages, m)
		case <-ctx.Done():
			c.ConnectionPool.CloseConnection(connection)
			return
		}
	}

	var response []*proxy.MessageContent
	for _, message := range messages {
		response = append(response, message.GetContext())
	}
	payload, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on message reading: %s", err), 500)
	}

	w.Header().Add(RequestIdHeader, connection.GetId())
	w.Write(payload)
}

func (c *Controller) clear(w http.ResponseWriter, r *http.Request) {
	c.Storage.RemoveAll()
}
