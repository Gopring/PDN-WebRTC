// Package memory provides an in-memory database implementation.
package memory

import (
	"fmt"
	"github.com/hashicorp/go-memdb"
	"log"
	"pdn/database"
	"time"
)

// DB is a memory-backed database.
type DB struct {
	db *memdb.MemDB
}

// New creates a new memory-backed database.
func New(setDefaultChannel bool) *DB {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	newDB := &DB{
		db: db,
	}
	if setDefaultChannel {
		if err := newDB.EnsureDefaultChannelInfo(database.DefaultChannelID, database.DefaultChannelKey); err != nil {
			panic(err)
		}
		log.Printf("default channel created: ID:%s, Key:%s", database.DefaultChannelID, database.DefaultChannelKey)
	}
	return newDB
}

// EnsureDefaultChannelInfo creates a new channel if it doesn't exist. This is
// used for testing and debugging purposes.
func (d *DB) EnsureDefaultChannelInfo(channelID, channelKey string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	existing, err := txn.First(tblChannels, idxChannelID, channelID)
	if err != nil {
		return fmt.Errorf("find channel by channelID: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("%s: %w", channelID, database.ErrChannelAlreadyExists)
	}
	info := &database.ChannelInfo{
		ID:  channelID,
		Key: channelKey,
	}
	if err := txn.Insert(tblChannels, info); err != nil {
		return fmt.Errorf("insert channel: %w", err)
	}
	txn.Commit()
	return nil
}

// FindChannelInfoByID finds a channel by its ID.
func (d *DB) FindChannelInfoByID(id string) (*database.ChannelInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First(tblChannels, idxChannelID, id)
	if err != nil {
		return nil, fmt.Errorf("find project by public key: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: %w", id, database.ErrChannelNotFound)
	}

	return raw.(*database.ChannelInfo).DeepCopy(), nil
}

// CreateClientInfo creates a new user if it doesn't exist.
func (d *DB) CreateClientInfo(channelID, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	existing, err := txn.First(tblUsers, idxUserID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("%s: %w", clientID, database.ErrClientAlreadyExists)
	}

	info := &database.ClientInfo{
		ChannelID: channelID,
		ID:        clientID,
		Class:     database.Candidate,
		CreatedAt: time.Now(),
	}
	if err := txn.Insert(tblUsers, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
}

// FindClientInfoToForward  finds a client by their ID.
func (d *DB) FindClientInfoToForward(channelID string, to string) (*database.ClientInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get(tblUsers, idxUserID, channelID)
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		info := raw.(*database.ClientInfo)
		if info.Class != database.Fetcher && info.ID != to {
			return info, nil
		}
	}

	return nil, nil
}

// DeleteClientInfoByID deletes a user by their ID.
func (d *DB) DeleteClientInfoByID(channelID, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblUsers, idxUserID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("user %s in channel %s: %w", clientID, channelID, database.ErrClientNotFound)
	}
	if err := txn.Delete(tblUsers, raw); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// CreatePushConnectionInfo creates a new connection between two users.
func (d *DB) CreatePushConnectionInfo(channelID, clientID, connectionID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First(tblConnections, idxConnTo, channelID, database.MediaServerID)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw != nil {
		return nil, fmt.Errorf("%s: %w", clientID, database.ErrPushConnectionExists)
	}

	raw, err = txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw != nil {
		return nil, fmt.Errorf("%s: %w", connectionID, database.ErrConnectionAlreadyExists)
	}

	newConn := &database.ConnectionInfo{
		ID:                  connectionID,
		ChannelID:           channelID,
		From:                clientID,
		To:                  database.MediaServerID,
		IsConnectWithServer: true,
		CreatedAt:           time.Now(),
	}

	if err := txn.Insert(tblConnections, newConn); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return newConn.DeepCopy(), nil
}

// CreatePullConnectionInfo creates a new connection between two users.
func (d *DB) CreatePullConnectionInfo(channelID, clientID, connectionID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw != nil {
		return nil, fmt.Errorf("%s: %w", connectionID, database.ErrConnectionAlreadyExists)
	}

	newConn := &database.ConnectionInfo{
		ID:                  connectionID,
		ChannelID:           channelID,
		IsConnectWithServer: true,
		From:                database.MediaServerID,
		To:                  clientID,
		CreatedAt:           time.Now(),
	}

	if err := txn.Insert(tblConnections, newConn); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return newConn.DeepCopy(), nil
}

// CreatePeerConnectionInfo creates a new connection between two clients.
func (d *DB) CreatePeerConnectionInfo(channelID, from, to, connectionID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	if raw != nil {
		return nil, fmt.Errorf("%s: %w", from, database.ErrConnectionAlreadyExists)
	}
	newConn := &database.ConnectionInfo{
		ID:                  connectionID,
		ChannelID:           channelID,
		From:                from,
		To:                  to,
		IsConnectWithServer: false,
		IsConnected:         false,
		CreatedAt:           time.Now(),
	}
	if err := txn.Insert(tblConnections, newConn); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return newConn.DeepCopy(), nil
}

// FindUpstreamInfo finds an upstream connection by its channel ID.
func (d *DB) FindUpstreamInfo(channelID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnTo, channelID, database.MediaServerID)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: %w", channelID, database.ErrConnectionNotFound)
	}
	return raw.(*database.ConnectionInfo).DeepCopy(), nil
}

// UpdateConnectionInfo updates the connection status.
func (d *DB) UpdateConnectionInfo(connected bool, connectionID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("%s: %w", connectionID, database.ErrConnectionNotFound)
	}
	info := raw.(*database.ConnectionInfo).DeepCopy()
	info.IsConnected = connected
	info.ConnectedAt = time.Now()
	if err := txn.Insert(tblConnections, info); err != nil {
		return fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return nil
}
