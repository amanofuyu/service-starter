package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"service-starter/service/internal/config"
)

// Shutdown 表示可观测性资源的关闭函数，调用方应在进程退出前执行。
type Shutdown func(context.Context) error

// Init 根据配置初始化 tracing；OTEL_TRACES_EXPORTER=none 时返回空关闭函数。
func Init(ctx context.Context, cfg config.Config) (Shutdown, error) {
	if cfg.OTELTracesExporter == "none" {
		// 返回空函数可以让调用方无条件 defer shutdown，减少分支判断。
		return func(context.Context) error { return nil }, nil
	}

	options := []otlptracehttp.Option{}
	if cfg.OTELExporterOTLPEnd != "" {
		// 使用完整 endpoint URL，便于 Compose 和非 Compose 环境共用同一个配置字段。
		options = append(options, otlptracehttp.WithEndpointURL(cfg.OTELExporterOTLPEnd))
	}
	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	// TracerProvider 是 OpenTelemetry SDK 的核心对象，负责批量导出 span。
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			// ServiceName 会显示在 Jaeger 等追踪系统里，克隆项目后应改成真实服务名。
			semconv.ServiceName(cfg.OTELServiceName),
		)),
	)
	// 设置全局 provider 后，后续业务代码可以通过 otel.Tracer 获取 tracer。
	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}
