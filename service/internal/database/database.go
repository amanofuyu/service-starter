package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool 创建 PostgreSQL 连接池。pgxpool.New 不会立刻要求数据库可达，连通性由 /readyz 负责检查。
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// 这里保留为一层很薄的封装，后续可以集中添加连接池参数、日志或 tracing 配置。
	return pgxpool.New(ctx, databaseURL)
}
