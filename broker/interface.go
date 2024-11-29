package broker

const (
	AUTH = iota
	CLIENT
	PULL
	PUSH
	MEDIA
	CLASSIFIER
	COORDINATOR
)

type TOPIC int
type DETAIL string

// Broker 인터페이스
type Broker interface {
	Publish(topic TOPIC, message any) error
	SendAndWait(topic TOPIC, detail DETAIL, message []byte) ([]byte, error)
	Register(topic TOPIC, detail DETAIL) error
	Unregister(topic TOPIC, detail DETAIL) error
	Subscribe(topic TOPIC, detail DETAIL) (<-chan []byte, error)
	Unsubscribe(topic TOPIC, ch <-chan []byte) error
}
