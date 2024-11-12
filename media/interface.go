// Package media contains managing channels and connections using WebRTC.
package media

// Media is an interface for managing channels and connections.
//
//go:generate mockgen -destination=mock_media.go -package=media . Media
type Media interface {
	AddSender(channelID string, userID string, sdp string) (string, error)
	AddReceiver(channelID string, userID string, sdp string) (string, error)
	AddForwarder(channelID string, userID string, sdp string) (string, error)
	GetForwarder(channelID string) (string, error)
}
