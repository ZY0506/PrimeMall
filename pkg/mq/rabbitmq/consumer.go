package rabbitmq

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxRetries = 3

// Subscribe 实现 Consumer 接口。
//
// 每个订阅独占一个 channel：amqp091 的 Channel 并非并发安全，
// 多个消费者共用 channel 做 ACK/NACK、又与生产者 Publish 并发写同一个 channel，会导致协议错乱。
// 本方法返回只代表「订阅注册完成」，实际消费在后台协程中执行。
func (c *Client) Subscribe(ctx context.Context, name string, handler func([]byte) error) error {
	logger := logx.WithContext(ctx)

	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("为队列 %s 打开消费 channel 失败: %w", name, err)
	}

	c.consumerMu.Lock()
	c.consumerChs = append(c.consumerChs, ch)
	c.consumerMu.Unlock()

	// 声明队列（同生产者保持一致）
	if _, err = ch.QueueDeclare(
		name, true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("声明队列 %s 失败: %w", name, err)
	}

	// 消费QoS：每次预取1条，避免积压
	if err = ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("设置队列 %s 的QoS失败: %w", name, err)
	}

	msgs, err := ch.Consume(
		name,  // queue
		"",    // consumer
		false, // auto-ack（手动ACK）
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("订阅队列 %s 失败: %w", name, err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("消息消费协程panic: queue=%s, err=%v", name, r)
			}
		}()

		var mu sync.Mutex
		retryCount := make(map[string]int) // bodyHash -> retryCount

		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					// channel 被关闭（broker 断连或主动 Close），消费循环退出
					logger.Errorf("队列 %s 的消费 channel 已关闭，停止消费", name)
					return
				}

				// 基于消息体 hash 跟踪重试次数
				key := fmt.Sprintf("%x", md5.Sum(d.Body))
				mu.Lock()
				retryCount[key]++
				attempts := retryCount[key]
				mu.Unlock()

				if attempts > maxRetries {
					logger.Errorf("消息重试已达上限(%d次)，丢弃: queue=%s, body=%s", maxRetries, name, d.Body)
					if ackErr := d.Ack(false); ackErr != nil {
						logger.Errorf("丢弃消息时ACK失败: queue=%s, err=%v", name, ackErr)
					}
					mu.Lock()
					delete(retryCount, key)
					mu.Unlock()
					continue
				}

				// 调用业务处理函数
				if handlerErr := handler(d.Body); handlerErr != nil {
					logger.Errorf("处理消息失败(第%d次), queue=%s, err=%v, body=%s", attempts, name, handlerErr, d.Body)
					if nackErr := d.Nack(false, true); nackErr != nil { // requeue
						logger.Errorf("Nack(requeue)失败: queue=%s, err=%v", name, nackErr)
					}
					continue
				}

				// 处理成功：手动ACK
				if ackErr := d.Ack(false); ackErr != nil {
					logger.Errorf("ACK失败: queue=%s, err=%v", name, ackErr)
				}
				mu.Lock()
				delete(retryCount, key)
				mu.Unlock()
			}
		}
	}()

	return nil
}
