package database

import "time"

const (
	// Candidate is the class of candidate. This means who not classified class yet.
	Candidate = "candidate"

	// Forwarder is the class of forwarder. This means who forwards the stream.
	Forwarder = "forwarder"

	// Fetcher is the class of fetcher. This means who fetches the stream.
	Fetcher = "fetcher"
)

// ClientInfo is a struct for client information.
type ClientInfo struct {
	ID        string
	ChannelID string
	Class     string
	CreatedAt time.Time
}

// DeepCopy creates a deep copy of the given ClientInfo.
func (u *ClientInfo) DeepCopy() *ClientInfo {
	return &ClientInfo{
		ID:        u.ID,
		ChannelID: u.ChannelID,
		Class:     u.Class,
		CreatedAt: u.CreatedAt,
	}
}
