// Package coordinator manages the WebRTC connections that Client to Media server and Client to Client.
package coordinator

import (
	"errors"
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	"log"
	"pdn/broker"
	"pdn/database"
	"pdn/metric"
	"pdn/pool"
	"pdn/types/client/response"
	"pdn/types/message"
	"runtime/debug"
)

var (
	// ErrNoForwarder is an error that occurs when there is no forwarder.
	ErrNoForwarder = fmt.Errorf("no forwarder")
)

// Coordinator manages the WebRTC connections that Client to Media server and Client to Client.
type Coordinator struct {
	config   Config
	broker   *broker.Broker
	metric   *metric.Metrics
	database database.Database
	pool     *pool.Pool
}

// New creates a new instance of Coordinator.
func New(c Config, b *broker.Broker, m *metric.Metrics, db database.Database, p *pool.Pool) *Coordinator {
	return &Coordinator{
		config:   c,
		broker:   b,
		metric:   m,
		database: db,
		pool:     p,
	}
}

// Start starts the Coordinator instance.
func (c *Coordinator) Start() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("!!! PANIC RECOVERED in handleDeactivate goroutine: %v", r)
			log.Printf("Stack trace:\n%s", debug.Stack())
		}
	}()

	activateEvent := c.broker.Subscribe(broker.Client, broker.ACTIVATE)
	deactivateEvent := c.broker.Subscribe(broker.Client, broker.DEACTIVATE)
	pushEvent := c.broker.Subscribe(broker.Client, broker.PUSH)
	pullEvent := c.broker.Subscribe(broker.Client, broker.PULL)
	mediaConnectedEvent := c.broker.Subscribe(broker.Media, broker.CONNECTED)
	mediaDisconnectedEvent := c.broker.Subscribe(broker.Media, broker.DISCONNECTED)
	peerFailedEvent := c.broker.Subscribe(broker.Peer, broker.FAILED)
	peerConnectedEvent := c.broker.Subscribe(broker.Peer, broker.CONNECTED)
	peerDisconnectedEvent := c.broker.Subscribe(broker.Peer, broker.DISCONNECTED)
	for {
		select {
		case event := <-activateEvent.Receive():
			go c.handleActivate(event)
		case event := <-deactivateEvent.Receive():
			go c.handleDeactivate(event)
		case event := <-pushEvent.Receive():
			go c.handlePush(event)
		case event := <-pullEvent.Receive():
			go c.handlePull(event)
		case event := <-mediaConnectedEvent.Receive():
			go c.handleMediaConnected(event)
		case event := <-mediaDisconnectedEvent.Receive():
			go c.handleMediaDisconnected(event)
		case event := <-peerFailedEvent.Receive():
			go c.handlePeerFailed(event)
		case event := <-peerConnectedEvent.Receive():
			go c.handlePeerConnected(event)
		case event := <-peerDisconnectedEvent.Receive():
			go c.handlePeerDisconnected(event)
		}
	}
}

// handleActivate handles the activate event. activate event means that a client
// requests to activate the connection.
func (c *Coordinator) handleActivate(event any) {
	msg, ok := event.(message.Activate)
	if !ok {
		log.Printf("error occurs in parsing activate message %v", event)
		return
	}

	if err := c.database.CreateClientInfo(msg.ChannelID, msg.ClientID); err != nil {
		log.Printf("error occurs in creating client info %v", err)
		return
	}
}

func (c *Coordinator) handleDeactivate(event any) {
	// --- 이 부분이 추가되어야 합니다. ---
	defer func() {
		if r := recover(); r != nil {
			log.Printf("!!! PANIC RECOVERED in handleDeactivate goroutine: %v", r)
			log.Printf("Stack trace:\n%s", debug.Stack())
		}
	}()
	// --- 추가된 부분 끝 ---

	// 01. Parse the event to message.Deactivate
	msg, ok := event.(message.Deactivate)
	if !ok {
		// activate message 대신 deactivate message로 로그 문구를 수정하는 것이 좋습니다.
		log.Printf("error occurs in parsing deactivate message %v", event)
		return
	}

	// 02. Find forwarding peer connections. Because the fetcher don't know the forwarder left or just temporal issue.
	// So we need to notify the fetcher that the forwarder left. And pull again.
	forwards, err := c.database.FindAllPeerConnectionInfoByFrom(msg.ChannelID, msg.ClientID)
	if err != nil {
		log.Printf("error occurs in finding connection info by from %v", err)
		// 에러 발생 시 여기에서 패닉이 발생할 수 있습니다.
		// 예를 들어, c.database가 nil인 경우 등.
		// 이 defer recover가 그런 패닉을 잡아줄 것입니다.
	}
	for _, forward := range forwards {
		if forward.IsConnected() {
			c.metric.DecrementPeerConnections()
		}
		log.Printf("publish closed")
		if err := c.broker.Publish(broker.ClientSocket, broker.Detail(forward.ChannelID+forward.To), response.Closed{
			Type:         response.CLOSED,
			ConnectionID: forward.ID,
		}); err != nil {
			log.Printf("error occurs in publishing close message %v", err)
		}
		if err := c.database.DeleteConnectionInfoByID(forward.ID); err != nil {
			log.Printf("error occurs in deleting connection info %v", err)
		}
	}

	// 03. Find fetching connections. Because the forwarder don't know the fetcher left or just temporal issue.
	// So we need to notify the forwarder that the fetcher left. Then Forwarder can clear the forwarding connection.
	fetches, err := c.database.FindAllPeerConnectionInfoByTo(msg.ChannelID, msg.ClientID)
	if err != nil {
		log.Printf("error occurs in finding connection info by to %v", err)
	}
	for _, fetch := range fetches {
		// 이 부분은 이미 위에서 forward 루프에서 처리되었을 수도 있는 fetch.ID를 다시 삭제 시도합니다.
		// 로직 상 의도된 것이 아니라면 중복 삭제 시도는 불필요하거나 오류가 발생할 수 있습니다.
		// 한 번 삭제된 ID를 다시 삭제 시도하는 것은 일반적으로 데이터베이스 에러는 아니지만,
		// 로직의 의미를 다시 확인해보시는 것이 좋습니다.
		if err := c.database.DeleteConnectionInfoByID(fetch.ID); err != nil {
			log.Printf("error occurs in deleting connection info %v", err)
		}
		switch fetch.Type {
		case database.PushToServer:
			if err := c.broker.Publish(broker.Media, broker.CLEAR, message.Clear{
				ConnectionID: fetch.ID,
			}); err != nil {
				log.Printf("error occurs in publishing close message %v", err)
			}
		case database.PullFromServer:
			if err := c.broker.Publish(broker.Media, broker.CLEAR, message.Clear{
				ConnectionID: fetch.ID,
			}); err != nil {
				log.Printf("error occurs in publishing close message %v", err)
			}
		case database.PeerToPeer:
			if err := c.broker.Publish(broker.ClientSocket, broker.Detail(fetch.ChannelID+fetch.From), response.Clear{
				Type:         response.CLEAR,
				ConnectionID: fetch.ID,
			}); err != nil {
				log.Printf("error occurs in publishing close message %v", err)
			}
			if fetch.IsConnected() {
				c.metric.DecrementPeerConnections()
				if err := c.pool.UpdateClientScore(fetch.From, fetch.ChannelID, c.config.MaxForwardingNumber); err != nil {
					log.Printf("error occurs in updating client score %v", err)
				}
			}
		default:
			// 여기에서 "unhandled default case" 패닉이 발생할 수 있습니다.
			// 이제 이 defer recover가 이 패닉을 잡아줄 것입니다.
		}

		// 이 부분도 위와 마찬가지로 중복 삭제 시도가 될 수 있습니다.
		// 의도를 확인해보세요.
		if err := c.database.DeleteConnectionInfoByID(fetch.ID); err != nil {
			log.Printf("error occurs in deleting connection info %v", err)
		}
	}

	if err := c.database.DeleteClientInfoByID(msg.ChannelID, msg.ClientID); err != nil {
		log.Printf("error occurs in deleting client info %v", err)
	}

	connInfo, err := c.database.FindUpstreamInfo(msg.ChannelID)
	if err != nil {
		log.Printf("error occurs in finding upstream info %v", err)
	}
	// connInfo가 nil일 가능성, 또는 connInfo.From이 msg.ClientID와 다를 경우를 고려해야 합니다.
	// connInfo가 nil일 경우, connInfo.From에 접근하면 nil dereference panic이 발생할 수 있습니다.
	// FindUpstreamInfo가 에러와 함께 nil connInfo를 반환할 수 있으므로,
	// err가 nil이더라도 connInfo가 유효한지 확인하는 것이 안전합니다.
	if connInfo != nil && connInfo.From == msg.ClientID { // nil 체크 추가
		if err := c.broker.Publish(broker.Media, broker.CLOSE, message.Close{
			ConnectionID: connInfo.ID,
		}); err != nil {
			log.Printf("error occurs in publishing close message %v", err)
		}
		if err := c.database.DeleteConnectionInfoByID(connInfo.ID); err != nil {
			log.Printf("error occurs in deleting connection info %v", err)
		}
		if err := c.database.DeleteChannelInfoByID(msg.ChannelID); err != nil {
			log.Printf("error occurs in deleting channel info %v", err)
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

// handleMediaConnected handles the connected event. This event is about Media server to client
func (c *Coordinator) handleMediaConnected(event any) {
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
	if err := c.balance(connInfo.ChannelID, connInfo.To); err != nil && !errors.Is(err, ErrNoForwarder) {
		log.Printf("error occurs in balancing %v", err)
		log.Printf("remain fetchfrom server")
		return
	}
}

// handleMediaDisconnected handles the disconnected event. This event is about Media server to client.
// Currently, When client disconnected, we got disconnected event from deactivation event in Signal server first.
// And Signal server will notify to Media server to close the connection. So we don't need to handle this event now.
// But if we consider that signal disconnected but media server still connected, we should think it again.
func (c *Coordinator) handleMediaDisconnected(_ any) {
	// 01. Parse the event to message.Disconnected
	//msg, ok := event.(message.Disconnected)
	//if !ok {
	//	log.Printf("error occurs in parsing disconnected message %v", event)
	//	return
	//}
}

// handlePeerFailed handles the failed event. This event is about client to client
func (c *Coordinator) handlePeerFailed(event any) {
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

	if err := c.balance(connInfo.ChannelID, connInfo.To); err != nil && !errors.Is(err, ErrNoForwarder) {
		log.Printf("error occurs in balancing %v", err)
		return
	}

	if err := c.balance(connInfo.ChannelID, connInfo.From); err != nil && !errors.Is(err, ErrNoForwarder) {
		log.Printf("error occurs in balancing %v", err)
		return
	}
}

// handlePeerConnected handles the succeed event. This event is about client to client
func (c *Coordinator) handlePeerConnected(event any) {
	msg, ok := event.(message.Connected)
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
	if err := c.database.DeleteConnectionInfoByID(serverConn.ID); err != nil {
		log.Printf("error occurs in deleting connection info %v", err)
		return
	}
	c.metric.IncrementPeerConnections()

	if err := c.broker.Publish(broker.Media, broker.CLEAR, message.Clear{
		ConnectionID: serverConn.ID,
	}); err != nil {
		log.Printf("error occurs in publishing closure message %v", err)
		return
	}
}

// handlePeerDisconnected handles the succeed event. This event is about client to client
func (c *Coordinator) handlePeerDisconnected(event any) {
	msg, ok := event.(message.Disconnected)
	if !ok {
		log.Printf("error occurs in parsing failed message %v", event)
		return
	}
	if err := c.database.DeleteConnectionInfoByID(msg.ConnectionID); err != nil {
		log.Printf("error occurs in updating connection info %v", err)
		return
	}
}

func (c *Coordinator) balance(channelID, fetcherID string) error {
	if !c.config.SetPeerConnection {
		return nil
	}
	log.Printf("balancing %s %s", channelID, fetcherID)

	fetcher, err := c.database.FindClientInfoByID(channelID, fetcherID)
	if err != nil {
		return fmt.Errorf("error finding client info: %v", err)
	}

	forwarderInfo := c.pool.GetTopForwarder(channelID)
	if forwarderInfo == nil {
		log.Printf("no forwarder found%v", forwarderInfo)

		if err := c.pool.AddClient(*fetcher); err != nil {
			return fmt.Errorf("error occurs in adding client info to forward %v", err)
		}
		log.Printf("added forward info to pool")
		return nil
	}
	log.Printf("found forwarder %v", forwarderInfo)

	peerConn, err := c.database.CreatePeerConnectionInfo(channelID, forwarderInfo.ID, fetcherID, shortuuid.New())
	if err != nil {
		return fmt.Errorf("error occurs in creating peer connection info %v", err)
	}

	c.metric.IncrementBalancingOccurs()
	if err := c.pool.UpdateClientScore(forwarderInfo.ID, channelID, c.config.MaxForwardingNumber); err != nil {
		return fmt.Errorf("error occurs in updating client score %v", err)
	}
	if err := c.broker.Publish(broker.ClientSocket, broker.Detail(channelID+fetcherID), response.Forward{
		Type:         response.FORWARD,
		ConnectionID: peerConn.ID,
	}); err != nil {
		return fmt.Errorf("error occurs in publishing fetch message %v", err)
	}
	return nil
}
