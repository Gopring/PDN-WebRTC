package database

import "time"

// ClientInfo is a struct for client information.
type ClientInfo struct {
	ID        string
	ChannelID string
	CreatedAt time.Time
}

// DeepCopy creates a deep copy of the given ClientInfo.
func (u *ClientInfo) DeepCopy() *ClientInfo {
	return &ClientInfo{
		ID:        u.ID,
		ChannelID: u.ChannelID,
		CreatedAt: u.CreatedAt,
	}
}
