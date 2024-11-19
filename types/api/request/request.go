// Package request contains api request type
package request

// Activate is data type for activating user
type Activate struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

// Signal is data type for signaling
type Signal struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}
