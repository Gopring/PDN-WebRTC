// Package classifier provides functionality to classify clients into roles such as forwarders or fetchers.
package classifier

import (
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Classifier is responsible for managing the classification of
type Classifier struct {
	config   Config
	broker   *broker.Broker
	database database.Database
}

// New creates a new instance of Classifier.
func New(c Config, b *broker.Broker, db database.Database) *Classifier {
	return &Classifier{
		config:   c,
		broker:   b,
		database: db,
	}
}

// Start starts the Classifier instance.
func (c *Classifier) Start() {
	mediaConnectedEvent := c.broker.Subscribe(broker.Media, broker.CONNECTED)
	peerFailedEvent := c.broker.Subscribe(broker.Peer, broker.FAILED)
	classifyResultEvent := c.broker.Subscribe(broker.Classification, broker.CLASSIFIED)
	peerConnectedEvent := c.broker.Subscribe(broker.Peer, broker.CONNECTED)

	for {
		select {
		case event := <-mediaConnectedEvent.Receive():
			go c.handleMediaConnected(event)
		case event := <-peerFailedEvent.Receive():
			go c.handlePeerFailed(event)
		case event := <-classifyResultEvent.Receive():
			go c.handleClassifyResult(event)
		case event := <-peerConnectedEvent.Receive():
			go c.handlePeerConnected(event)
		}
	}
}

// handleMediaConnected handles events when a media connection is successfully established.
func (c *Classifier) handleMediaConnected(event any) {
	msg, ok := event.(message.Connected)
	if !ok {
		log.Printf("Invalid classification request: %v", event)
		return
	}
	connInfo, err := c.database.FindConnectionInfoByID(msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in finding connection info by connection id %v", err)
		return
	}
	Candidates, err := c.database.FindClientInfoByClass(connInfo.ChannelID, database.Candidate)
	if err != nil {
		log.Printf("Error fetching potential forwarders: %v", err)
		return
	}

	fetchers, err := c.database.FindClientInfoByClass(connInfo.ChannelID, database.Fetcher)
	if err != nil {
		log.Printf("Error fetching fetchers: %v", err)
		return
	}

	if len(fetchers) == 0 {
		log.Println("No fetchers available for classification.")
		return
	}

	currentFetcherIndex := 0
	for _, candidate := range Candidates {
		for i := 0; i < len(fetchers); i++ {
			fetcher := fetchers[currentFetcherIndex]
			currentFetcherIndex = (currentFetcherIndex + 1) % len(fetchers)

			log.Printf("Sending classification request: PotentialForwarder %s -> Fetcher %s", candidate.ID, fetcher.ID)
			if err := c.classify(candidate, fetcher); err != nil {
				return
			}
		}
	}
}

// handlePeerFailed handles events when a peer connection attempt fails.
func (c *Classifier) handlePeerFailed(event any) {
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

	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.From, database.Fetcher); err != nil {
		log.Printf("error occurs in updating client info %v", err)
		return
	}

	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.To, database.Fetcher); err != nil {
		log.Printf("error occurs in updating client info %v", err)
		return
	}

	Candidates, err := c.database.FindClientInfoByClass(connInfo.ChannelID, database.Candidate)
	if err != nil {
		log.Printf("Error fetching potential forwarders: %v", err)
		return
	}

	fetchers, err := c.database.FindClientInfoByClass(connInfo.ChannelID, database.Fetcher)
	if err != nil {
		log.Printf("Error fetching fetchers: %v", err)
		return
	}

	if len(fetchers) == 0 {
		log.Println("No fetchers available for classification.")
		return
	}

	currentFetcherIndex := 0
	for _, candidate := range Candidates {
		for i := 0; i < len(fetchers); i++ {
			fetcher := fetchers[currentFetcherIndex]
			currentFetcherIndex = (currentFetcherIndex + 1) % len(fetchers)

			log.Printf("Sending classification request: PotentialForwarder %s -> Fetcher %s", candidate.ID, fetcher.ID)
			if err := c.classify(candidate, fetcher); err != nil {
				return
			}
		}
	}
}

// handlePeerConnected handles events when a peer-to-peer connection is successfully established.
func (c *Classifier) handlePeerConnected(event any) {
	msg, ok := event.(message.Connected)
	if !ok {
		log.Printf("error occurs in parsing failed message %v", event)
		return
	}
	connInfo, _ := c.database.FindConnectionInfoByID(msg.ConnectionID)
	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.To, database.Candidate); err != nil {
		log.Printf("error occurs in updating client info %v", err)
	}
	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.From, database.Candidate); err != nil { //nolint:lll
		log.Printf("error occurs in updating client info %v", err)
	}
}

// classify attempts to establish a connection between a candidate and a fetcher.
func (c *Classifier) classify(forwarder *database.ClientInfo, fetcher *database.ClientInfo) error {

	classifyConn, err := c.database.CreateClassifyConnectionInfo(fetcher.ChannelID, forwarder.ID, fetcher.ID, shortuuid.New()) //nolint:lll
	if err != nil {
		return fmt.Errorf("error occurs in creating peer connection info %v", err)
	}
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(fetcher.ChannelID+fetcher.ID), response.Classified{
		Type:         response.CLASSIFIED,
		ConnectionID: classifyConn.ID,
	}); err != nil {
		log.Printf("error occurs in publishing fetch message %v", err)
	}
	return nil
}

// handleClassifyResult processes classification results from fetchers.
func (c *Classifier) handleClassifyResult(event any) {
	msg, ok := event.(message.Classified)
	if !ok {
		log.Printf("Invalid classify result: %v", event)
		return
	}
	connInfo, err := c.database.FindConnectionInfoByID(msg.ConnectionID)
	if err != nil {
		log.Printf("error occurs in finding connection info by connection id %v", err)
		return
	}
	if msg.Success {
		log.Printf("PeerID %s classified successfully as Forwarder", connInfo.From)
		c.promoteToForwarder(connInfo.From, msg.ChannelID)
	} else {
		log.Printf("PeerID %s classification failed, demoting to Fetcher", connInfo.From)
		c.demoteToFetcher(connInfo.From, msg.ChannelID)
	}
}

// promoteToForwarder updates the client class to Forwarder in the database.
func (c *Classifier) promoteToForwarder(peerID, channelID string) {
	if err := c.database.UpdateClientInfoClass(channelID, peerID, database.Forwarder); err != nil {
		log.Printf("Error promoting PeerID %s to Forwarder: %v", peerID, err)
	}
}

// demoteToFetcher updates the client class to Fetcher in the database.
func (c *Classifier) demoteToFetcher(peerID, channelID string) {
	if err := c.database.UpdateClientInfoClass(channelID, peerID, database.Fetcher); err != nil {
		log.Printf("Error demoting PeerID %s to Fetcher: %v", peerID, err)
	}
}
