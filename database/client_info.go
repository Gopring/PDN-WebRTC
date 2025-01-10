package database

import "time"

const (
	// Newbie is the class of a newly connected user.
	Newbie = iota

	// Publisher is the class of publisher. This means who publishes the stream.
	Publisher

	// Forwarder is the class of forwarder. This means who forwards the stream.
	Forwarder

	// Fetcher is the class of fetcher. This means who fetches the stream.
	Fetcher

	// Candidate is the class of a user who might be eligible to become a forwarder.
	Candidate
)

const (
	// SERVER indicates that the stream is being fetched from the server.
	SERVER = "server"

	// NONE indicates that there is no active fetch connection for the client.
	NONE = "none"
)

// ClientInfo is a struct for client information.
type ClientInfo struct {
	ID              string
	ChannelID       string
	Class           int
	ConnectionCount int
	FetchFrom       string
	CreatedAt       time.Time
}

// IncreaseConnectionCount increases the ConnectionCount field.
func (u *ClientInfo) IncreaseConnectionCount() {
	u.ConnectionCount = u.ConnectionCount + 1
}

// DecreaseConnectionCount decreases the ConnectionCount field.
func (u *ClientInfo) DecreaseConnectionCount() {
	u.ConnectionCount = u.ConnectionCount - 1
}

// UpdateClass updates the class field with the provided value.
func (u *ClientInfo) UpdateClass(class int) {
	u.Class = class
}

// CanForward returns whether the client can forward the stream.
func (u *ClientInfo) CanForward() bool {
	return u.Class == Newbie || u.Class == Forwarder || u.Class == Candidate
}

// UpdateFetchFrom updates the FetchFrom field with the provided value.
func (u *ClientInfo) UpdateFetchFrom(clientID string) {
	u.FetchFrom = clientID
}

// DeepCopy creates a deep copy of the given ClientInfo.
func (u *ClientInfo) DeepCopy() *ClientInfo {
	return &ClientInfo{
		ID:              u.ID,
		ChannelID:       u.ChannelID,
		CreatedAt:       u.CreatedAt,
		Class:           u.Class,
		FetchFrom:       u.FetchFrom,
		ConnectionCount: u.ConnectionCount,
	}
}
