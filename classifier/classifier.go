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
	"time"
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
	go c.StartCronJob()

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

// StartCronJob starts a periodic task that classifies clients.
// It uses a ticker to trigger the classification task every minute.
func (c *Classifier) StartCronJob() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running periodic classification task")
		c.handleCronJob()
	}
}

// handleCronJob performs periodic classification tasks.
func (c *Classifier) handleCronJob() {
	channels, err := c.database.FindAllChannelInfos()
	if err != nil {
		log.Println(err)
	}
	for _, channel := range channels {
		log.Printf("Processing channel %s", channel.ID)

		candidates, err := c.database.FindAllClientInfoByClass(channel.ID, database.Candidate)
		if err != nil {
			log.Printf("Error fetching candidates for channel %s: %v", channel.ID, err)
			continue
		}

		fetchers, err := c.database.FindAllClientInfoByClass(channel.ID, database.Fetcher)
		if err != nil {
			log.Printf("Error fetching fetchers for channel %s: %v", channel.ID, err)
			continue
		}

		if len(candidates) == 0 || len(fetchers) == 0 {
			log.Printf("No candidates or fetchers available for channel %s. Skipping classification.", channel.ID)
			continue
		}

		currentFetcherIndex := 0
		for _, candidate := range candidates {
			for i := 0; i < len(fetchers); i++ {
				fetcher := fetchers[currentFetcherIndex]
				currentFetcherIndex = (currentFetcherIndex + 1) % len(fetchers)

				log.Printf("Classifying Candidate %s with Fetcher %s for channel %s", candidate.ID, fetcher.ID, channel.ID)
				if err := c.classify(candidate, fetcher); err != nil {
					log.Printf("Error during classification for Candidate %s with Fetcher %s: %v", candidate.ID, fetcher.ID, err)
				}
			}
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
	obj, err := c.database.FindClientInfoByID(connInfo.ChannelID, connInfo.To)
	if err != nil || obj == nil {
		log.Printf("Error finding client info for ID %s: %v", connInfo.To, err)
		return
	}
	fetcher, err := c.database.FindClientInfoByClass(connInfo.ChannelID, database.Fetcher)
	if err != nil || fetcher == nil {
		log.Printf("error occurs in finding client info by class %v", err)
		return
	}

	if err := c.classify(obj, fetcher); err != nil {
		return
	}
}

// handlePeerFailed handles events when a peer connection attempt fails.
func (c *Classifier) handlePeerFailed(event any) {
	msg, ok := event.(message.Failed)
	if !ok {
		log.Printf("Invalid failed message event: %v", event)
		return
	}

	// Find connection info based on the connection ID
	connInfo, err := c.database.FindConnectionInfoByID(msg.ConnectionID)
	if err != nil {
		log.Printf("Error finding connection info by connection ID %v: %v", msg.ConnectionID, err)
		return
	}

	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.From, database.Fetcher); err != nil {
		log.Printf("Error updating peer %s to Fetcher: %v", connInfo.From, err)
		return
	}

	if err := c.database.UpdateClientInfoClass(connInfo.ChannelID, connInfo.To, database.Fetcher); err != nil {
		log.Printf("Error updating peer %s to Fetcher: %v", connInfo.To, err)
		return
	}

	fetcherFirst, err := c.database.FindClientInfoByID(connInfo.ChannelID, connInfo.From)
	if err != nil {
		log.Printf("Error finding fetcherFirst %s in channel %s: %v", connInfo.From, connInfo.ChannelID, err)
		return
	}

	fetcherSecond, err := c.database.FindClientInfoByID(connInfo.ChannelID, connInfo.To)
	if err != nil {
		log.Printf("Error finding fetcherSecond %s in channel %s: %v", connInfo.To, connInfo.ChannelID, err) //nolint:govet
		return
	}

	candidates, err := c.database.FindAllClientInfoByClass(connInfo.ChannelID, database.Candidate)
	if err != nil {
		log.Printf("Error fetching candidates for channel %s: %v", connInfo.ChannelID, err)
		return
	}

	if len(candidates) == 0 {
		log.Printf("No candidates available for channel %s. Skipping classification.", connInfo.ChannelID)
		return
	}

	candidateFirst := candidates[0]
	log.Printf("Classifying Candidate %s with Peer %s in channel %s", candidateFirst.ID, fetcherFirst.ID, connInfo.ChannelID) //nolint:lll
	if err := c.classify(candidateFirst, fetcherFirst); err != nil {                                                          //nolint:lll
		log.Printf("Error classifying Candidate %s with Peer %s: %v", candidateFirst.ID, fetcherFirst.ID, err)
	}

	if len(candidates) > 1 {
		candidateSecond := candidates[1]
		log.Printf("Classifying Candidate %s with Peer %s in channel %s", candidateSecond.ID, fetcherSecond.ID, connInfo.ChannelID) //nolint:lll
		if err := c.classify(candidateSecond, fetcherSecond); err != nil {                                                          //nolint:lll
			log.Printf("Error classifying Candidate %s with Peer %s: %v", candidateSecond.ID, fetcherSecond.ID, err)
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
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(fetcher.ChannelID+fetcher.ID), response.Classifying{
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
