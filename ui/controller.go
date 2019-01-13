package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"proxio/proxy"
	"regexp"
	"time"
)

type Controller struct {
	MessagesChan chan *proxy.Message
	Storage      *proxy.Storage
}

func (c *Controller) static(w http.ResponseWriter, r *http.Request) {
	publicDir := "ui/public"

	fileName := filterURI(r.RequestURI)

	if fileName == "/" {
		fileName = "/index.html"
	}

	io, err := ioutil.ReadFile(publicDir + fileName)
	if err != nil {
		http.Error(w, "Unable to serve "+fileName, 500)
	}

	w.Write(io)
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
	response := make([]*proxy.MessageContent, 0)
	done := make(chan bool)

	go func() {
		for {
			select {
			case m := <-c.MessagesChan:
				response = append(response, m.GetContext())
			case <-time.After(time.Millisecond):
				if len(response) > 0 {
					done <- true
					return
				}
			}
		}
	}()
	<-done

	payload, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on message reading: %s", err), 500)
	}

	w.Write(payload)
}

func (c *Controller) clear(w http.ResponseWriter, r *http.Request) {
	c.Storage.RemoveAll()
}
