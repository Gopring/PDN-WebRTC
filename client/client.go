// Package client contains pdn client files.
package client

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"io"
	"net/url"
	"pdn/types/api/request"
)

// message type
const (
	PULL    = "PULL"
	PUSH    = "PUSH"
	FORWARD = "FORWARD"
	ARRANGE = "ARRANGE"
	FETCH   = "FETCH"
)

// Client is user of pdn. this client will be used for test.
type Client struct {
	userID     string
	channelID  string
	serverURL  string
	localTrack *webrtc.TrackLocalStaticRTP
	socket     *websocket.Conn
}

// New creates a new WebSocket client.
func New(serverURL, userID, channelID string) (*Client, error) {
	return &Client{
		serverURL: serverURL,
		userID:    userID,
		channelID: channelID,
	}, nil
}

func (c *Client) dial() error {
	u := url.URL{Scheme: "ws", Host: c.serverURL, Path: "/ws"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	c.socket = wsConn
	return nil
}

// push sends a signal to the WebSocket server and receives a response.
func (c *Client) push(data any) (string, error) {
	// Push the data as JSON to the WebSocket server
	if err := c.socket.WriteJSON(data); err != nil {
		return "", fmt.Errorf("failed to PUSH data: %w", err)
	}

	// Read the response
	_, message, err := c.socket.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(message), nil
}

// activate activates the client on the server.
func (c *Client) activate() error {
	_, err := c.push(request.Activate{
		ChannelID: c.channelID,
		UserID:    c.userID,
	})
	if err != nil {
		return err
	}
	return nil
}

// Push sends a local track to the server.
func (c *Client) Push(localTrack *webrtc.TrackLocalStaticRTP) error {
	// 0. Dial the WebSocket server
	if err := c.dial(); err != nil {
		return err
	}

	// 1. Activate the client
	if err := c.activate(); err != nil {
		return err
	}

	// 2. Create a new peer connection
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	// 3. Create an SDP offer to start broadcasting
	offer, err := conn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		return err
	}

	// 4. Add track to PUSH
	sender, err := conn.AddTrack(localTrack)
	if err != nil {
		return err
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// 5. Push the offer (SDP) to the server to initiate broadcasting
	//    and PULL the server's SDP answer.
	remoteSDP, err := c.push(request.Signal{
		Type: PUSH,
		SDP:  offer.SDP,
	})
	if err != nil {
		return err
	}

	// 6. Set the server's SDP answer as the remote description for broadcasting
	if err = conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: remoteSDP}); err != nil {
		return err
	}

	return nil
}

// Pull receives a remote track from the server.
func (c *Client) Pull(consume func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) error {
	// 1. Create a new peer connection
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	// 2. Create an SDP offer to start broadcasting
	offer, err := conn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		return err
	}
	// 3. Add track to receive
	conn.OnTrack(consume)

	// 4. Push the offer (SDP) to the server to initiate broadcasting
	//    and PULL the server's SDP answer.
	signal := request.Signal{
		Type: PULL,
		SDP:  offer.SDP,
	}
	remoteSDP, err := c.push(signal)
	if err != nil {
		return err
	}
	// 5. Set the server's SDP answer as the remote description for broadcasting
	if err = conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: remoteSDP}); err != nil {
		return err
	}
	return nil
}

// Fetch fetches a remote track from the forwarder.
func (c *Client) Fetch(consume func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) error {
	//NOTE(window9u): FETCH logic is same as PULL. For clients, it is same
	// that PUSH their SDP and get SDP from server. But now, we detach for test and
	// make it more explicit.

	// 1. Create a new peer connection
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	// 2. Create an SDP offer to start broadcasting
	offer, err := conn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		return err
	}
	// 3. Add track to PUSH
	conn.OnTrack(consume)

	// 4. Push the offer (SDP) to the server to initiate broadcasting
	//    and PULL the server's SDP answer.
	signal := request.Signal{
		Type: FETCH,
		SDP:  offer.SDP,
	}
	remoteSDP, err := c.push(signal)
	if err != nil {
		return err
	}
	// 5. Set the server's SDP answer as the remote description for broadcasting
	if err = conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: remoteSDP}); err != nil {
		return err
	}
	return nil
}

// Forward forwards a remote track to the server. Forwarding logic is same as
// PULL. But it makes notice to server and also, it is more explicit for test.
func (c *Client) Forward() error {
	// 1. Create a new peer connection
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	// 2. Create an SDP offer to start broadcasting
	offer, err := conn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		return err
	}
	// 3. Add track to PUSH
	conn.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var newTrackErr error
		c.localTrack, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", c.userID)
		if newTrackErr != nil {
			panic(newTrackErr)
		}

		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}
			if _, err := c.localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}
	})

	// 4. Push the offer (SDP) to the server to initiate broadcasting
	//    and PULL the server's SDP answer.
	signal := request.Signal{
		Type: FORWARD,
		SDP:  offer.SDP,
	}
	remoteSDP, err := c.push(signal)
	if err != nil {
		return err
	}
	// 5. Set the server's SDP answer as the remote description for broadcasting
	if err = conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: remoteSDP}); err != nil {
		return err
	}
	return nil
}

// Arrange arranges a remote track to the server.
func (c *Client) Arrange(offer string) error {
	// server push ARRANGE to FORWARD to fetcher. client ARRANGE
	// their SDP and PUSH it to server. and server toss it to fetcher.
	// so it makes connections with fetcher and forwarder.

	// 1. Create a new peer connection
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	// 2. Create an SDP answer to start broadcasting
	if err := conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer, SDP: offer}); err != nil {
		return err
	}
	answer, err := conn.CreateAnswer(&webrtc.AnswerOptions{})
	if err != nil {
		return err
	}
	// 3. Add track to PUSH
	sender, err := conn.AddTrack(c.localTrack)
	if err != nil {
		return err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	sig := request.Signal{
		Type: ARRANGE,
		SDP:  answer.SDP,
	}

	// 6. Set sdp answer to the server for broadcasting
	if _, err := c.push(sig); err != nil {
		return err
	}
	return nil
}
