package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Checker 是 Kafka readiness 适配器，只在配置了 KAFKA_BROKERS 时创建。
type Checker struct {
	client *kgo.Client
}

// NewChecker 使用 broker 列表创建 franz-go client。
func NewChecker(brokers []string) (*Checker, error) {
	// SeedBrokers 告诉 client 从哪些 broker 开始发现 Kafka 集群。
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return nil, err
	}
	return &Checker{client: client}, nil
}

// Ping 用于 /readyz 判断 Kafka broker 是否可达。
func (c *Checker) Ping(ctx context.Context) error {
	// franz-go 的 Ping 会尝试和 broker 通信；失败时 readiness 会返回 503。
	return c.client.Ping(ctx)
}

// Close 释放 Kafka client；允许 nil 调用，方便 app.Run 做统一清理。
func (c *Checker) Close() {
	if c != nil && c.client != nil {
		c.client.Close()
	}
}
