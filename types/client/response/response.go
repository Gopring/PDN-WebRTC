// Package response provides data types for server response to client.
package response

// Constants for response types
const (
	ACTIVATE = "ACTIVATE"
	PUSH     = "PUSH"
	PULL     = "PULL"
)

// Activate is data type for activating user
type Activate struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Push is data type for response to push
type Push struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Pull is data type for response to pull
type Pull struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Forward is data type for server sent response to command user forwarding
type Forward struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	To           string `json:"user_id"`
}

// Fetch is data type for server sent response to command user fetching
type Fetch struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	From         string `json:"user_id"`
}
