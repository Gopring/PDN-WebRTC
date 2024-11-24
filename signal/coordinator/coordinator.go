// Package coordinator contains handling socket more clearly
package coordinator

import (
	"fmt"
	"pdn/media"
	"pdn/pkg/socket"
	"time"
)

// TODO(window9u): we should add timeout configuration
const waitResponse = 10 * time.Second
const waitReceive = 10 * time.Second

// MemoryCoordinator managing socket. MemoryCoordinator relay messages between users.
type MemoryCoordinator struct {
	//TODO(window9u): we should add lock for channels
	//NOTE(window9u): MemoryCoordinator may manage user directly (channelID+userID
	// for key). But future, there is case of broadcast to all user of channel.
	// This is why managing users in channel
	channels map[string]*Channel
	media    media.Media
}

// New creates a new instance of MemoryCoordinator.
func New(med media.Media) *MemoryCoordinator {
	return &MemoryCoordinator{
		channels: map[string]*Channel{},
		media:    med,
	}
}

// Activate register and activate user
func (c *MemoryCoordinator) Activate(channelID string, userID string, s socket.Socket) error {
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

// Send process send signal
func (c *MemoryCoordinator) Send(channelID, userID, sdp string) (string, error) {
	return c.media.AddSender(channelID, userID, sdp)
}

// Receive process receive signal
func (c *MemoryCoordinator) Receive(channelID, userID, sdp string) (string, error) {
	return c.media.AddReceiver(channelID, userID, sdp)
}

// Forward process signal
func (c *MemoryCoordinator) Forward(channelID, userID, sdp string) (string, error) {
	return c.media.AddForwarder(channelID, userID, sdp)
}

// Fetch process a signal.
func (c *MemoryCoordinator) Fetch(channelID, _, fetcherSDP string) (string, error) {
	forwarderID, err := c.media.GetForwarder(channelID)
	if err != nil {
		return "", err
	}
	forwarderSDP, err := c.requestResponse(channelID, forwarderID, fetcherSDP)
	if err != nil {
		return "", err
	}
	return forwarderSDP, nil
}

// Arrange process arrange signal.
func (c *MemoryCoordinator) Arrange(channelID, userID, sdp string) (string, error) {
	err := c.response(channelID, userID, sdp)
	if err != nil {
		return "", err
	}
	return "", nil
}

// Reconnect process reconnect signal.
func (c *MemoryCoordinator) Reconnect(channelID, userID, sdp string) (string, error) {
	return c.media.AddReceiver(channelID, userID, sdp)
}

// RequestResponse send data to user and wait for response
func (c *MemoryCoordinator) requestResponse(channelID string, userID string, data string) (string, error) {
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

// Response send data to user
func (c *MemoryCoordinator) response(channelID, userID string, data string) error {
	user, err := c.getUser(channelID, userID)
	if err != nil {
		return err
	}
	if err := user.Response(data, waitReceive); err != nil {
		return fmt.Errorf("failed to answer %s", userID)
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
