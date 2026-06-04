package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"service-starter/service/internal/health"
)

type fakeChecker struct {
	err error
}

func (f fakeChecker) Ping(context.Context) error {
	return f.err
}

func TestHealthzReturnsProcessHealth(t *testing.T) {
	handler := health.NewHandler(health.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("response body = %s, want ok true", rec.Body.String())
	}
}

func TestReadyzReturnsOKWhenDependenciesPass(t *testing.T) {
	handler := health.NewHandler(health.Dependencies{
		Postgres: fakeChecker{},
		Redis:    fakeChecker{},
	})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.Readyz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"postgres":"ok"`) {
		t.Fatalf("response body = %s, want postgres ok", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"redis":"ok"`) {
		t.Fatalf("response body = %s, want redis ok", rec.Body.String())
	}
}

func TestReadyzReturnsUnavailableWhenDependencyFails(t *testing.T) {
	handler := health.NewHandler(health.Dependencies{
		Postgres: fakeChecker{err: errors.New("database down")},
		Redis:    fakeChecker{},
	})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.Readyz(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), `"postgres":"error"`) {
		t.Fatalf("response body = %s, want postgres error", rec.Body.String())
	}
}

type pointerChecker struct {
	err error
}

func (p *pointerChecker) Ping(context.Context) error {
	return p.err
}

func TestReadyzSkipsTypedNilChecker(t *testing.T) {
	var kafkaChecker *pointerChecker
	handler := health.NewHandler(health.Dependencies{
		Postgres: fakeChecker{},
		Redis:    fakeChecker{},
		Kafka:    kafkaChecker,
	})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.Readyz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), `"kafka"`) {
		t.Fatalf("response body = %s, want typed nil kafka checker to be skipped", rec.Body.String())
	}
}
