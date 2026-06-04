package httpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"service-starter/service/internal/health"
	"service-starter/service/internal/httpserver"
)

type fakeChecker struct{}

func (fakeChecker) Ping(context.Context) error {
	return nil
}

func TestRouterServesPing(t *testing.T) {
	router := httpserver.NewRouter("service", health.Dependencies{
		Postgres: fakeChecker{},
		Redis:    fakeChecker{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"ok":true,"service":"service"}` {
		t.Fatalf("body = %s, want ping JSON", got)
	}
}
