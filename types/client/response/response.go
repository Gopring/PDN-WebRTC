// Package response provides data types for server response to client.
package response

// Constants for response types
const (
	ACTIVATE = "ACTIVATE"
	FORWARD  = "FORWARD"
	FETCH    = "FETCH"
	EXCHANGE = "EXCHANGE"
)

// Activate is data type for activating user
type Activate struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Forward is data type for server sent response to command user forwarding
type Forward struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Fetch is data type for server sent response to command user fetching
type Fetch struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Exchange is data type for exchanging SDP
type Exchange struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	DataType     string `json:"data_type"`
	Data         string `json:"data"`
}
