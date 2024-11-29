package broker

const (
	CLIENT = iota
	PULL
	PUSH
)

const (
	CONTROLLER  = "CLIENT"
	COORDINATOR = "COORDINATOR"
	CLASSIFIER  = "CLASSIFIER"
)

type TOPIC int
type FROM string

type Broker struct {
}

func New() *Broker {
	return &Broker{}
}

func (b *Broker) Publish(topic TOPIC, from FROM, message interface{}) error {
	return nil
}

func (b *Broker) Send(topic TOPIC, to string, message interface{}) error {
	return nil
}

func (b *Broker) Register(topic TOPIC, from FROM) error {
	return nil
}

func (b *Broker) Unregister(topic TOPIC, from FROM) error {
	return nil
}

func (b *Broker) Subscribe(topic TOPIC, from FROM) (<-chan any, error) {
	return nil, nil
}

func (b *Broker) Unsubscribe(topic TOPIC, from FROM, ch <-chan []byte) error {
	return nil
}
