// Package classifier provides functionality to classify clients into roles such as forwarders or fetchers.
package classifier

import (
	"github.com/lithammer/shortuuid/v4"
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Classifier is responsible for managing the classification of
// clients into different roles (e.g., forwarders, fetchers).
type Classifier struct {
	config   Config
	broker   *broker.Broker
	database database.Database
}

// New creates a new instance of Classifier.
func New(c Config, b *broker.Broker, db database.Database) *Classifier {
	if c.TimeoutDuration == 0 {
		c.TimeoutDuration = DefaultTimeoutDuration
	}
	return &Classifier{
		config:   c,
		broker:   b,
		database: db,
	}
}

// Start starts the Classifier instance.
func (c *Classifier) Start() {
	classificationEvent := c.broker.Subscribe(broker.Classification, broker.CLASSIFY)
	classifyResultEvent := c.broker.Subscribe(broker.Classification, broker.CLASSIFYRESULT)
	for {
		select {
		case event := <-classificationEvent.Receive():
			go c.handleClassification(event)
		case event := <-classifyResultEvent.Receive():
			go c.handleClassifyResult(event)
		}
	}
}

// ClassifyPotentialForwarders identifies and classifies potential forwarders in a given channel.
// It attempts to connect potential forwarders with fetchers using a round-robin mechanism to ensure
// that fetchers are evenly distributed among potential forwarders.
// The round-robin approach ensures that no single fetcher is overloaded and that all fetchers
// are utilized fairly during the classification process.
func (c *Classifier) handleClassification(event any) {
	msg, ok := event.(message.ClassifyRequest)
	if !ok {
		log.Printf("Invalid classification request: %v", event)
		return
	}
	potentialForwarders, err := c.database.FindClientInfoByClass(msg.ChannelID, database.PotentialForwarder)
	if err != nil {
		log.Printf("Error fetching potential forwarders: %v", err)
		return
	}

	fetchers, err := c.database.FindClientInfoByClass(msg.ChannelID, database.Fetcher)
	if err != nil {
		log.Printf("Error fetching fetchers: %v", err)
		return
	}

	if len(fetchers) == 0 {
		log.Println("No fetchers available for classification.")
		return
	}

	currentFetcherIndex := 0
	for _, potential := range potentialForwarders {
		for i := 0; i < len(fetchers); i++ {
			fetcher := fetchers[currentFetcherIndex]
			currentFetcherIndex = (currentFetcherIndex + 1) % len(fetchers)

			log.Printf("Sending classification request: PotentialForwarder %s -> Fetcher %s", potential.ID, fetcher.ID)
			c.classify(potential, fetcher)
		}
	}
}

// classify attempts to establish a connection between a potential forwarder and a fetcher.
func (c *Classifier) classify(forwarder *database.ClientInfo, fetcher *database.ClientInfo) {
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(fetcher.ChannelID+fetcher.ID), response.ClassifyFetch{
		Type:         response.CLASSIFYFETCH,
		PeerID:       forwarder.ID,
		ConnectionID: shortuuid.New(),
	}); err != nil {
		log.Printf("error occurs in publishing fetch message %v", err)
	}
}

// handleClassifyResult processes classification results from fetchers.
func (c *Classifier) handleClassifyResult(event any) {
	result, ok := event.(message.ClassifyResult)
	if !ok {
		log.Printf("Invalid classify result: %v", event)
		return
	}

	if result.Success {
		log.Printf("PeerID %s classified successfully as Forwarder", result.PeerID)
		c.promoteToForwarder(result.PeerID, result.ChannelID)
	} else {
		log.Printf("PeerID %s classification failed, demoting to Fetcher", result.PeerID)
		c.demoteToFetcher(result.PeerID, result.ChannelID)
	}
}

// promoteToForwarder updates the client role to Forwarder in the database.
func (c *Classifier) promoteToForwarder(peerID, channelID string) {
	if err := c.database.UpdateClientInfoClass(channelID, peerID, database.Forwarder); err != nil {
		log.Printf("Error promoting PeerID %s to Forwarder: %v", peerID, err)
	}
}

// demoteToFetcher updates the client role to Fetcher in the database.
func (c *Classifier) demoteToFetcher(peerID, channelID string) {
	if err := c.database.UpdateClientInfoClass(channelID, peerID, database.Fetcher); err != nil {
		log.Printf("Error demoting PeerID %s to Fetcher: %v", peerID, err)
	}
}
