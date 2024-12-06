package database

import (
	"time"
)

// Status is the status of the connection.
const (
	Initialized = iota
	Connected
)

// ConnectionInfo is a struct for WebRTC connection information.
type ConnectionInfo struct {
	ID          string
	ChannelID   string
	To          string
	From        string
	Status      int
	CreatedAt   time.Time
	ConnectedAt time.Time
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

// IsUpstream checks if the connection is an upstream connection.
func (c *ConnectionInfo) IsUpstream() bool {
	return c.To == MediaServerID
}

// IsPeerConnection checks if the connection is a peer connection.
func (c *ConnectionInfo) IsPeerConnection() bool {
	return c.To != MediaServerID && c.From != MediaServerID
}

// DeepCopy creates a deep copy of the given ConnectionInfo.
func (c *ConnectionInfo) DeepCopy() *ConnectionInfo {
	return &ConnectionInfo{
		ID:          c.ID,
		ChannelID:   c.ChannelID,
		To:          c.To,
		From:        c.From,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		ConnectedAt: c.ConnectedAt,
	}
}
