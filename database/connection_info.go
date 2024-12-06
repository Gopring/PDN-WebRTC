package database

import "time"

// ConnectionInfo is a struct for WebRTC connection information.
type ConnectionInfo struct {
	ID                  string
	ChannelID           string
	To                  string
	From                string
	IsConnectWithServer bool
	IsConnected         bool
	CreatedAt           time.Time
	ConnectedAt         time.Time
}

// DeepCopy creates a deep copy of the given ConnectionInfo.
func (c *ConnectionInfo) DeepCopy() *ConnectionInfo {
	return &ConnectionInfo{
		ID:          c.ID,
		ChannelID:   c.ChannelID,
		To:          c.To,
		From:        c.From,
		CreatedAt:   c.CreatedAt,
		ConnectedAt: c.ConnectedAt,
	}
}
