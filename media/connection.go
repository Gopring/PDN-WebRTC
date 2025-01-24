// Package media contains managing streams and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
	"log"
)

// NewInboundConnection creates a new inbound connection.
func (med *Media) NewInboundConnection(config webrtc.Configuration) (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("failed to register default codecs: %w", err)
	}
	s := webrtc.SettingEngine{}

	// note: see https://stackoverflow.com/questions/68959096/pion-custom-sfu-server-not-working-inside-docker
	s.SetNAT1To1IPs([]string{med.config.IP}, webrtc.ICECandidateTypeHost)
	log.Printf("IP : %s", med.config.IP)

	err := med.config.SetPortRange(&s)
	if err != nil {
		log.Fatalf("Error setting port range: %v", err)
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
	peerConnection, err := webrtc.NewAPI(
		webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i),
		webrtc.WithSettingEngine(s),
	).NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	return peerConnection, nil
}

// NewOutboundConnection creates a new outbound connection.
func (med *Media) NewOutboundConnection(config webrtc.Configuration) (*webrtc.PeerConnection, error) {
	s := webrtc.SettingEngine{}
	s.SetNAT1To1IPs([]string{med.config.IP}, webrtc.ICECandidateTypeHost)
	log.Printf("IP : %s", med.config.IP)
	err := med.config.SetPortRange(&s)
	if err != nil {
		log.Fatalf("Error setting port range: %v", err)
	}

	peerConnection, err := webrtc.NewAPI(webrtc.WithSettingEngine(s)).NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}
	return peerConnection, nil
}

// StartICE starts ICE.
func StartICE(conn *webrtc.PeerConnection, sdp string) error {
	var err error
	broadOffer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}
	if err = conn.SetRemoteDescription(broadOffer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	answer, err := conn.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	gatherComplete := webrtc.GatheringCompletePromise(conn)

	err = conn.SetLocalDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}
	<-gatherComplete
	return nil
}
