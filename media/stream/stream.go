// Package stream contains managing connections using WebRTC
package stream

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	"log"
)

// Stream manages connections
type Stream struct {
	Track *webrtc.TrackLocalStaticRTP
}

// New creates a new Stream instance.
func New() *Stream {
	return &Stream{}
}

// SetUpstream sets the Track connection.
func (s *Stream) SetUpstream(conn *webrtc.PeerConnection, id string) {
	conn.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var newTrackErr error
		s.Track, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", id)
		if newTrackErr != nil {
			// TODO(window9u): we should handle this panic properly.
			log.Println(newTrackErr)
		}

		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				log.Println(newTrackErr)
				return
			}
			if _, err := s.Track.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				log.Println(newTrackErr)
				return
			}
		}
	})
}

// SetDownstream sets the downstream connection.
func (s *Stream) SetDownstream(conn *webrtc.PeerConnection) error {
	if s.Track == nil {
		return errors.New("track not exists")
	}

	rtpSender, err := conn.AddTrack(s.Track)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	// ReadJSON RTCP packets
	// TODO(window9u): we should control this goroutine properly.
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
