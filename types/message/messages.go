// Package message provides data types for broker message.
package message

// Common is common data type for broker message
type Common struct {
	RequestID int `json:"request_id"`
}

// Push is data type for broker push
type Push struct {
	Common
	ChannelID string
	UserID    string
	SDP       string
}

// Pull is data type for broker pull
type Pull struct {
	Common
	ChannelID string
	UserID    string
	SDP       string
}

// Connected is data type for broker connected
type Connected struct {
	ChannelID string
	UserID    string
}

// Disconnected is data type for broker disconnected
type Disconnected struct {
	ChannelID string
	UserID    string
}
