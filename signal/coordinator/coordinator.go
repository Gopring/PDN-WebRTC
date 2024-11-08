package coordinator

import (
	"fmt"
	"pdn/signal/controller/socket"
	"time"
)

type Coordinator struct {
	channels map[string]*Channel
}

type Channel struct {
	users map[string]*User
}

type User struct {
	socket   *socket.Socket
	response chan string
}

func (u *User) SendToSocket(data string) error {
	err := u.socket.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) WaitForResponse(d time.Duration) (string, error) {
	select {
	case sdp := <-u.response:
		return sdp, nil
	case <-time.After(d):
		return "", fmt.Errorf("timeout")
	}
}

func (u *User) Enqueue(data string) error {
	u.response <- data
	return nil
}

func (c *Coordinator) RequestResponse(channelID string, userID string, data string) (string, error) {
	channel, exists := c.channels[channelID]
	if !exists {
		return "", fmt.Errorf("channel %s doesn't exists", channelID)
	}

	user, exists := channel.users[userID]
	if !exists {
		return "", fmt.Errorf("user %s doesn't exists", userID)
	}

	if err := user.SendToSocket(data); err != nil {
		return "", fmt.Errorf("failed to send user")
	}

	sdp, err := user.WaitForResponse(10 * time.Second)
	if err != nil {
		return "", err
	}

	return sdp, nil
}

func (c *Coordinator) AddUser(channelID string, userID string, s *socket.Socket) error {
	_, exists := c.channels[channelID]
	if !exists {
		c.channels[channelID] = &Channel{}
	}
	ch := c.channels[channelID]
	ch.users[userID] = &User{
		socket:   s,
		response: make(chan string),
	}
	return nil
}

func (c *Coordinator) Deliver(channelID, userID string, data string) error {
	channel, exists := c.channels[channelID]
	if !exists {
		return fmt.Errorf("channel %s doesn't exists", channelID)
	}

	user, exists := channel.users[userID]
	if !exists {
		return fmt.Errorf("user %s doesn't exists", userID)
	}

	if err := user.Enqueue(data); err != nil {
		return fmt.Errorf("failed to answer %s", userID)
	}
	return nil
}

func (c *Coordinator) Remove(channelID, userID string) error {
	channel, exists := c.channels[channelID]
	if !exists {
		return fmt.Errorf("channel %s doesn't exists", channelID)
	}
	delete(channel.users, userID)
	return nil
}
