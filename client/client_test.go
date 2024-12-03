package client

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/assert"
	"log"
	"pdn/metric"
	"pdn/signal"
	"testing"
)

// NewTestConfig creates a new signal.Config for testing.
func NewTestSignalConfig() signal.Config {
	return signal.Config{
		Port:     8080,
		CertFile: "",
		KeyFile:  "",
		Debug:    true,
	}
}

// NewTestMetricConfig creates a new metric.Config for testing.
func NewTestMetricConfig() metric.Config {
	return metric.Config{
		Port: 9090,
		Path: "/metrics",
	}
}

// StartTestSignal starts a signal server for testing.
func StartTestSignal() {
	signalConfig := NewTestSignalConfig()
	metricConfig := NewTestMetricConfig()
	s := signal.New(signalConfig, metricConfig)

	_ = s.Start()
}

// TestBroadcast tests basic workflow of broadcast and view.
func TestBroadcast(t *testing.T) {
	// I skipped this test because this test logic is not implemented yet.
	// for details, client in this test code doesn't send ice-ufrag and ice-pwd.
	// this is because the client sends sdp before get all ice candidates.
	t.Skipf("Skip this test because it is not implemented yet")
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
