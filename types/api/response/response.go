package response

// Activate is data type for activating user
type Activate struct {
	RequestID  int    `json:"request_id"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Signal is data type for signaling
type Signal struct {
	RequestID  int    `json:"request_id"`
	StatusCode int    `json:"status_code"`
	SDP        string `json:"sdp"`
}

type Arrange struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
	SDP        string `json:"sdp"`
}
