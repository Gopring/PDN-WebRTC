// Package request contains api request type
package request

import "encoding/json"

// Constants for request types
const (
	ACTIVATE     = "ACTIVATE"
	PUSH         = "PUSH"
	PULL         = "PULL"
	FORWARDING   = "FORWARDING"
	SIGNAL       = "SIGNAL"
	FORWARDED    = "FORWARDED"
	DISCONNECTED = "DISCONNECTED"
	FAILED       = "FAILED"
)

// Common is data type that must be implemented in all request
type Common struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Activate is data type for activating user
type Activate struct {
	ChannelID  string `json:"channel_id"`
	ChannelKey string `json:"channel_key"`
	ClientID   string `json:"client_id"`
}

// Push is data type for push stream
type Push struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Pull is data type for push stream
type Pull struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Forwarding is data type for push stream
type Forwarding struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Signal is data type for exchanging SDP
type Signal struct {
	ConnectionID string `json:"connection_id"`
	SignalType   string `json:"signal_type"`
	SignalData   string `json:"signal_data"`
}

// Forwarded is data type for success response
type Forwarded struct {
	ConnectionID string `json:"connection_id"`
}

// Failed is data type for fail response
type Failed struct {
	ConnectionID string `json:"connection_id"`
}

// Disconnected is data type for disconnecting user
type Disconnected struct {
	ConnectionID string `json:"connection_id"`
}
