package coordinator

const DefaultMaxForwardingNumber = 1
const DefaultSetPeerConnection = true

type Config struct {
	MaxForwardingNumber int
	SetPeerConnection   bool
}
