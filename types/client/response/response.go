// Package response provides data types for server response to client.
package response

// Activate is data type for activating user
type Activate struct {
	RequestID  int    `json:"request_id"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Push is data type for response to push
type Push struct {
	RequestID  int    `json:"request_id"`
	StatusCode int    `json:"status_code"`
	SDP        string `json:"sdp"`
}

// Pull is data type for response to pull
type Pull struct {
	RequestID  int    `json:"request_id"`
	StatusCode int    `json:"status_code"`
	SDP        string `json:"sdp"`
}
