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

// Authorize checks if the given channel ID and client ID are authorized.
func (c *ConnectionInfo) Authorize(channelID, clientID string) bool {
	return c.ChannelID == channelID && (c.To == clientID || c.From == clientID)
}

// GetCounterpart returns the counterpart of the given client ID.
func (c *ConnectionInfo) GetCounterpart(clientID string) string {
	if c.To == clientID {
		return c.From
	}
	return c.To
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
