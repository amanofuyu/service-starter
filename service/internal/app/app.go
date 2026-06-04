package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"service-starter/service/internal/config"
	"service-starter/service/internal/database"
	"service-starter/service/internal/health"
	"service-starter/service/internal/httpserver"
	"service-starter/service/internal/kafka"
	"service-starter/service/internal/observability"
	redisstore "service-starter/service/internal/redis"
)

// Run 装配服务依赖并阻塞运行 HTTP server，直到收到退出信号或启动失败。
func Run(ctx context.Context) error {
	// 第一步先加载配置。配置不完整时直接返回错误，避免服务启动后才在请求中暴露问题。
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 使用结构化 JSON 日志，方便 Docker、Loki 或其他日志系统按字段检索。
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	// tracing 是可选能力。默认配置会返回一个空 shutdown 函数，不会连接 Jaeger。
	otelShutdown, err := observability.Init(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			logger.Error("failed to shut down tracing", "error", err)
		}
	}()

	// 创建连接池不等于数据库已经可用；真正的连通性由 /readyz 里的 Ping 检查。
	pgPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pgPool.Close()

	// Redis client 创建后也不会主动执行 ping，避免启动阶段被临时网络抖动放大。
	redisClient, err := redisstore.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	var kafkaChecker *kafka.Checker
	// Kafka 是可选组件。只有配置了 broker，服务才创建 checker 并把 Kafka 纳入 readiness。
	if len(cfg.KafkaBrokers) > 0 {
		kafkaChecker, err = kafka.NewChecker(cfg.KafkaBrokers)
		if err != nil {
			return err
		}
		defer kafkaChecker.Close()
	}

	// readiness 依赖在这里注入；Kafka checker 为空时 /readyz 会自动跳过 Kafka。
	server := &http.Server{
		Addr:              ":" + cfg.ServicePort,
		Handler:           httpserver.NewRouter(cfg.OTELServiceName, health.Dependencies{Postgres: pgPool, Redis: redisClient, Kafka: kafkaChecker}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// signal.NotifyContext 会在收到 Ctrl+C(SIGINT) 或容器停止(SIGTERM) 时取消 runCtx。
	runCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		// ListenAndServe 会一直阻塞；放到 goroutine 后，主流程才能同时等待退出信号。
		logger.Info("starting service", "addr", server.Addr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-runCtx.Done():
		// 给正在处理的请求留出有限时间完成，避免 SIGTERM 时直接中断连接。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		// 等 ListenAndServe goroutine 退出，确保没有隐藏的 server 启动错误被丢掉。
		return <-errCh
	case err := <-errCh:
		// 如果 server 自己提前退出，直接把错误交给 main，由 main 记录并结束进程。
		return err
	}
}
