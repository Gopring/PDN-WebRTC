// Package request contains api request type
package request

import "encoding/json"

type Common struct {
	RequestID int             `json:"request_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// Activate is data type for activating user
type Activate struct {
	RequestID int    `json:"request_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

// Push is data type for push stream
type Push struct {
	SDP string `json:"sdp"`
}

// Pull is data type for push stream
type Pull struct {
	SDP string `json:"sdp"`
}
