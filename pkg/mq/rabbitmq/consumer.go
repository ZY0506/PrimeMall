package rabbitmq

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

// Subscribe 实现 Consumer 接口
func (c *Client) Subscribe(ctx context.Context, name string, handler func([]byte) error) error {
	// 声明队列（同生产者保持一致）
	_, err := c.channel.QueueDeclare(
		name, true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	// 消费QoS：每次预取1条，避免积压
	_ = c.channel.Qos(1, 0, false)

	msgs, err := c.channel.Consume(
		name,  // queue
		"",    // consumer
		false, // auto-ack（手动ACK）
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.WithContext(ctx).Errorf("消息消费协程panic: %v", r)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				// 调用业务处理函数
				if err = handler(d.Body); err != nil {
					logx.WithContext(ctx).Errorf("处理消息失败, err=%v, body=%s", err, d.Body)
					_ = d.Nack(false, true)
					continue
				}
				// 处理成功：手动ACK
				_ = d.Ack(false)
			}
		}
	}()
	return nil
}
