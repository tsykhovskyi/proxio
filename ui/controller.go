package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proxio/proxy"
	"regexp"
)

func NewController(storage *Storage) *Controller {
	return &Controller{
		Storage: storage,
	}
}

type Controller struct {
	Storage *Storage
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

func (c *Controller) clear(w http.ResponseWriter, r *http.Request) {
	c.Storage.RemoveAll()
}
