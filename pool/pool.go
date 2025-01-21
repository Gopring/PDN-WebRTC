// Package pool manages the sorted set of forwarder candidates and their scores.
package pool

import (
	"github.com/wangjia184/sortedset"
	"pdn/database"
	"sync"
	"time"
)

// Constants defining the bit allocation for score calculation.
const (
	ConnectionCountBits = 61
	CreatedAtBits       = 29
)

// channelSet manages a single channel's sorted set and its lock
type channelSet struct {
	mutex sync.RWMutex
	set   *sortedset.SortedSet
}

// Pool manages the sorted sets of forwarder candidates for each channel
type Pool struct {
	globalMutex sync.RWMutex
	sets        map[string]*channelSet
	database    database.Database
}

// New initializes a new Pool with a database reference
func New(db database.Database) *Pool {
	return &Pool{
		sets:     make(map[string]*channelSet),
		database: db,
	}
}

// getOrCreateSet ensures a channelSet exists for the given channelID
func (p *Pool) getOrCreateSet(channelID string) *channelSet {
	p.globalMutex.RLock()
	if cs, exists := p.sets[channelID]; exists {
		p.globalMutex.RUnlock()
		return cs
	}
	p.globalMutex.RUnlock()

	p.globalMutex.Lock()
	defer p.globalMutex.Unlock()
	cs, exists := p.sets[channelID]
	if !exists {
		cs = &channelSet{
			set: sortedset.New(),
		}
		p.sets[channelID] = cs
		return cs
	}

	return cs
}

// calculateScore calculates the score based on connection count and created time
func calculateScore(connectionCount int64, createdAt time.Time) int64 {
	elapsedSeconds := int64(time.Since(createdAt).Seconds())
	return (connectionCount << ConnectionCountBits) | (elapsedSeconds << CreatedAtBits)
}

// getConnectionCount retrieves the connection count for a client ID from the database
func (p *Pool) getConnectionCount(clientID, channelID string) (int64, error) {
	connections, err := p.database.FindAllPeerConnectionInfoByFrom(channelID, clientID)
	if err != nil {
		return 0, err
	}
	return int64(len(connections)), nil
}

// AddClient adds a new ClientInfo to the pool for a specific channel
func (p *Pool) AddClient(client database.ClientInfo) error {
	cs := p.getOrCreateSet(client.ChannelID)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	connectionCount, err := p.getConnectionCount(client.ID, client.ChannelID)
	if err != nil {
		return err
	}

	score := calculateScore(connectionCount, client.CreatedAt)
	cs.set.AddOrUpdate(client.ID, sortedset.SCORE(score), client)
	return nil
}

// UpdateClientScore recalculates the score for a specific client in a channel
func (p *Pool) UpdateClientScore(clientID, channelID string, maxForwardingNum int) error {
	cs := p.getOrCreateSet(channelID)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	node := cs.set.GetByKey(clientID)
	if node == nil {
		clientInfo, err := p.database.FindClientInfoByID(channelID, clientID)
		if err != nil {
			return err
		}
		if err := p.AddClient(*clientInfo); err != nil {
			return err
		}
		return nil
	}
	client := node.Value.(database.ClientInfo)
	connectionCount, err := p.getConnectionCount(clientID, channelID)
	if err != nil {
		return err
	}
	if connectionCount >= int64(maxForwardingNum) {
		cs.set.Remove(client.ID)
		return nil
	}
	newScore := calculateScore(connectionCount, client.CreatedAt)
	cs.set.AddOrUpdate(client.ID, sortedset.SCORE(newScore), client)
	return nil
}

// GetTopForwarder retrieves the highest scored forwarder for a specific channel
func (p *Pool) GetTopForwarder(channelID string) *database.ClientInfo {
	cs := p.getOrCreateSet(channelID)
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	topNode := cs.set.PeekMax()
	if topNode == nil {
		return nil
	}
	client := topNode.Value.(database.ClientInfo)
	return &client
}

// RemoveClient removes a client from the pool for a specific channel
func (p *Pool) RemoveClient(clientID, channelID string) {
	cs := p.getOrCreateSet(channelID)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.set.Remove(clientID)
}
