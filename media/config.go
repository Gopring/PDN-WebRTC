// Package media contains managing streams and connections using WebRTC.
package media

import (
	"fmt"
	"github.com/pion/webrtc/v4"
	"strconv"
)

// Config defines the configuration for the media server.
type Config struct {
	IP         string // ip for media server.
	MinUdpPort string // Minimum UDP port for WebRTC
	MaxUdpPort string // Maximum UDP port for WebRTC
}

// SetPortRange sets the ephemeral UDP port range for WebRTC.
func (med *Config) SetPortRange(s *webrtc.SettingEngine) error {
	minPort, err := strconv.Atoi(med.MinUdpPort)
	if err != nil || minPort < 0 || minPort > 65535 {
		return fmt.Errorf("invalid MinUdpPort: %s, error: %v", med.MinUdpPort, err)
	}

	maxPort, err := strconv.Atoi(med.MaxUdpPort)
	if err != nil || maxPort < 0 || maxPort > 65535 {
		return fmt.Errorf("invalid MaxUdpPort: %s, error: %v", med.MaxUdpPort, err)
	}

	// Check if the range is valid
	if minPort > maxPort {
		return fmt.Errorf("invalid port range: MinUdpPort (%d) > MaxUdpPort (%d)", minPort, maxPort)
	}

	// Apply the port range to the setting engine
	err = s.SetEphemeralUDPPortRange(uint16(minPort), uint16(maxPort))
	if err != nil {
		return fmt.Errorf("failed to set ephemeral UDP port range: %w", err)
	}

	return nil
}
