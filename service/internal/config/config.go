package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config 汇总服务运行所需配置。字段只来自环境变量，避免代码依赖 Compose 专用变量。
type Config struct {
	// AppEnv 标识运行环境，常见值是 development、staging、production。
	AppEnv string
	// DatabaseURL 是 PostgreSQL 最终连接串，应用不再单独读取数据库用户名和密码。
	DatabaseURL string
	// RedisURL 是 Redis 最终连接串，包含密码、主机、端口和 DB 编号。
	RedisURL string
	// ServicePort 是 HTTP server 在容器内监听的端口。
	ServicePort string
	// OTELServiceName 会写入 trace resource，用于在 Jaeger 中区分服务。
	OTELServiceName string
	// OTELTracesExporter 控制是否启用 tracing；当前只支持 none 和 otlp。
	OTELTracesExporter string
	// OTELExporterOTLPEnd 是 OTLP HTTP endpoint，例如 http://jaeger:4318。
	OTELExporterOTLPEnd string
	// KafkaBrokers 为空时表示不启用 Kafka readiness 检查。
	KafkaBrokers []string
	// KafkaTopicPrefix 用于给不同环境的主题名加前缀。
	KafkaTopicPrefix string
}

// Load 读取环境变量、应用默认值，并校验启动所需的最小配置。
func Load() (Config, error) {
	cfg := Config{
		AppEnv:              getenvDefault("APP_ENV", "development"),
		DatabaseURL:         strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:            strings.TrimSpace(os.Getenv("REDIS_URL")),
		ServicePort:         getenvDefault("SERVICE_PORT", "8081"),
		OTELServiceName:     getenvDefault("OTEL_SERVICE_NAME", "service"),
		OTELTracesExporter:  getenvDefault("OTEL_TRACES_EXPORTER", "none"),
		OTELExporterOTLPEnd: strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		KafkaBrokers:        splitCSV(os.Getenv("KAFKA_BROKERS")),
		KafkaTopicPrefix:    strings.TrimSpace(os.Getenv("KAFKA_TOPIC_PREFIX")),
	}

	var missing []string
	// 数据库和 Redis 是核心栈必需依赖；缺失时启动没有意义。
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// exporter 值越少越容易排错。新增 exporter 类型时，应同步扩展 observability.Init。
	switch cfg.OTELTracesExporter {
	case "none", "otlp":
	default:
		return Config{}, errors.New("OTEL_TRACES_EXPORTER must be either none or otlp")
	}
	if err := validatePort("SERVICE_PORT", cfg.ServicePort); err != nil {
		return Config{}, err
	}
	if err := validateURLScheme("DATABASE_URL", cfg.DatabaseURL, "postgres", "postgresql"); err != nil {
		return Config{}, err
	}
	if err := validateURLScheme("REDIS_URL", cfg.RedisURL, "redis", "rediss"); err != nil {
		return Config{}, err
	}
	if cfg.OTELTracesExporter == "otlp" {
		if cfg.OTELExporterOTLPEnd == "" {
			return Config{}, errors.New("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_TRACES_EXPORTER is otlp")
		}
		if err := validateURLScheme("OTEL_EXPORTER_OTLP_ENDPOINT", cfg.OTELExporterOTLPEnd, "http", "https"); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

// validatePort 确认端口是 TCP/UDP 常见的 1-65535 范围。
func validatePort(key, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%s must be a port between 1 and 65535", key)
	}
	return nil
}

// validateURLScheme 校验 URL 格式和协议，避免启动后才暴露明显配置错误。
func validateURLScheme(key, value string, allowed ...string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid URL", key)
	}
	for _, scheme := range allowed {
		if parsed.Scheme == scheme {
			return nil
		}
	}
	return fmt.Errorf("%s must use one of these schemes: %s", key, strings.Join(allowed, ", "))
}

// getenvDefault 将空字符串视为未设置，避免空环境变量覆盖默认值。
func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// splitCSV 解析逗号分隔列表，并忽略空白项，适用于 Kafka broker 列表。
func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		// 允许用户写成 "kafka:9092, localhost:19092"，减少配置格式的小坑。
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
