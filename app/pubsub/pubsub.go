package pubsub

import "context"

type Message interface {
	Topic() string
	Value() []byte
}

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic ...string) MessageReader
}

type MessageReader interface {
	// Next returns the next message from the Reader.
	// if pre message exists and not commit,this will commit it before read next message
	Next() (Message, error)
}
