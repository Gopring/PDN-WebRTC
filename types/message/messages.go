// Package message provides data types for broker message.
package message

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

// Succeed is data type for forwarding succeed
type Succeed struct {
	ConnectionID string
}

// Failed is data type for forwarding failed
type Failed struct {
	ConnectionID string
}

// Closure is data type for closing connection
type Closure struct {
	ConnectionID string
}
