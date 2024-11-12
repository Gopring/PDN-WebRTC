// Package coordinator contains handling socket more clearly
package coordinator

import (
	"fmt"
	"pdn/signal/controller/socket"
	"time"
)

// TODO(window9u): we should add timeout configuration
const waitResponse = 10 * time.Second
const waitReceive = 10 * time.Second

// Coordinator is an interface for managing socket.
type Coordinator interface {
	RequestResponse(channelID string, userID string, data string) (string, error)
	AddUser(channelID string, userID string, s *socket.Socket) error
	Response(channelID, userID string, data string) error
	Remove(channelID, userID string) error
}

// MemoryCoordinator managing socket. MemoryCoordinator relay messages between users.
type MemoryCoordinator struct {
	//TODO(window9u): we should add lock for channels
	//NOTE(window9u): MemoryCoordinator may manage user directly (channelID+userID
	// for key). But future, there is case of broadcast to all user of channel.
	// This is why managing users in channel
	channels map[string]*Channel
}

// New creates a new instance of MemoryCoordinator.
func New() *MemoryCoordinator {
	return &MemoryCoordinator{
		channels: map[string]*Channel{},
	}
}

// RequestResponse send data to user and wait for response
func (c *MemoryCoordinator) RequestResponse(channelID string, userID string, data string) (string, error) {
	user, err := c.getUser(channelID, userID)
	if err != nil {
		return "", err
	}

	if err := user.Request(data); err != nil {
		return "", fmt.Errorf("failed to send user")
	}

	sdp, err := user.WaitForResponse(waitResponse)
	if err != nil {
		return "", err
	}

	return sdp, nil
}

// AddUser adds user to channel
func (c *MemoryCoordinator) AddUser(channelID string, userID string, s *socket.Socket) error {
	_, exists := c.channels[channelID]
	if !exists {
		c.channels[channelID] = &Channel{
			users: map[string]*User{},
		}
	}
	ch := c.channels[channelID]
	ch.users[userID] = &User{
		socket:   s,
		response: make(chan string),
	}
	return nil
}

// Response send data to user
func (c *MemoryCoordinator) Response(channelID, userID string, data string) error {
	user, err := c.getUser(channelID, userID)
	if err != nil {
		return err
	}
	if err := user.Response(data, waitReceive); err != nil {
		return fmt.Errorf("failed to answer %s", userID)
	}
	return nil
}

// Remove removes user from channel
func (c *MemoryCoordinator) Remove(channelID, userID string) error {
	channel, exists := c.channels[channelID]
	if !exists {
		return fmt.Errorf("channel %s doesn't exists", channelID)
	}
	delete(channel.users, userID)
	if len(channel.users) == 0 {
		delete(c.channels, channelID)
	}
	return nil
}

// getUser returns user from channel
func (c *MemoryCoordinator) getUser(channelID, userID string) (*User, error) {
	channel, exists := c.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel %s doesn't exists", channelID)
	}

	user, exists := channel.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s doesn't exists", userID)
	}
	return user, nil
}
