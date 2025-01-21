// Package message provides data types for broker message.
package message

// Activate is data type for deactivating user
type Activate struct {
	ChannelID string
	ClientID  string
}

// Deactivate is data type for deactivating user
type Deactivate struct {
	ChannelID string
	ClientID  string
}

// Push is data type for broker push
type Push struct {
	ConnectionID string
	ChannelID    string
	ClientID     string
	SDP          string
}

// Pull is data type for broker pull
type Pull struct {
	ConnectionID string
	ChannelID    string
	ClientID     string
	SDP          string
}

// Upstream is data type for broker upstream
type Upstream struct {
	ConnectionID string
	Key          string
	SDP          string
}

// Downstream is data type for broker downstream
type Downstream struct {
	ConnectionID string
	StreamID     string
	Key          string
	SDP          string
}

// Connected is data type for broker connected
type Connected struct {
	ConnectionID string
}

// Disconnected is data type for broker disconnected
type Disconnected struct {
	ConnectionID string
}

// Failed is data type for forwarding failed
type Failed struct {
	ConnectionID string
}

// Clear is data type for clearing connection
type Clear struct {
	ConnectionID string
}
