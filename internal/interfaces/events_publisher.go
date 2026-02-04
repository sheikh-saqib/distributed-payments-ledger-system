package interfaces

type EventPublisher interface {
	Publish(topic string, event any) error
}
