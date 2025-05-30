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
func New(config database.Config) *DB {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	newDB := &DB{
		db: db,
	}
	if config.SetDefaultChannel {
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

// FindOrCreateChannelInfoByID finds a channel by its ID.
func (d *DB) FindOrCreateChannelInfoByID(id string) (*database.ChannelInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblChannels, idxChannelID, id)
	if err != nil {
		return nil, fmt.Errorf("find project by public key: %w", err)
	}
	if raw == nil {
		// Channel not found, create a new one
		info := &database.ChannelInfo{
			ID:  id,
			Key: id,
		}
		if err := txn.Insert(tblChannels, info); err != nil {
			return nil, fmt.Errorf("insert channel: %w", err)
		}
		txn.Commit()
		return info.DeepCopy(), nil
	}

	return raw.(*database.ChannelInfo).DeepCopy(), nil
}

// FindAllChannelInfos retrieves all channel information from the database.
func (d *DB) FindAllChannelInfos() ([]*database.ChannelInfo, error) {
	txn := d.db.Txn(false) // Read-only transaction
	defer txn.Abort()

	it, err := txn.Get(tblChannels, idxChannelID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving channels: %w", err)
	}

	var channelInfos []*database.ChannelInfo

	for obj := it.Next(); obj != nil; obj = it.Next() {
		channel, ok := obj.(*database.ChannelInfo)
		if !ok {
			return nil, fmt.Errorf("invalid data type in channels table")
		}

		channelInfos = append(channelInfos, channel.DeepCopy())
	}

	return channelInfos, nil
}

// DeleteChannelInfoByID deletes a channel by its ID.
func (d *DB) DeleteChannelInfoByID(id string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblChannels, idxChannelID, id)
	if err != nil {
		return fmt.Errorf("find channel by channelID: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("%s: %w", id, database.ErrChannelNotFound)
	}
	if err := txn.Delete(tblChannels, raw); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	txn.Commit()
	return nil
}

// CreateClientInfo creates a new user if it doesn't exist.
func (d *DB) CreateClientInfo(channelID, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	existing, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("%s: %w", clientID, database.ErrClientAlreadyExists)
	}

	info := &database.ClientInfo{
		ChannelID: channelID,
		ID:        clientID,
		CreatedAt: time.Now(),
	}
	if err := txn.Insert(tblClients, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
}

// FindClientInfoByID finds a user by their ID.
func (d *DB) FindClientInfoByID(channelID, clientID string) (*database.ClientInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: %w", clientID, database.ErrClientNotFound)
	}
	return raw.(*database.ClientInfo).DeepCopy(), nil
}

// DeleteClientInfoByID deletes a user by their ID.
func (d *DB) DeleteClientInfoByID(channelID, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("user %s in channel %s: %w", clientID, channelID, database.ErrClientNotFound)
	}
	if err := txn.Delete(tblClients, raw); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	txn.Commit()
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
		ID:        connectionID,
		ChannelID: channelID,
		From:      clientID,
		To:        database.MediaServerID,
		Status:    database.Initialized,
		Type:      database.PushToServer,
		CreatedAt: time.Now(),
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
		ID:        connectionID,
		ChannelID: channelID,
		From:      database.MediaServerID,
		To:        clientID,
		Status:    database.Initialized,
		Type:      database.PullFromServer,
		CreatedAt: time.Now(),
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
		ID:        connectionID,
		ChannelID: channelID,
		From:      from,
		To:        to,
		Status:    database.Initialized,
		Type:      database.PeerToPeer,
		CreatedAt: time.Now(),
	}
	if err := txn.Insert(tblConnections, newConn); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return newConn.DeepCopy(), nil
}

// FindUpstreamInfo finds an upstream connection by its channel ID.
func (d *DB) FindUpstreamInfo(channelID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(false)
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

// FindDownstreamInfo finds a downstream connection by its channel ID and client ID.
func (d *DB) FindDownstreamInfo(channelID, to string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get(tblConnections, idxConnTo, channelID, to)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		info := raw.(*database.ConnectionInfo)
		if info.IsDownstream() {
			return info.DeepCopy(), nil
		}
	}
	return nil, database.ErrConnectionNotFound
}

// FindAllPeerConnectionInfoByFrom finds a connection by its from field.
func (d *DB) FindAllPeerConnectionInfoByFrom(channelID, from string) ([]*database.ConnectionInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get(tblConnections, idxConnFrom, channelID, from)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	var connections []*database.ConnectionInfo
	for {
		raw := iter.Next()
		if raw == nil {
			break
		}

		info := raw.(*database.ConnectionInfo)
		if info.Type == database.PeerToPeer {
			connections = append(connections, info.DeepCopy())
		}
	}
	return connections, nil
}

// FindAllPeerConnectionInfoByTo finds a connection by its to field.
func (d *DB) FindAllPeerConnectionInfoByTo(channelID, to string) ([]*database.ConnectionInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get(tblConnections, idxConnTo, channelID, to)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	var connections []*database.ConnectionInfo
	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		info := raw.(*database.ConnectionInfo)

		if info.Type == database.PeerToPeer {
			connections = append(connections, info.DeepCopy())
		}
	}
	return connections, nil
}

// FindConnectionInfoByID finds a connection by its ID.
func (d *DB) FindConnectionInfoByID(ConnectionID string) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: %w", ConnectionID, database.ErrConnectionNotFound)
	}
	return raw.(*database.ConnectionInfo).DeepCopy(), nil
}

// UpdateConnectionInfo updates the connection status.
func (d *DB) UpdateConnectionInfo(connectionID string, status int) (*database.ConnectionInfo, error) {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: %w", connectionID, database.ErrConnectionNotFound)
	}
	info := raw.(*database.ConnectionInfo).DeepCopy()
	info.Status = status
	info.ConnectedAt = time.Now()
	if err := txn.Insert(tblConnections, info); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return info, nil
}

// DeleteConnectionInfoByID deletes a connection by its ID.
func (d *DB) DeleteConnectionInfoByID(connectionID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblConnections, idxConnID, connectionID)
	if err != nil {
		return fmt.Errorf("find connection by connectionID: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("%s: %w", connectionID, database.ErrConnectionNotFound)
	}
	if err := txn.Delete(tblConnections, raw); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	txn.Commit()
	return nil
}
