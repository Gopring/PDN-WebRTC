// Package response provides data types for server response to client.
package response

// Constants for response types
const (
	ACTIVATE    = "ACTIVATE"
	FORWARDING  = "FORWARDING"
	FORWARD     = "FORWARD"
	CLOSED      = "CLOSED"
	CLEAR       = "CLEAR"
	SIGNAL      = "SIGNAL"
	CLASSIFY    = "CLASSIFY"
	CLASSIFYING = "CLASSIFYING"
)

// Activate is data type for activating user
type Activate struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Forwarding is data type for server sent response to command user forwarding
type Forwarding struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Forward is data type for server sent response to command user fetching
type Forward struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
}

// Closed is data type for server sent response to command user closing
type Closed struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
}

// Clear is data type for server sent response to command user clearing
type Clear struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
}

// Signal is data type for exchanging SDP
type Signal struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SignalType   string `json:"signal_type"`
	SignalData   string `json:"signal_data"`
}

// Classifying is data type for server sent response to command user forwarding while classifying
type Classifying struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
	SDP          string `json:"sdp"`
}

// Classify is data type for server sent response to command user fetching while classifying
type Classify struct {
	Type         string `json:"type"`
	ConnectionID string `json:"connection_id"`
}
