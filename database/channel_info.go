package database

import "time"

// ChannelInfo is a struct for channel information.
type ChannelInfo struct {
	ID        string
	Key       string
	CreatedAt time.Time
}

// Authenticate checks if the given channel ID and key are authenticated.
func (c *ChannelInfo) Authenticate(key string) bool {
	return c.Key == key
}

// DeepCopy creates a deep copy of the given ChannelInfo.
func (c *ChannelInfo) DeepCopy() *ChannelInfo {
	return &ChannelInfo{
		ID:  c.ID,
		Key: c.Key,
	}
}
