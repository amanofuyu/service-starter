package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"service-starter/service/internal/config"
	"service-starter/service/internal/health"
)

func TestNewServerUsesConfiguredAddressAndTimeouts(t *testing.T) {
	cfg := config.Config{
		ServicePort:     "18081",
		OTELServiceName: "service",
	}

	server := newServer(cfg, health.Dependencies{})

	if server.Addr != ":18081" {
		t.Fatalf("Addr = %q, want :18081", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout = %s, want 10s", server.ReadTimeout)
	}
	if server.WriteTimeout != 30*time.Second {
		t.Fatalf("WriteTimeout = %s, want 30s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want 60s", server.IdleTimeout)
	}
	if server.Handler == nil {
		t.Fatal("expected server handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
