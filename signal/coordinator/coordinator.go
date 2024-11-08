package coordinator

import "pdn/signal/controller/socket"

type Coordinator struct {
	channels map[string]string
	user     map[string]*socket.Socket
}

func (c *Coordinator) Get(channelID string, userID string, data string) (string, error) {

	return "", nil
}

func (c *Coordinator) Add(id string, s *socket.Socket) error {

}

func (c *Coordinator) Send(id string, s *socket.Socket) error {

}

func (c *Coordinator) Remove(id string) error {

}
