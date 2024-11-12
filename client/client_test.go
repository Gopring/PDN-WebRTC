package client

import (
	"pdn/signal"
	"testing"
)

func NewTestConfig() signal.Config {
	return signal.Config{
		Port:     8080,
		CertFile: "",
		KeyFile:  "",
		Debug:    false,
	}
}

func StartTestSignal() {
	s := signal.New(NewTestConfig())
	_ = s.Start()
}

func TestBroadcast(t *testing.T) {
	go StartTestSignal()

}
