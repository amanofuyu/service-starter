package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// Client 包装 go-redis client，使其符合 health.Checker 的 Ping 接口。
type Client struct {
	inner *goredis.Client
}

// NewClient 从 Redis URL 创建 client；实际连通性由 Ping/readiness 检查。
func NewClient(_ context.Context, redisURL string) (*Client, error) {
	// ParseURL 负责解析 redis://:password@host:port/db 这种标准 URL。
	options, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	// NewClient 只创建客户端对象，不会立即发起网络请求。
	client := goredis.NewClient(options)
	return &Client{inner: client}, nil
}

// Ping 用于 /readyz 判断 Redis 是否可用。
func (c *Client) Ping(ctx context.Context) error {
	return c.inner.Ping(ctx).Err()
}

// Close 释放 Redis client 持有的连接资源。
func (c *Client) Close() error {
	return c.inner.Close()
}
