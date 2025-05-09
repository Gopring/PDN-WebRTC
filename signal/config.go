// Package signal contains business logic and prerequisites for WebRTC streaming.
package signal

import (
	"errors"
	"fmt"
	"os"
)

const (
	// DefaultPort is the default port number for the server.
	DefaultPort = 7070
)

// Below is the Error message for the server.
var (
	ErrInvalidPort     = errors.New("invalid port")
	ErrInvalidCertFile = errors.New("invalid cert file")
	ErrInvalidKeyFile  = errors.New("invalid key file")
)

// Config is the configuration for creating a Server instance.
type Config struct {
	Port     int
	Debug    bool
	CertFile string
	KeyFile  string
}

// IsSame checks if the given config is the same as the current one.
func (c Config) IsSame(config Config) bool {
	return c.Port == config.Port && c.CertFile == config.CertFile && c.KeyFile == config.KeyFile
}

// Validate validates the port number and the files for certification.
func (c Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("must be between 1 and 65535, given %d: %w", c.Port, ErrInvalidPort)
	}

	if c.CertFile == "" && c.KeyFile == "" {
		return nil
	}

	if _, err := os.Stat(c.CertFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %w", c.CertFile, ErrInvalidCertFile)
		}
		return fmt.Errorf("unable to access %s: %w", c.CertFile, ErrInvalidCertFile)
	}

	if _, err := os.Stat(c.KeyFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %w", c.CertFile, ErrInvalidKeyFile)
		}
		return fmt.Errorf("unable to access %s: %w", c.CertFile, ErrInvalidKeyFile)
	}

	return nil
}
