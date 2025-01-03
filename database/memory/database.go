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
	existing, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("%s: %w", clientID, database.ErrClientAlreadyExists)
	}

	info := &database.ClientInfo{
		ChannelID:       channelID,
		ID:              clientID,
		Class:           database.Candidate,
		ConnectionCount: 0,
		CreatedAt:       time.Now(),
		LastUpdated:     time.Now(),
	}
	if err := txn.Insert(tblClients, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
}

// findClientInfoByID finds a user by their ID.
func (d *DB) findClientInfoByID(channelID, clientID string) (*database.ClientInfo, error) {
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

// UpdateClientInfoClass updates the user class.
func (d *DB) UpdateClientInfoClass(channelID string, clientID string, class int) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("user %s in channel %s: %w", clientID, channelID, database.ErrClientNotFound)
	}
	info := raw.(*database.ClientInfo).DeepCopy()
	info.UpdateClass(class)
	info.UpdateLastUpdated()
	if err := txn.Insert(tblClients, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
}

// IncreaseClientInfoConnCount updates the connection count + 1.
func (d *DB) IncreaseClientInfoConnCount(channelID string, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("user %s in channel %s: %w", clientID, channelID, database.ErrClientNotFound)
	}
	info := raw.(*database.ClientInfo).DeepCopy()
	info.IncreaseConnectionCount()
	info.UpdateLastUpdated()
	if err := txn.Insert(tblClients, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
}

// DecreaseClientInfoConnCount updates the connection count - 1.
func (d *DB) DecreaseClientInfoConnCount(channelID string, clientID string) error {
	txn := d.db.Txn(true)
	defer txn.Abort()
	raw, err := txn.First(tblClients, idxClientID, channelID, clientID)
	if err != nil {
		return fmt.Errorf("find user by username: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("user %s in channel %s: %w", clientID, channelID, database.ErrClientNotFound)
	}
	info := raw.(*database.ClientInfo).DeepCopy()
	info.DecreaseConnectionCount()
	info.UpdateLastUpdated()
	if err := txn.Insert(tblClients, info); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	txn.Commit()
	return nil
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

// FindClientInfoByClass finds a user by their Class.
func (d *DB) FindClientInfoByClass(channelID string, class int) ([]*database.ClientInfo, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get(tblClients, idxClientChannelID, channelID)
	if err != nil {
		return nil, fmt.Errorf("error fetching forwarders by channel ID: %w", err)
	}
	var results []*database.ClientInfo
	for obj := it.Next(); obj != nil; obj = it.Next() {
		raw := obj.(*database.ClientInfo)

		if raw.Class == class {
			clientInfo, err := d.findClientInfoByID(raw.ChannelID, raw.ID)
			if err != nil {
				return nil, fmt.Errorf("error fetching client info for forwarder ID %s: %w", raw.ID, err)
			}

			results = append(results, clientInfo)
		}
	}

	return results, nil
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

// FindConnectionInfoByFrom finds a connection by its from field.
func (d *DB) FindConnectionInfoByFrom(channelID, from string) ([]*database.ConnectionInfo, error) {
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
		connections = append(connections, info.DeepCopy())
	}
	return connections, nil
}

// FindConnectionInfoByTo finds a connection by its to field.
func (d *DB) FindConnectionInfoByTo(channelID, to string) ([]*database.ConnectionInfo, error) {
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
		connections = append(connections, info.DeepCopy())
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

// FindForwarderInfo  finds a client by their ID.
func (d *DB) FindForwarderInfo(channelID string, fetcher string, maxForwardNum int) (*database.ClientInfo, error) {
	weights := map[string]float64{
		"connectionCount": 1.0, // example weight for connectionCount
		"createdTime":     0.5, // example weight for createdTime
		"networkSpeed":    0.5, // example weight for networkSpeed
		"packetLossRate":  0.3, // example weight for packetLossRate
		// Todo: insert more fields
	}
	optimalForwarder, err := d.findOptimalForwarder(channelID, fetcher, maxForwardNum, weights)
	if err == nil && optimalForwarder != nil {
		clientInfo, err := d.findClientInfoByID(optimalForwarder.ChannelID, optimalForwarder.ID)
		if err == nil {
			return clientInfo, nil
		}
		log.Printf("error in converting optimal forwarder to client info: %v", err)
	}
	txn := d.db.Txn(false)
	defer txn.Abort()
	return nil, nil
}

// findOptimalForwarder finds the best forwarder based on provided metrics and weights.
// User whose class is Forwarder or Potential Forwarder should be chosen.
func (d *DB) findOptimalForwarder(channelID, fetcher string, maxForwardNum int, weights map[string]float64) (*database.ClientInfo, error) { //nolint:lll
	txn := d.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(tblClients, idxClientChannelID, channelID)
	if err != nil {
		return nil, fmt.Errorf("find forwarders by channelID: %w", err)
	}

	var bestForwarder *database.ClientInfo
	var bestScore float64

	log.Printf("Starting optimal forwarder selection for channel: %s", channelID)

	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		candidate := raw.(*database.ClientInfo)

		log.Printf("Checking candidate: ID=%s, CanForward=%t, Class=%d, Fetcher=%s", candidate.ID, candidate.CanForward(), candidate.Class, fetcher) //nolint:lll

		if !candidate.CanForward() || candidate.ID == fetcher || candidate.ConnectionCount > maxForwardNum {
			log.Printf("Candidate %s skipped: Cannot forward or is fetcher", candidate.ID)
			continue
		}

		// Calculate the score using weights and available metrics
		score := calculateScore(candidate, weights)

		if bestForwarder == nil || score > bestScore {
			bestForwarder = candidate.DeepCopy()
			bestScore = score
		}
	}

	if bestForwarder == nil {
		log.Printf("no suitable forwarder found for channel %s", channelID)
		return nil, nil
	}
	log.Printf("Selected Best Forwarder: ID=%s, ChannelID=%s, Score=%f",
		bestForwarder.ID, bestForwarder.ChannelID, bestScore)
	return bestForwarder, nil
}

// calculateScore dynamically calculates the score based on available metrics and weights.
func calculateScore(forwarder *database.ClientInfo, weights map[string]float64) float64 {
	score := 0.0
	// Assign base scores based on role
	roleScores := map[int]float64{
		database.Forwarder:          5.0,
		database.PotentialForwarder: 3.0,
		database.Candidate:          0.0,
	}
	// Add role score if applicable
	if baseScore, ok := roleScores[forwarder.Class]; ok {
		score += baseScore
	}
	// Check if weights are provided for each metric, then calculate the score
	if weight, ok := weights["connectionCount"]; ok {
		score += weight / float64(forwarder.ConnectionCount+1) // Avoid division by zero
	}
	if weight, ok := weights["createdTime"]; ok {
		// Add score based on how long the forwarder has existed
		elapsedTime := time.Since(forwarder.CreatedAt).Minutes() // Minutes since creation
		score += weight * elapsedTime
	}

	if weight, ok := weights["networkSpeed"]; ok {
		// Replace forwarder.NetworkSpeed when implemented
		networkSpeed := float64(0)
		score += weight * networkSpeed
	}
	if weight, ok := weights["packetLossRate"]; ok {
		// Replace forwarder.PacketLossRate when implemented
		packetLossRate := float64(0)
		score -= weight * packetLossRate
	}

	return score
}

// CreateClassifyConnectionInfo creates a new connection between two clients.
func (d *DB) CreateClassifyConnectionInfo(channelID, from, to, connectionID string) (*database.ConnectionInfo, error) {
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
		Type:      database.Classify,
		CreatedAt: time.Now(),
	}
	if err := txn.Insert(tblConnections, newConn); err != nil {
		return nil, fmt.Errorf("insert connection: %w", err)
	}
	txn.Commit()
	return newConn.DeepCopy(), nil
}
