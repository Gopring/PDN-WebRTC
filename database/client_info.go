package database

import "time"

const (
	Candidate = "candidate"
	Forwarder = "forwarder"
	Fetcher   = "fetcher"
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
