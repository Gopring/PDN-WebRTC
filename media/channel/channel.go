// Package channel manages the connection between the media server and the client.
package channel

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	"pdn/media/connection"
)

// Channel manages connections
type Channel struct {
	// TODO(window9u): we should add locker for connections.
	connections map[string]*connection.Connection
	forwarders  []string
	upstream    *webrtc.TrackLocalStaticRTP
}

// New creates a new Channel instance.
func New() *Channel {
	return &Channel{
		connections: map[string]*connection.Connection{},
	}
}

// SetUpstream sets the upstream connection.
func (c *Channel) SetUpstream(conn *connection.Connection, id string) {
	c.connections[id] = conn
	conn.Ontrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var newTrackErr error
		c.upstream, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", id)
		if newTrackErr != nil {
			// TODO(window9u): we should handle this panic properly.
			panic(newTrackErr)
		}

		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				// TODO(window9u): we should handle this panic properly.
				panic(readErr)
			}
			if _, err := c.upstream.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}
	})
}

// SetDownstream sets the downstream connection.
func (c *Channel) SetDownstream(conn *connection.Connection, id string) error {
	if c.upstream == nil {
		return errors.New("upstream not exists")
	}

	c.connections[id] = conn
	rtpSender, err := conn.Addtrack(c.upstream)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	// Read RTCP packets
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			// Read RTCP packets
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	return nil
}

// AddForwarder adds the specified user ID to the list of forwarders for the channel.
func (c *Channel) AddForwarder(id string) {
	c.forwarders = append(c.forwarders, id)
}

// GetForwarder retrieves the first forwarder ID from the list of forwarders.
func (c *Channel) GetForwarder() string {
	return c.forwarders[0]
}
