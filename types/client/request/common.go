// Package request defines structures for client request messages.
package request

import "encoding/json"

// Common represents a generic request structure used in WebSocket communication.
type Common struct {
	RequestID int             `json:"request_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}
