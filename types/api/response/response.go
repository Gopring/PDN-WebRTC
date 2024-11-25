package response

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
