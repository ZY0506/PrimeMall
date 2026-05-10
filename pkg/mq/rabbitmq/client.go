package rabbitmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"sync"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	mu      sync.Mutex
	url     string // 重连的时候会用到
}

// NewClient 创建 RabbitMQ 客户端
func NewClient(url string) (*Client, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("打开 channel 失败: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close 优雅关闭连接
func (c *Client) Close() error {
	var err error
	if c.channel != nil {
		err = c.channel.Close()
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}
