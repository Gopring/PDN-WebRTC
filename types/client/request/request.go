// Package request contains api request type
package request

import "encoding/json"

// Constants for request types
const (
	ACTIVATE        = "ACTIVATE"
	PUSH            = "PUSH"
	PULL            = "PULL"
	FORWARD         = "FORWARD"
	SIGNAL          = "SIGNAL"
	CONNECTED       = "CONNECTED"
	DISCONNECTED    = "DISCONNECTED"
	FAILED          = "FAILED"
	CLASSIFYFORWARD = "CLASSIFYFORWARD"
	CLASSIFYSIGNAL  = "CLASSIFYSIGNAL"
	CLASSIFYRESULT  = "CLASSIFYRESULT"
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

// Forward is data type for push stream
type Forward struct {
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Signal is data type for exchanging SDP
type Signal struct {
	ConnectionID string `json:"connection_id"`
	SignalType   string `json:"signal_type"`
	SignalData   string `json:"signal_data"`
}

// Connected is data type for success response
type Connected struct {
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

// ClassifySignal is data type for exchanging SDP while classifying
type ClassifySignal struct {
	ConnectionID string `json:"connection_id"`
	PeerID       string `json:"peer_id"`
	SignalType   string `json:"signal_type"`
	SignalData   string `json:"signal_data"`
}

// ClassifyForward is data type for forwarding while classifying
type ClassifyForward struct {
	ConnectionID string `json:"connection_id"`
	PeerID       string `json:"peer_id"`
	SDP          string `json:"sdp"`
}

// ClassifyResult is data type for classifying result
type ClassifyResult struct {
	ConnectionID string `json:"connection_id"`
	PeerID       string `json:"peer_id"`
	Success      bool   `json:"success"`
	ChannelID    string `json:"channel_id"`
}
