package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"service-starter/service/internal/health"
)

type middlewareFakeChecker struct{}

func (middlewareFakeChecker) Ping(context.Context) error {
	return nil
}

func TestRouterAddsRequestIDResponseHeader(t *testing.T) {
	router := NewRouter("service", health.Dependencies{
		Postgres: middlewareFakeChecker{},
		Redis:    middlewareFakeChecker{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id response header")
	}
}

func TestMiddlewareRecoversPanic(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	handler := withMiddleware("service", r)
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
