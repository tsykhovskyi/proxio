package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"proxio/client"
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

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	http.ServeFile(w, r, wd+"/ui/web"+fileName)
}

func filterURI(uri string) string {
	re := regexp.MustCompile("[?].*$")
	uri = re.ReplaceAllString(uri, "")

	return uri
}

func (c *Controller) allMessages(w http.ResponseWriter, r *http.Request) {
	messages := c.Storage.All("itsykhovskyi.proxio.rd")

	response := make([]*client.MessageContent, len(messages))

	for i, m := range messages {
		response[i] = m.GetContext()
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

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = filepath.Join(h.staticPath, path)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func NewSpaHandler() http.Handler {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return spaHandler{staticPath: wd + "/ui/web", indexPath: "index.html"}
}
