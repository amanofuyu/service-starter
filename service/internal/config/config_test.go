package config_test

import (
	"testing"

	"service-starter/service/internal/config"
)

func TestLoadRequiresDatabaseAndRedisURLs(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("SERVICE_PORT", "8081")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected missing DATABASE_URL and REDIS_URL to fail")
	}
}

func TestLoadAppliesDefaultsAndOptionalSettings(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "postgres://service:secret@localhost:5432/service?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://:secret@localhost:6379/0")
	t.Setenv("SERVICE_PORT", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("KAFKA_BROKERS", "kafka:9092,localhost:19092")
	t.Setenv("KAFKA_TOPIC_PREFIX", "local")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://jaeger:4318")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.ServicePort != "8081" {
		t.Fatalf("ServicePort = %q, want 8081", cfg.ServicePort)
	}
	if cfg.OTELServiceName != "service" {
		t.Fatalf("OTELServiceName = %q, want service", cfg.OTELServiceName)
	}
	if cfg.OTELTracesExporter != "otlp" {
		t.Fatalf("OTELTracesExporter = %q, want otlp", cfg.OTELTracesExporter)
	}
	if got, want := len(cfg.KafkaBrokers), 2; got != want {
		t.Fatalf("len(KafkaBrokers) = %d, want %d", got, want)
	}
}
