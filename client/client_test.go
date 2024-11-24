package client

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/assert"
	"log"
	"pdn/signal"
	"testing"
)

// NewTestConfig creates a new signal.Config for testing.
func NewTestConfig() signal.Config {
	return signal.Config{
		Port:     8080,
		CertFile: "",
		KeyFile:  "",
		Debug:    true,
	}
}

// StartTestSignal starts a signal server for testing.
func StartTestSignal() {
	s := signal.New(NewTestConfig())
	_ = s.Start()
}

// TestBroadcast tests basic workflow of broadcast and view.
func TestBroadcast(t *testing.T) {
	go StartTestSignal()
	broadcaster, err := New("localhost:8080", "test", "test")
	assert.NoError(t, err)
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video", "test")
	assert.NoError(t, err)
	log.Println("broadcasting")
	assert.NoError(t, broadcaster.Push(track))

	receiver, err := New("localhost:8080", "test", "test")
	assert.NoError(t, err)
	consumerTrack := func(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remote.Read(rtpBuf)
			if readErr != nil {
				// TODO(window9u): we should handle this panic properly.
				panic(readErr)
			}
			packet := &rtp.Packet{}
			if err := packet.Unmarshal(rtpBuf[:i]); err != nil {
				panic(err)
			}
			assert.Equal(t, uint8(96), packet.PayloadType)
			assert.Equal(t, uint16(1), packet.SequenceNumber)
			assert.Equal(t, []byte{0x00, 0x02}, packet.Payload)
		}
	}
	log.Println("receiving")
	assert.NoError(t, receiver.Pull(consumerTrack))
	assert.NoError(t, track.WriteRTP(&rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			SequenceNumber: 1,
			PayloadType:    96,
			Padding:        true,
		},
		Payload: []byte{0x00, 0x02},
	}))

}
