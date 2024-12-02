// Package media contains managing channels and connections using WebRTC.
package media

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	"log"
)

// Channel manages connections
type Channel struct {
	upstream *webrtc.TrackLocalStaticRTP
}

// NewChannel creates a new Channel instance.
func NewChannel() *Channel {
	return &Channel{}
}

// SetUpstream sets the upstream connection.
func (c *Channel) SetUpstream(conn *webrtc.PeerConnection, id string) {
	conn.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var newTrackErr error
		c.upstream, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", id)
		if newTrackErr != nil {
			// TODO(window9u): we should handle this panic properly.
			log.Println(newTrackErr)
		}

		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				// TODO(window9u): we should handle this panic properly.
				log.Println(newTrackErr)
				return
			}
			if _, err := c.upstream.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				log.Println(newTrackErr)
				return
			}
		}
	})
}

// SetDownstream sets the downstream connection.
func (c *Channel) SetDownstream(conn *webrtc.PeerConnection, _ string) error {
	if c.upstream == nil {
		return errors.New("upstream not exists")
	}

	rtpSender, err := conn.AddTrack(c.upstream)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	// ReadJSON RTCP packets
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			// ReadJSON RTCP packets
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	return nil
}
