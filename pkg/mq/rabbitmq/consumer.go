package rabbitmq

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxRetries = 3

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

	var mu sync.Mutex
	retryCount := make(map[string]int) // bodyHash -> retryCount

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

				// 基于消息体 hash 跟踪重试次数
				key := fmt.Sprintf("%x", md5.Sum(d.Body))
				mu.Lock()
				retryCount[key]++
				attempts := retryCount[key]
				mu.Unlock()

				if attempts > maxRetries {
					logx.WithContext(ctx).Errorf("消息重试已达上限(%d次)，丢弃: body=%s", maxRetries, d.Body)
					_ = d.Ack(false)
					mu.Lock()
					delete(retryCount, key)
					mu.Unlock()
					continue
				}

				// 调用业务处理函数
				if err = handler(d.Body); err != nil {
					logx.WithContext(ctx).Errorf("处理消息失败(第%d次), err=%v, body=%s", attempts, err, d.Body)
					_ = d.Nack(false, true) // requeue
					continue
				}
				// 处理成功：手动ACK
				_ = d.Ack(false)
				mu.Lock()
				delete(retryCount, key)
				mu.Unlock()
			}
		}
	}()
	return nil
}
