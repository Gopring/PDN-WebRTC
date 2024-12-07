// Package database provides an interface for database operations.
package database

import (
	"errors"
)

const (
	// DefaultChannelID is the default channel ID. it is registered if flag is set.
	DefaultChannelID = "channel-id"

	// DefaultChannelKey is the default channel key. it is registered if flag is set.
	DefaultChannelKey = "channel-key"

	// MediaServerID is the default media server ID. It is used for From or To in ConnectionInfo.
	MediaServerID = "media-server-id"
)

var (
	// ErrClientAlreadyExists is returned when the client already exists.
	ErrClientAlreadyExists = errors.New("client already exists")

	// ErrChannelAlreadyExists is returned when the channel already exists.
	ErrChannelAlreadyExists = errors.New("channel already exists")

	// ErrConnectionAlreadyExists is returned when the connection already exists.
	ErrConnectionAlreadyExists = errors.New("connection already exists")

	// ErrPushConnectionExists is returned when the push connection already exists.
	ErrPushConnectionExists = errors.New("push connection already exists")

	// ErrChannelNotFound is returned when the channel is not found.
	ErrChannelNotFound = errors.New("channel not found")

	// ErrClientNotFound is returned when the client is not found.
	ErrClientNotFound = errors.New("client not found")

	// ErrConnectionNotFound is returned when the connection is not found.
	ErrConnectionNotFound = errors.New("connection not found")
)

// Database is an interface for database operations.
type Database interface {
	EnsureDefaultChannelInfo(channelID, channelKey string) error
	FindChannelInfoByID(id string) (*ChannelInfo, error)

	CreateClientInfo(channelID, clientID string) error
	FindClientInfoByID(channelID, clientID string) (*ClientInfo, error)
	FindForwarderInfo(channelID string, fetcherID string, max int) (*ClientInfo, error)
	UpdateClientInfo(channelID, clientID string, class, fetchFrom int) (*ClientInfo, error)
	DeleteClientInfoByID(channelID, clientID string) error

	CreatePushConnectionInfo(channelID, clientID, connectionID string) (*ConnectionInfo, error)
	CreatePullConnectionInfo(channelID, clientID, connectionID string) (*ConnectionInfo, error)
	CreatePeerConnectionInfo(channelID, from, to, connectionID string) (*ConnectionInfo, error)
	FindUpstreamInfo(channelID string) (*ConnectionInfo, error)
	FindDownstreamInfo(channelID, to string) (*ConnectionInfo, error)
	FindConnectionInfoByFrom(channelID, from string) ([]*ConnectionInfo, error)
	FindConnectionInfoByTo(channelID, from string) ([]*ConnectionInfo, error)
	FindConnectionInfoByID(ConnectionID string) (*ConnectionInfo, error)
	UpdateConnectionInfo(connectionID string, status int) (*ConnectionInfo, error)
	DeleteConnectionInfoByID(connectionID string) error
}
