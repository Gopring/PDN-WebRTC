package coordinator

// Default values for the coordinator. If the values are not set, these values are used.
const (
	DefaultMaxForwardingNumber = 1
	DefaultSetPeerConnection   = false
)

// Config contains the configuration for the coordinator.
type Config struct {
	MaxForwardingNumber int
	SetPeerConnection   bool
}
