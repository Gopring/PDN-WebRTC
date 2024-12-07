package database

import "time"

const (
	// Candidate is the class of candidate. This means who not classified class yet.
	Candidate = iota

	// Publisher is the class of publisher. This means who publishes the stream.
	Publisher

	// Forwarder is the class of forwarder. This means who forwards the stream.
	Forwarder

	// Fetcher is the class of fetcher. This means who fetches the stream.
	Fetcher
)

// ClientInfo is a struct for client information.
type ClientInfo struct {
	ID        string
	ChannelID string
	Class     int
	CreatedAt time.Time
}

// CanForward returns whether the client can forward the stream.
func (u *ClientInfo) CanForward() bool {
	return u.Class == Candidate || u.Class == Forwarder
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
