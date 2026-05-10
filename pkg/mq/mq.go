package mq

import "context"

type Producer interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Close() error
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func([]byte) error) error
	Close() error
}

type Client interface {
	Producer
	Consumer
	Close() error
}
