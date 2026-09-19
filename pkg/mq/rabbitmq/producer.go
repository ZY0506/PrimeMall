package rabbitmq

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// InitDelayQueue 初始化延迟队列（基于 DLX + 消息级 TTL）。
//
// 链路：投递到 <delayQueueName>_exchange → 进入 delayQueueName 队列等待 TTL 到期
// → 变成死信路由到 dlxExchange(routing key = dlxRoutingKey) → 进入 dlxRoutingKey 队列 → 被消费者处理。
//
// 注意：TTL 是「消息级」的（发布时写在消息的 Expiration 字段上），
// 队列本身不设 x-message-ttl。队列级 TTL 会让队尾消息被前面的长 TTL 消息堵住（队头阻塞），
// 一旦将来出现不同超时时长的消息，实际延迟会严重偏大。
func (c *Client) InitDelayQueue(delayQueueName, dlxExchange, dlxRoutingKey string) error {
	c.producerMu.Lock()
	defer c.producerMu.Unlock()

	ch, err := c.producerChannelLocked()
	if err != nil {
		return err
	}

	// 1. 声明死信交换机
	if err = ch.ExchangeDeclare(
		dlxExchange, // 交换器名称
		"direct",    // 类型
		true,        // durable
		false,       // autoDelete
		false,       // internal
		false,       // noWait
		nil,
	); err != nil {
		return fmt.Errorf("声明死信交换机失败: %w", err)
	}

	// 2. 声明延迟队列（仅做死信路由，不带 x-message-ttl）
	if _, err = ch.QueueDeclare(
		delayQueueName, // 延迟队列名
		true,           // 持久化
		false,          // 自动删除
		false,          // exclusive
		false,          // noWait
		amqp091.Table{
			"x-dead-letter-exchange":    dlxExchange,   // 死信交换机
			"x-dead-letter-routing-key": dlxRoutingKey, // 死信路由键
		},
	); err != nil {
		// 队列已存在但参数不一致时，broker 会直接关掉 channel，这里补一句可执行的运维提示
		return fmt.Errorf("声明延迟队列失败（队列 %s 参数不可变，需先删除旧队列或用新队列名）: %w",
			delayQueueName, err)
	}

	// 3. 声明延迟消息发布交换机，绑定延迟队列
	delayExchange := delayQueueName + "_exchange"
	if err = ch.ExchangeDeclare(
		delayExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("声明延迟交换机失败: %w", err)
	}
	if err = ch.QueueBind(
		delayQueueName,
		delayQueueName,
		delayExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("绑定延迟队列失败: %w", err)
	}

	// 4. 声明业务超时处理队列（消费者实际监听的队列）
	if _, err = ch.QueueDeclare(
		dlxRoutingKey,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("声明超时处理队列失败: %w", err)
	}

	// 5. 绑定超时处理队列到死信交换机
	if err = ch.QueueBind(
		dlxRoutingKey, // queue name
		dlxRoutingKey, // routing key
		dlxExchange,   // exchange
		false,
		nil,
	); err != nil {
		return fmt.Errorf("绑定超时队列失败: %w", err)
	}

	return nil
}

// Publish 实现 Producer 接口：发布普通消息，并等待 broker 确认
func (c *Client) Publish(ctx context.Context, exchange, key string, msg []byte) error {
	return c.publish(ctx, exchange, key, msg, 0)
}

// PublishDelay 发布延迟消息：TTL 挂在消息自身，到期后作为死信进入 dlxRoutingKey 队列。
// ttl <= 0 时等同于普通发布。
func (c *Client) PublishDelay(ctx context.Context, exchange, key string, msg []byte, ttl time.Duration) error {
	return c.publish(ctx, exchange, key, msg, ttl)
}

func (c *Client) publish(ctx context.Context, exchange, key string, msg []byte, ttl time.Duration) error {
	c.producerMu.Lock()
	defer c.producerMu.Unlock()

	ch, err := c.producerChannelLocked()
	if err != nil {
		return err
	}

	// 丢弃上一轮「确认超时」可能残留的回执，否则会把旧回执误判成本次发布成功
	c.drainConfirms()

	publishing := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         msg,
		DeliveryMode: amqp091.Persistent, // 持久化消息
	}
	if ttl > 0 {
		publishing.Expiration = strconv.FormatInt(ttl.Milliseconds(), 10)
	}

	// 写入用调用方 ctx（调用方已放弃就没必要再写）；
	// 但等待确认用独立的 context——调用方 ctx 可能在这期间超时，
	// 那属于「发送成功但确认未回」，不能当成发送失败触发业务回滚。
	if err = ch.PublishWithContext(ctx, exchange, key, false, false, publishing); err != nil {
		return fmt.Errorf("发布消息失败: exchange=%s, key=%s: %w", exchange, key, err)
	}

	confirmCtx, cancel := context.WithTimeout(context.Background(), publishConfirmTimeout)
	defer cancel()

	select {
	case conf, ok := <-c.confirms:
		if !ok {
			return fmt.Errorf("等待发布确认失败: confirm channel 已关闭, exchange=%s, key=%s", exchange, key)
		}
		if !conf.Ack {
			return fmt.Errorf("broker 拒绝消息(Nack): exchange=%s, key=%s", exchange, key)
		}
		return nil
	case <-confirmCtx.Done():
		return fmt.Errorf("等待发布确认超时(%s): exchange=%s, key=%s", publishConfirmTimeout, exchange, key)
	}
}

// drainConfirms 非阻塞清空 confirm 通道中的残留回执
func (c *Client) drainConfirms() {
	for {
		select {
		case <-c.confirms:
			continue
		default:
			return
		}
	}
}
