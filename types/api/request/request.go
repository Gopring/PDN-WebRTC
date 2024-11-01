// Package request contains api request type
package request

// Request is type of client's post request
type Request struct {
	UserID    string `json:"user_id"`
	SDP       string `json:"sdp"`
	ChannelID string `json:"channel_id"`
}
