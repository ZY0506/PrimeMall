package rabbitmq

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// publishConfirmTimeout 等待 broker 发布确认的超时时间
const publishConfirmTimeout = 5 * time.Second

// Client RabbitMQ 客户端。
//
// channel 模型：连接共享，channel 隔离。
// amqp091 的 Channel 明确不是并发安全的，而本服务里同时存在三类写操作：
// 业务协程投递消息、订单创建消费者 ACK/NACK、订单超时消费者 ACK/NACK。
// 因此生产者独占一个 channel，每个 Subscribe 也各自独占一个 channel，
// 避免多个 goroutine 并发写同一个 channel 导致协议错乱。
type Client struct {
	conn *amqp091.Connection
	url  string // 重连的时候会用到

	producerMu sync.Mutex       // 保护 producerCh，同时串行化「发布 → 等确认」
	producerCh *amqp091.Channel // 专用于发布，已开启 publisher confirm
	confirms   chan amqp091.Confirmation

	consumerMu  sync.Mutex         // 保护 consumerChs
	consumerChs []*amqp091.Channel // 每个 Subscribe 独占一个 channel
}

// NewClient 创建 RabbitMQ 客户端
func NewClient(url string) (*Client, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	c := &Client{
		conn: conn,
		url:  url,
	}
	if err = c.openProducerChannel(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return c, nil
}

// openProducerChannel 打开专用发布 channel 并开启 publisher confirm。
// 开启后 broker 会对每条消息回执 ACK/NACK，Publish 只有拿到 ACK 才算真正投递成功，
// 否则「Publish 返回 nil」只代表消息写进了本地 TCP 缓冲，不代表 broker 收到了。
func (c *Client) openProducerChannel() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("打开生产者 channel 失败: %w", err)
	}
	if err = ch.Confirm(false); err != nil {
		_ = ch.Close()
		return fmt.Errorf("开启 publisher confirm 失败: %w", err)
	}

	c.producerCh = ch
	c.confirms = ch.NotifyPublish(make(chan amqp091.Confirmation, 1))
	return nil
}

// producerChannel 返回生产者 channel（调用方需自行持有 producerMu）
func (c *Client) producerChannelLocked() (*amqp091.Channel, error) {
	if c.producerCh == nil {
		return nil, errors.New("生产者 channel 不可用（连接可能已关闭）")
	}
	return c.producerCh, nil
}

// Close 优雅关闭：先关所有 channel，再关连接
func (c *Client) Close() error {
	c.producerMu.Lock()
	if c.producerCh != nil {
		_ = c.producerCh.Close()
		c.producerCh = nil
	}
	c.producerMu.Unlock()

	c.consumerMu.Lock()
	for _, ch := range c.consumerChs {
		_ = ch.Close()
	}
	c.consumerChs = nil
	c.consumerMu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
