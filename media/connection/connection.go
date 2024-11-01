// Package connection contains managing connections using WebRTC.
package connection

import (
	"fmt"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

// Connection contains the WebRTC connection.
type Connection struct {
	conn *webrtc.PeerConnection
	sdp  string
}

// NewInbound creates a new inbound connection.
func NewInbound(config webrtc.Configuration, sdp string) (*Connection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("failed to register default codecs: %w", err)
	}

	// This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, fmt.Errorf("failed to register default interceptors: %w", err)
	}

	// This interceptor sends a PLI every 3 seconds. A PLI causes a video keyframe to be generated by the sender.
	// This makes our video seekable and more error resilent, but at a cost of lower picture quality and higher bitrates
	// A real world application should process incoming RTCP packets from viewers and forward them to senders
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return nil, fmt.Errorf("failed to create interval pli factory: %w", err)
	}
	i.Add(intervalPliFactory)

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i)).NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	return &Connection{conn: peerConnection, sdp: sdp}, nil
}

// NewOutbound creates a new outbound connection.
func NewOutbound(config webrtc.Configuration, sdp string) (*Connection, error) {
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	return &Connection{conn: peerConnection, sdp: sdp}, nil
}

// StartICE starts ICE.
func (c *Connection) StartICE() error {
	var err error
	broadOffer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: c.sdp}
	if err = c.conn.SetRemoteDescription(broadOffer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	answer, err := c.conn.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	gatherComplete := webrtc.GatheringCompletePromise(c.conn)

	err = c.conn.SetLocalDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}
	<-gatherComplete
	return nil
}

// ServerSDP returns the server SDP.
func (c *Connection) ServerSDP() string {
	return c.conn.LocalDescription().SDP
}

// Ontrack sets the on track handler. This is used for receiving tracks.
func (c *Connection) Ontrack(f func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	c.conn.OnTrack(f)
}

// Addtrack adds a track. This is used for sending tracks.
func (c *Connection) Addtrack(stream *webrtc.TrackLocalStaticRTP) (*webrtc.RTPSender, error) {
	return c.conn.AddTrack(stream)
}
