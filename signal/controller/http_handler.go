// Package controller handles HTTP logic.
package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pdn/media"
	"pdn/types/api/request"
)

const (
	broadcastPath = "/broadcast"
	viewPath      = "/view"
	maxBodySize   = 1 << 20 // 1 MB
)

// Controller handles HTTP requests.
type Controller struct {
	media *media.Media
	debug bool
}

// New creates a new instance of Controller.
func New(m *media.Media, isDebug bool) *Controller {
	return &Controller{
		media: m,
		debug: isDebug,
	}
}

// ServeHTTP handles HTTP requests.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case broadcastPath:
		c.handleRequest(w, r, c.media.AddSender)
	case viewPath:
		c.handleRequest(w, r, c.media.AddReceiver)
	default:
		c.Error(w, fmt.Errorf("wrong path"), http.StatusNotFound)
	}
}

// handleRequest handles HTTP requests.
func (c *Controller) handleRequest(w http.ResponseWriter, r *http.Request, media media.Func) {
	req, err := parse(w, r)
	if err != nil {
		c.Error(w, err, http.StatusBadRequest)
		return
	}

	sdp, err := media(req.ChannelID, req.UserID, req.SDP)
	if err != nil {
		c.Error(w, fmt.Errorf("failed in media: %w", err), http.StatusBadRequest)
		return
	}

	if err = responseWrite(w, string(sdp)); err != nil {
		c.Error(w, err, http.StatusInternalServerError)
	}
}

func parse(w http.ResponseWriter, r *http.Request) (request.Request, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer func(r io.ReadCloser) {
		if err := r.Close(); err != nil {
			//TODO(window9u): handle this err more clearly
			log.Println(err)
		}
	}(r.Body)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return request.Request{}, fmt.Errorf("failed to read body")
	}

	req := request.Request{}
	if err = json.Unmarshal(data, &req); err != nil {
		return request.Request{}, fmt.Errorf("failed to parse body")
	}

	return req, nil
}

func (c *Controller) Error(w http.ResponseWriter, err error, statusCode int) {
	if !c.debug {
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}
	log.Println(err)
	http.Error(w, err.Error(), statusCode)
}

func responseWrite(w http.ResponseWriter, msg string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprint(w, msg)
	return err
}
