package client

import (
	"github.com/stretchr/testify/assert"
	"pdn/signal"
	"testing"
)

func NewTestConfig() signal.Config {
	return signal.Config{
		Port:     8080,
		CertFile: "",
		KeyFile:  "",
		Debug:    true,
	}
}

func StartTestSignal() {
	s := signal.New(NewTestConfig())
	_ = s.Start()
}

func TestBroadcast(t *testing.T) {
	go StartTestSignal()
	client, err := New("localhost:8080", "test", "test")
	assert.NoError(t, err)
	assert.NoError(t, client.Dial())
	defer assert.NoError(t, client.Disconnect())
}
