// Package database provides an interface for database operations.
package database

import (
	"errors"
)

const (
	DefaultChannelID  = "channel-id"
	DefaultChannelKey = "channel-key"
	ServerID          = "media-server-id"
)

var (
	ErrClientAlreadyExists     = errors.New("client already exists")
	ErrChannelAlreadyExists    = errors.New("channel already exists")
	ErrConnectionAlreadyExists = errors.New("connection already exists")
	ErrPushConnectionExists    = errors.New("push connection already exists")
	ErrChannelNotFound         = errors.New("channel not found")
	ErrClientNotFound          = errors.New("client not found")
	ErrConnectionNotFound      = errors.New("connection not found")
)

// Database is an interface for database operations.
type Database interface {
	EnsureDefaultChannelInfo(channelID, channelKey string) error
	FindChannelInfoByID(id string) (*ChannelInfo, error)

	CreateClientInfo(channelID, clientID string) error
	FindClientInfoToForward(channelID string, to string) (*ClientInfo, error)
	DeleteClientInfoByID(channelID, clientID string) error

	CreateServerConnectionInfo(isPush bool, channelID, clientID, connectionID string) (*ConnectionInfo, error)
	CreateClientConnectionInfo(channelID, from, to, connectionID string) (*ConnectionInfo, error)
	FindUpstreamInfo(channelID string) (*ConnectionInfo, error)
	UpdateConnectionInfo(connected bool, connectionID string) error
}
