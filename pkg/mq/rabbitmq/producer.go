package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
)

// InitDelayQueue 初始化延迟队列
func (c *Client) InitDelayQueue(delayQueueName, dlxExchange, dlxRoutingKey string, ttlMs int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. 声明死信交换机
	err := c.channel.ExchangeDeclare(
		dlxExchange, // 交换器名称
		"direct",    // 类型
		true,        // durable
		false,       // autoDelete
		false,       // internal
		false,       // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明死信交换机失败: %w", err)
	}

	// 2. 声明延迟队列
	args := amqp091.Table{
		"x-dead-letter-exchange":    dlxExchange,   // 死信交换机
		"x-dead-letter-routing-key": dlxRoutingKey, // 死信路由键
		"x-message-ttl":             int64(ttlMs),  // 队列全局 TTL（毫秒）
	}
	_, err = c.channel.QueueDeclare(
		delayQueueName, // 延迟队列名
		true,           // 持久化
		false,          // 自动删除
		false,          // exclusive
		false,          // noWait
		args,
	)
	if err != nil {
		return fmt.Errorf("声明延迟队列失败: %w", err)
	}

	// 4. 声明业务超时处理队列（消费者实际监听的队列）
	_, err = c.channel.QueueDeclare(
		dlxRoutingKey,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明超时处理队列失败: %w", err)
	}

	// 5. 绑定超时处理队列到死信交换机
	err = c.channel.QueueBind(
		dlxRoutingKey, // queue name
		dlxRoutingKey, // routing key
		dlxExchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("绑定超时队列失败: %w", err)
	}

	return nil
}

// Publish 实现 Producer 接口
func (c *Client) Publish(ctx context.Context, exchange, key string, msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.channel.PublishWithContext(ctx,
		exchange,
		key,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         msg,
			DeliveryMode: amqp091.Persistent, // 持久化消息
		},
	)
	return err
}
