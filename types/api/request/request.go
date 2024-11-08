// Package request contains api request type
package request

// Signal is data type for signaling
type Signal struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	SDP       string `json:"sdp"`
	ChannelID string `json:"channel_id"`
}
