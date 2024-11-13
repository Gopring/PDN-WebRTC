// Package client contains pdn client files.
package client

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"io"
	"pdn/types/api/request"
)

// message type
const (
	RECEIVE = "RECEIVE"
	SEND    = "SEND"
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
	conn       *websocket.Conn
}

// New creates a new WebSocket client.
func New(serverURL, userID, channelID string) (*Client, error) {
	return &Client{
		serverURL: serverURL,
		userID:    userID,
		channelID: channelID,
	}, nil
}

func (c *Client) Dial() error {
	u := url.URL{Scheme: "ws", Host: c.serverURL, Path: "/ws"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to Dial: %w", err)
	}
	c.conn = wsConn
	return nil
}

// Disconnect gracefully closes the WebSocket connection.
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return fmt.Errorf("connection doesn't exist")
	}
	return c.conn.Close()
}

// signaling sends a signal to the WebSocket server and receives a response.
func (c *Client) signaling(signal request.Signal) (string, error) {
	// Send the signal as JSON to the WebSocket server
	if err := c.conn.WriteJSON(signal); err != nil {
		return "", fmt.Errorf("failed to SEND signal: %w", err)
	}

	// Read the response
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(message), nil
}

// Send sends a local track to the server.
func (c *Client) Send(localTrack *webrtc.TrackLocalStaticRTP) error {
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

	// 3. Add track to SEND
	sender, err := conn.AddTrack(localTrack)
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// 4. Send the offer (SDP) to the server to initiate broadcasting
	//    and RECEIVE the server's SDP answer.
	signal := request.Signal{
		Type:      SEND,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.signaling(signal)
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

func (c *Client) Receive(consume func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) error {
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
	// 3. Add track to SEND
	conn.OnTrack(consume)

	// 4. Send the offer (SDP) to the server to initiate broadcasting
	//    and RECEIVE the server's SDP answer.
	signal := request.Signal{
		Type:      RECEIVE,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.signaling(signal)
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

func (c *Client) Fetch(consume func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) error {
	//NOTE(window9u): FETCH logic is same as RECEIVE. For clients, it is same
	// that SEND their SDP and get SDP from server. But now, we detach for test and
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
	// 3. Add track to SEND
	conn.OnTrack(consume)

	// 4. Send the offer (SDP) to the server to initiate broadcasting
	//    and RECEIVE the server's SDP answer.
	signal := request.Signal{
		Type:      FETCH,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.signaling(signal)
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
	// 3. Add track to SEND
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

	// 4. Send the offer (SDP) to the server to initiate broadcasting
	//    and RECEIVE the server's SDP answer.
	signal := request.Signal{
		Type:      FORWARD,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.signaling(signal)
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

func (c *Client) Arrange(offer string) error {
	// server request ARRANGE to FORWARD to fetcher. client ARRANGE
	// their SDP and SEND it to server. and server toss it to fetcher.
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
	// 3. Add track to SEND
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
		Type:      ARRANGE,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       answer.SDP,
	}

	// 6. Set sdp answer to the server for broadcasting
	if _, err := c.signaling(sig); err != nil {
		return err
	}
	return nil
}
