// Package media manages streams and connections using WebRTC.
package media

import (
	"fmt"
	"log"
	"pdn/media/stream"
	"pdn/metric"
	"sync"

	"github.com/pion/webrtc/v4"
	"pdn/broker"
	"pdn/types/client/response"
	"pdn/types/message"
)

// Media manages streams and connection configurations.
// NOTE: In the future, the media package could be detached from pdn
// and be used as a standalone package.
type Media struct {
	mu               sync.RWMutex
	broker           *broker.Broker
	metric           *metric.Metrics
	streams          map[string]*stream.Stream
	connections      map[string]*webrtc.PeerConnection
	connectionConfig webrtc.Configuration
	config           Config
}

// Default WebRTC configuration.
var defaultWebrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// New creates a new Media instance.
// TODO: Add more configuration options.
func New(c Config, b *broker.Broker, m *metric.Metrics) *Media {
	return &Media{
		config:           c,
		broker:           b,
		metric:           m,
		streams:          make(map[string]*stream.Stream),
		connections:      make(map[string]*webrtc.PeerConnection),
		connectionConfig: defaultWebrtcConfig,
	}
}

// Start starts the Media instance.
func (m *Media) Start() {
	upEvent := m.broker.Subscribe(broker.Media, broker.UPSTREAM)
	downEvent := m.broker.Subscribe(broker.Media, broker.DOWNSTREAM)
	clearEvent := m.broker.Subscribe(broker.Media, broker.CLEAR)
	closeEvent := m.broker.Subscribe(broker.Media, broker.CLOSE)

	for {
		var err error
		select {
		case event := <-upEvent.Receive():
			go m.handleUpstream(event)
		case event := <-downEvent.Receive():
			go m.handleDownstream(event)
		case event := <-clearEvent.Receive():
			go m.handleClear(event)
		case event := <-closeEvent.Receive():
			go m.handleCloseChannel(event)
		}
		if err != nil {
			log.Printf("Failed to handle event in Media: %v", err)
		}
	}
}

// handleUpstream handles a push event.
func (m *Media) handleUpstream(event any) {
	up, ok := event.(message.Upstream)
	if !ok {
		log.Printf("failed to cast event to Upstream: %v", event)
		return
	}
	serverSDP, err := m.AddUpstream(up.ConnectionID, up.SDP)
	if err != nil {
		log.Printf("failed to add upstream: %v", err)
		return
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(up.Key), response.Signal{
		Type:         response.SIGNAL,
		ConnectionID: up.ConnectionID,
		SignalType:   "answer",
		SignalData:   serverSDP,
	}); err != nil {
		log.Printf("failed to publish up response: %v", err)
		return
	}
}

// handleDownstream handles a pull event.
func (m *Media) handleDownstream(event any) {
	down, ok := event.(message.Downstream)
	if !ok {
		log.Printf("failed to cast event to Downstream: %v", event)
		return
	}
	log.Printf("AddDownstream called with connectionID: %s, streamID: %s, SDP: %s", down.ConnectionID, down.StreamID, down.SDP)
	serverSDP, err := m.AddDownstream(down.ConnectionID, down.StreamID, down.SDP)
	if err != nil {
		log.Printf("failed to add downstream: %v", err)
		return
	}
	if err := m.broker.Publish(broker.ClientSocket, broker.Detail(down.Key), response.Signal{
		Type:         response.SIGNAL,
		ConnectionID: down.ConnectionID,
		SignalType:   "answer",
		SignalData:   serverSDP,
	}); err != nil {
		log.Printf("failed to publish down response: %v", err)
		return
	}
}

// handleClear handles a close event.
func (m *Media) handleClear(event any) {
	clr, ok := event.(message.Clear)
	if !ok {
		log.Printf("failed to cast event to Closed: %v", event)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	conn, ok := m.connections[clr.ConnectionID]
	if !ok {
		log.Printf("connection not found: %s", clr.ConnectionID)
		return
	}
	log.Printf("Media: closing connection: %s", clr.ConnectionID)
	if err := conn.Close(); err != nil {
		log.Printf("failed to clr connection: %v", err)
	}
	delete(m.connections, clr.ConnectionID)
}

// AddUpstream creates a new upstream connection and adds it to the channel.
func (m *Media) AddUpstream(connectionID, sdp string) (string, error) {
	conn, err := m.createPushConn(connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to create connection: %w", err)
	}

	s, err := m.createUpstream(conn, connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}

	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	m.registerConnection(connectionID, conn)
	m.registerStream(connectionID, s)

	return conn.LocalDescription().SDP, nil
}

func (m *Media) handleCloseChannel(event any) {
	clr, ok := event.(message.Close)
	if !ok {
		log.Printf("failed to cast event to Close: %v", event)
		return
	}
	connectionID := clr.ConnectionID

	m.mu.Lock()
	defer m.mu.Unlock()
	conn, ok := m.connections[connectionID]
	if !ok {
		log.Printf("connection not found: %s", connectionID)
		return
	}
	log.Printf("Media: closing connection: %s", connectionID)
	if err := conn.Close(); err != nil {
		log.Printf("failed to clr connection: %v", err)
	}
	delete(m.connections, connectionID)
	delete(m.streams, connectionID)
	log.Printf("remove connection: %s and stream: %s in Media", connectionID, connectionID)
}

// AddDownstream creates a new downstream connection and adds it to the channel.
func (m *Media) AddDownstream(connectionID, streamID, sdp string) (string, error) {
	conn, err := m.createPullConn(connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to create connection: %w", err)
	}

	if err = m.setDownstream(conn, streamID); err != nil {
		return "", fmt.Errorf("failed to set downstream: %w", err)
	}

	if err = StartICE(conn, sdp); err != nil {
		return "", fmt.Errorf("failed to start ICE: %w", err)
	}

	m.registerConnection(connectionID, conn)
	return conn.LocalDescription().SDP, nil
}

// createPushConn creates a new connection.
func (m *Media) createPushConn(connectionID string) (*webrtc.PeerConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connections[connectionID]; ok {
		return nil, fmt.Errorf("connection already exists: %s", connectionID)
	}
	conn, err := m.NewInboundConnection(m.connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbound connection: %w", err)
	}
	m.publishStateChange(conn, connectionID)
	return conn, nil
}

func (m *Media) createPullConn(connectionID string) (*webrtc.PeerConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connections[connectionID]; ok {
		return nil, fmt.Errorf("connection already exists: %s", connectionID)
	}
	conn, err := m.NewOutboundConnection(m.connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbound connection: %w", err)
	}
	m.publishStateChange(conn, connectionID)
	return conn, nil
}

// createUpstream adds a channel to the media.
func (m *Media) createUpstream(conn *webrtc.PeerConnection, connectionID string) (*stream.Stream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.streams[connectionID]
	if ok {
		return nil, fmt.Errorf("upstream already exists: %s", connectionID)
	}
	s := stream.New()
	s.SetUpstream(conn, connectionID)
	return s, nil
}

// setDownstream sets a downstream connection.
func (m *Media) setDownstream(conn *webrtc.PeerConnection, streamID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.streams[streamID]
	if !ok {
		return fmt.Errorf("upstream does not exist: %s", streamID)
	}
	if err := s.SetDownstream(conn); err != nil {
		return fmt.Errorf("failed to set downstream: %w", err)
	}
	return nil
}

// publishStateChange publishes the state change of a connection.
func (m *Media) publishStateChange(conn *webrtc.PeerConnection, connectionID string) {
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		//log.Printf("Peer %s: ICE Peer State has changed to %s", connectionID, state.String())
		switch state {
		case webrtc.PeerConnectionStateConnected:
			log.Printf("Media: connection %s: Connected", connectionID)
			m.metric.IncrementWebRTCConnections()
			if err := m.broker.Publish(broker.Media, broker.CONNECTED, message.Connected{
				ConnectionID: connectionID,
			}); err != nil {
				log.Printf("failed to publish connected message: %v", err)
			}
		case webrtc.PeerConnectionStateClosed:
			log.Printf("Media: connection %s: Closed", connectionID)
			m.metric.DecrementWebRTCConnections()
			//if err := m.broker.Publish(broker.Media, broker.DISCONNECTED, message.Disconnected{
			//	ConnectionID: connectionID,
			//}); err != nil {
			//	log.Printf("failed to publish disconnected message: %v", err)
			//}
		case webrtc.PeerConnectionStateDisconnected:
			log.Printf("Media: connection %s: Disconnected", connectionID)
		case webrtc.PeerConnectionStateFailed:
			log.Printf("Media: connection %s: Failed", connectionID)
		default:
		}
	})
}

// registerConnection registers a connection.
func (m *Media) registerConnection(connectionID string, conn *webrtc.PeerConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[connectionID] = conn
}

// registerStream registers a stream.
func (m *Media) registerStream(connectionID string, s *stream.Stream) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streams[connectionID] = s
}
