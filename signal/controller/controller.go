// Package controller handles HTTP logic.
package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pdn/media"
	"pdn/types/api/request"
)

const (
	send      = "send"
	receive   = "receive"
	forward   = "forward"
	fetch     = "fetch"
	reconnect = "reconnect"
)

// Controller handles HTTP requests.
type Controller struct {
	media    *media.Media
	debug    bool
	upgrader websocket.Upgrader
}

// New creates a new instance of Handler.
func New(m *media.Media, isDebug bool) *Controller {
	return &Controller{
		media:    m,
		debug:    isDebug,
		upgrader: websocket.Upgrader{},
	}
}

// ServeHTTP handles HTTP requests.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Parse incoming message
		req, err := parseRequest(conn)
		if err != nil {
			log.Println("Parse error:", err)
			return
		}

		// Route message based on request type
		sdp, err := routeRequest(req, c.media)
		if err != nil {
			log.Println("Route error:", err)
			return
		}

		// Send response back to client
		if err := conn.WriteMessage(websocket.TextMessage, []byte(sdp)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
	//TODO(window9u): Clear Media resource correctly when signaling fails
}

// parseRequest reads a message from the WebSocket connection and unmarshal it into a Signal request
func parseRequest(conn *websocket.Conn) (request.Signal, error) {
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return request.Signal{}, fmt.Errorf("read message error: %w", err)
	}
	if messageType != websocket.TextMessage {
		return request.Signal{}, fmt.Errorf("unexpected message type: %d", messageType)
	}

	var req request.Signal
	if err := json.Unmarshal(message, &req); err != nil {
		return request.Signal{}, fmt.Errorf("unmarshal error: %w", err)
	}
	return req, nil
}

// routeRequest routes a parsed request based on its type
func routeRequest(req request.Signal, media *media.Media) (string, error) {
	switch req.Type {
	case send:
		return media.AddSender(req.ChannelID, req.UserID, req.SDP)
	case receive:
		return media.AddReceiver(req.ChannelID, req.UserID, req.SDP)
	case forward:
		//NOTE(window9u): forward is forwarder request that forwarding stream.
		//TODO(window9u): Managing forwarder in signaling server and make media server
		// control just for connection, not for channel.
		return media.AddReceiver(req.ChannelID, req.UserID, req.SDP)
	case fetch:
		//TODO(window9u): fetch is the request from fetch from forwarder.
		// 1. get sdp from forwarder
		// 2. send sdp to fetcher
		return "", nil
	case reconnect:
		// Handle reconnect case
	default:
		return "", fmt.Errorf("unknown request type: %s", req.Type)
	}
	return "", fmt.Errorf("unhandled request type: %s", req.Type)
}
