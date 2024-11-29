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

type Broker struct {
}

func New() *Broker {
	return &Broker{}
}

func (b *Broker) Publish(topic TOPIC, message interface{}) error {
	return nil
}

func (b *Broker) SendAndWait(topic TOPIC, detail DETAIL, message []byte) ([]byte, error) {
	return nil, nil
}

func (b *Broker) Register(topic TOPIC, detail DETAIL) error {
	return nil
}

func (b *Broker) Unregister(topic TOPIC, detail DETAIL) error {
	return nil
}

func (b *Broker) Subscribe(topic TOPIC, detail DETAIL) (<-chan []byte, error) {
	return nil, nil
}

func (b *Broker) Unsubscribe(topic TOPIC, ch <-chan []byte) error {
	return nil
}
