package coordinator

import (
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/types/message"
)

// Coordinator manages the WebRTC connections that Client to Media server and Client to Client.
type Coordinator struct {
	broker   *broker.Broker
	database database.Database
}

// New creates a new instance of Coordinator.
func New(b *broker.Broker, db database.Database) *Coordinator {
	return &Coordinator{
		broker:   b,
		database: db,
	}
}

// Run starts the Coordinator instance.
func (c *Coordinator) Run() {
	pushEvent := c.broker.Subscribe(broker.ClientMessage, broker.PUSH)
	pullEvent := c.broker.Subscribe(broker.ClientMessage, broker.PULL)
	connectedEvent := c.broker.Subscribe(broker.Media, broker.CONNECTED)
	disconnectedEvent := c.broker.Subscribe(broker.Media, broker.DISCONNECTED)
	failedEvent := c.broker.Subscribe(broker.Connection, broker.FAILED)
	succeedEvent := c.broker.Subscribe(broker.Connection, broker.SUCCEED)
	for {
		select {
		case event := <-pushEvent.Receive():
			go c.handlePush(event)
		case event := <-pullEvent.Receive():
			go c.handlePull(event)
		case event := <-connectedEvent.Receive():
			go c.handleConnected(event)
		case event := <-disconnectedEvent.Receive():
			go c.handleDisconnected(event)
		case event := <-failedEvent.Receive():
			go c.handleFailed(event)
		case event := <-succeedEvent.Receive():
			go c.handleSucceed(event)
		}
	}
}

// handlePush handles the push event. push event means that a client requests
// to push stream to Media server.
func (c *Coordinator) handlePush(event any) {
	// 01. Parse the event to message.Push
	msg, ok := event.(message.Push)
	if !ok {
		log.Printf("error occurs in parsing push message %v", event)
		return
	}

	// 02. Create connection info
	connInfo, err := c.database.CreateServerConnectionInfo(true, msg.ChannelID, msg.ClientID, msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in creating connection info %v", err)
		return
	}

	// 03. Publish event to Media server
	if err := c.broker.Publish(broker.Media, broker.UPSTREAM, message.Upstream{
		ConnectionID: connInfo.ID,
		Key:          connInfo.ChannelID + connInfo.From,
		SDP:          msg.SDP,
	}); err != nil {
		log.Printf("error occurs in publishing push message %v", err)
		return
	}
}

// handlePull handles the pull event. pull event means that a client requests
// to pull stream. Currently, stream is pulled only from Media server. In the
// future, it could be pulled from other clients directly.
func (c *Coordinator) handlePull(event any) {
	// 01. Parse the event to message.Pull
	msg, ok := event.(message.Pull)
	if !ok {
		log.Printf("error occurs in parsing pull message %v", event)
		return
	}

	// 02. Create connection info
	connInfo, err := c.database.CreateServerConnectionInfo(false, msg.ChannelID, msg.ClientID, msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in creating connection info %v", err)
		return
	}

	streamInfo, err := c.database.FindUpstreamInfo(msg.ChannelID)
	if err != nil {
		log.Printf("error occurs in finding upstream info %v", err)
		return
	}

	// 03. Publish event to Media server
	if err := c.broker.Publish(broker.Media, broker.DOWNSTREAM, message.Downstream{
		ConnectionID: connInfo.ID,
		StreamID:     streamInfo.ID,
		Key:          connInfo.ChannelID + connInfo.To,
		SDP:          msg.SDP,
	}); err != nil {
		log.Printf("error occurs in publishing pull message %v", err)
		return
	}
}

// handleConnected handles the connected event. This event is about Media server to client
func (c *Coordinator) handleConnected(event any) {
	// 01. Parse the event to message.Connected
	msg, ok := event.(message.Connected)
	if !ok {
		log.Printf("error occurs in parsing connected message %v", event)
		return
	}

	// 02. Update connection info
	if err := c.database.UpdateConnectionInfo(true, msg.ConnectionID); err != nil {
		log.Printf("error occurs in update connection info %v", err)
		return
	}

	//// 03. Find user info to forward
	//userInfo, err := c.database.FindClientInfoToForward(msg.ChannelID, msg.ClientID)
	//if err != nil {
	//	log.Printf("error occurs in finding user info to forward %v", err)
	//	return
	//}
	//if userInfo == nil {
	//	log.Printf("no user info to forward")
	//	return
	//}
	//
	//// 04. Create connection info between two clients
	//connInfo, err := c.database.CreateClientConnectionInfo(msg.ChannelID, userInfo.ID, msg.ClientID, shortuuid.New())
	//if err != nil {
	//	log.Printf("error occurs in creating connection info between two clients %v", err)
	//	return
	//}
	//
	//// 05. Publish forward and fetch message
	//if err := c.broker.Publish(broker.ClientSocket, broker.Detail(msg.ChannelID+msg.ClientID), response.Fetch{
	//	ConnectionID: connInfo.ID,
	//	From:         userInfo.ID,
	//}); err != nil {
	//	log.Printf("error occurs in publishing forward message %v", err)
	//	return
	//}
	//if err := c.broker.Publish(broker.ClientSocket, broker.Detail(msg.ChannelID+userInfo.ID), response.Forward{
	//	ConnectionID: connInfo.ID,
	//	To:           msg.ClientID,
	//}); err != nil {
	//	log.Printf("error occurs in publishing forward message %v", err)
	//	return
	//}
}

// handleDisconnected handles the disconnected event. This event is about Media server to client
func (c *Coordinator) handleDisconnected(event any) {
	// 01. Parse the event to message.Disconnected
	//msg, ok := event.(message.Disconnected)
	//if !ok {
	//	log.Printf("error occurs in parsing disconnected message %v", event)
	//	return
	//}

	//// 02. Find user info by channel id and user id
	//userInfo, err := c.database.FindClientInfoByID(msg.ChannelID, msg.ClientID)
	//if err != nil {
	//	log.Printf("error occurs in finding user info by channel id and user id %v", err)
	//	return
	//}
}

// handleFailed handles the failed event. This event is about client to client
func (c *Coordinator) handleFailed(event any) {
	// 01. Parse the event to message.Failure
	//msg, ok := event.(message.Failure)
	//if !ok {
	//	log.Printf("error occurs in parsing failed message %v", event)
	//	return
	//}

	//// 02. Find user info by channel id and user id
	//userInfo, err := c.database.FindClientInfoByID(msg.Channel
}

func (c *Coordinator) handleSucceed(event any) {

}
