// Package client contains pdn client files.
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"io"
	"net/http"
	"pdn/types/api/request"
)

// Client is user of pdn. this client will be used for test.
type Client struct {
	Type       string
	userID     string
	localTrack *webrtc.TrackLocalStaticRTP
	channelID  string
	conn       *websocket.Conn
}

// NewClient creates a new WebSocket client.
func NewClient(serverURL, clientType, userID, channelID string) (*Client, error) {
	wsConn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	return &Client{
		Type:      clientType,
		userID:    userID,
		channelID: channelID,
		conn:      wsConn,
	}, nil
}

// closeConnection gracefully closes the WebSocket connection.
func (c *Client) closeConnection() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// SendSignal sends a signal to the WebSocket server and receives a response.
func (c *Client) SendSignal(signal request.Signal) (string, error) {
	// Send the signal as JSON to the WebSocket server
	if err := c.conn.WriteJSON(signal); err != nil {
		return "", fmt.Errorf("failed to send signal: %w", err)
	}

	// Read the response
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(message), nil
}

func (c *Client) Receive(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver), serverURL string) error {
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
	// 3. Add track to send
	conn.OnTrack(f)

	// 4. Send the offer (SDP) to the server to initiate broadcasting
	//    and receive the server's SDP answer.
	signal := request.Signal{
		Type:      c.Type,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.SendSignal(signal)
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

func (c *Client) Fetch(localTrack *webrtc.TrackLocalStaticRTP, serverURL string) error {
	//NOTE(window9u): fetch logic is same as receive. For clients, it is same
	// that send their SDP and get SDP from server. But now, we detach for test and
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
	// 3. Add track to send
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
	//    and receive the server's SDP answer.
	signal := request.Signal{
		Type:      c.Type,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.SendSignal(signal)
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

func (c *Client) Forward(serverURL string) error {
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
	// 3. Add track to send
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
	//    and receive the server's SDP answer.
	signal := request.Signal{
		Type:      c.Type,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       offer.SDP,
	}
	remoteSDP, err := c.SendSignal(signal)
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

func (c *Client) Arrange(offer string, serverURL string) error {
	// server request arrange to forward to fetcher. client arrange
	// their SDP and send it to server. and server toss it to fetcher.
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
	// 3. Add track to send
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
	data, err := json.Marshal(request.Signal{
		Type:      c.Type,
		ChannelID: c.channelID,
		UserID:    c.userID,
		SDP:       answer.SDP,
	})
	// 6. Set sdp answer to the server for broadcasting
	res, err := http.Post(serverURL, "application/json", bytes.NewReader(data))
	if res.StatusCode >= 400 {
		return fmt.Errorf("fails")
	}

	return nil
}
