// Package coordinator manages the WebRTC connections that Client to Media server and Client to Client.
package coordinator

import (
	"github.com/lithammer/shortuuid/v4"
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/types/client/response"
	"pdn/types/message"
)

const MaxForwardingNumber = 3

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
	msg, ok := event.(message.Push)
	if !ok {
		log.Printf("error occurs in parsing push message %v", event)
		return
	}

	connInfo, err := c.database.CreatePushConnectionInfo(msg.ChannelID, msg.ClientID, msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in creating connection info %v", err)
		return
	}

	if _, err := c.database.UpdateClientInfo(msg.ChannelID, msg.ClientID, database.Publisher); err != nil {
		log.Printf("error occurs in updating client info %v", err)
		return
	}

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
	msg, ok := event.(message.Pull)
	if !ok {
		log.Printf("error occurs in parsing pull message %v", event)
		return
	}

	connInfo, err := c.database.CreatePullConnectionInfo(msg.ChannelID, msg.ClientID, msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in creating connection info %v", err)
		return
	}

	streamInfo, err := c.database.FindUpstreamInfo(msg.ChannelID)
	if err != nil {
		log.Printf("error occurs in finding upstream info %v", err)
		return
	}

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
	msg, ok := event.(message.Connected)
	if !ok {
		log.Printf("error occurs in parsing connected message %v", event)
		return
	}

	connInfo, err := c.database.UpdateConnectionInfo(msg.ConnectionID, database.Connected)
	if err != nil {
		log.Printf("error occurs in update connection info %v", err)
		return
	}

	if connInfo.IsUpstream() {
		return
	}

	clientInfo, err := c.database.FindForwarderInfo(connInfo.ChannelID, connInfo.To, MaxForwardingNumber)
	if err != nil {
		log.Printf("error occurs in finding user info to forward %v", err)
		return
	}
	if clientInfo == nil {
		log.Printf("no user info to forward")
		return
	}
	log.Printf("forwarding to %v", clientInfo)

	peerConn, err := c.database.CreatePeerConnectionInfo(connInfo.ChannelID, clientInfo.ID, connInfo.To, shortuuid.New())
	if err != nil {
		log.Printf("error occurs in creating connection info between two clients %v", err)
		return
	}

	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(connInfo.ChannelID+connInfo.To), response.Fetch{
		Type:         response.FETCH,
		ConnectionID: peerConn.ID,
	}); err != nil {
		log.Printf("error occurs in publishing forward message %v", err)
		return
	}
}

// handleDisconnected handles the disconnected event. This event is about Media server to client
func (c *Coordinator) handleDisconnected(event any) {
	// 01. Parse the event to message.Disconnected
	msg, ok := event.(message.Disconnected)
	if !ok {
		log.Printf("error occurs in parsing disconnected message %v", event)
		return
	}
	// 02. Find user info by connection id
	if err := c.database.DeleteConnectionInfoByID(msg.ConnectionID); err != nil {
		log.Printf("error occurs in finding connection info by connection id %v", err)
		return
	}

	// We should also delete fetcher connection info if connection was forwarder
}

// handleFailed handles the failed event. This event is about client to client
func (c *Coordinator) handleFailed(event any) {
	msg, ok := event.(message.Failed)
	if !ok {
		log.Printf("error occurs in parsing failed message %v", event)
		return
	}

	connInfo, err := c.database.FindConnectionInfoByID(msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in finding connection info by connection id %v", err)
		return
	}

	if _, err := c.database.UpdateClientInfo(connInfo.ChannelID, connInfo.From, database.Fetcher); err != nil {
		log.Printf("error occurs in updating client info %v", err)
		return
	}

	if _, err := c.database.UpdateClientInfo(connInfo.ChannelID, connInfo.To, database.Fetcher); err != nil {
		log.Printf("error occurs in updating client info %v", err)
		return
	}

	// coordinate again
}

// handleSucceed handles the succeed event. This event is about client to client
func (c *Coordinator) handleSucceed(event any) {
	msg, ok := event.(message.Succeed)
	if !ok {
		log.Printf("error occurs in parsing failed message %v", event)
		return
	}

	peerConn, err := c.database.UpdateConnectionInfo(msg.ConnectionID, database.Connected)
	if err != nil {
		log.Printf("error occurs in updating connection info %v", err)
		return
	}

	serverConn, err := c.database.FindDownstreamInfo(peerConn.ChannelID, peerConn.To)
	if err != nil {
		log.Printf("error occurs in finding downstream info %v", err)
		return
	}

	if err := c.broker.Publish(broker.Media, broker.CLOSURE, message.Closure{
		ConnectionID: serverConn.ID,
	}); err != nil {
		log.Printf("error occurs in publishing closure message %v", err)
		return
	}
}
