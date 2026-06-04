package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"service-starter/service/internal/health"
)

// NewRouter 注册服务的 HTTP 路由。serviceName 会出现在 /api/ping 响应中，方便克隆项目后验证重命名是否完成。
func NewRouter(serviceName string, deps health.Dependencies) http.Handler {
	r := chi.NewRouter()
	healthHandler := health.NewHandler(deps)

	// /healthz 不检查外部依赖，适合用作进程存活探针。
	r.Get("/healthz", healthHandler.Healthz)
	// /readyz 会检查数据库、Redis 和可选 Kafka，适合用作接流量前的就绪探针。
	r.Get("/readyz", healthHandler.Readyz)
	// /api/ping 是最小业务示例，新手可以从这个 handler 学习如何写 JSON 响应。
	r.Get("/api/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"service": serviceName,
		})
	})

	return withMiddleware(serviceName, r)
}

// withMiddleware 给所有路由统一增加请求标识、panic 恢复和 HTTP tracing。
func withMiddleware(serviceName string, next http.Handler) http.Handler {
	handler := middleware.Recoverer(next)
	handler = middleware.RealIP(handler)
	handler = requestIDResponseHeader(handler)
	handler = middleware.RequestID(handler)
	return otelhttp.NewHandler(handler, serviceName)
}

// requestIDResponseHeader 把 chi 生成或读取的 request id 回写给调用方，便于日志关联。
func requestIDResponseHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID := middleware.GetReqID(r.Context()); requestID != "" {
			w.Header().Set(middleware.RequestIDHeader, requestID)
		}
		next.ServeHTTP(w, r)
	})
}
