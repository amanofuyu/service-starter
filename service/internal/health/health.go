package health

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"
)

// Checker 表示可以被 readiness 探测的外部依赖。
type Checker interface {
	Ping(context.Context) error
}

// Dependencies 汇总 /readyz 需要检查的依赖；nil 依赖会被跳过。
type Dependencies struct {
	// Postgres 通常传入 pgxpool.Pool，它本身提供 Ping 方法。
	Postgres Checker
	// Redis 传入 internal/redis.Client，它把 go-redis 包装成 Checker。
	Redis Checker
	// Kafka 是可选依赖；未配置 KAFKA_BROKERS 时保持 nil。
	Kafka Checker
}

// Handler 提供健康检查 HTTP handler。
type Handler struct {
	deps Dependencies
}

// NewHandler 创建健康检查 handler。
func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

// Healthz 只表示进程仍在响应，不检查数据库、Redis 或 Kafka。
func (h Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Readyz 检查已配置依赖是否可用，任一依赖失败都会返回 503。
func (h Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	// readiness 检查不能无限等待，否则调用方会被卡住；这里给所有依赖共享 3 秒超时。
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ok := true
	checks := map[string]string{}
	check := func(name string, checker Checker) {
		// nil 表示该依赖未启用，例如没有配置 Kafka 时不应让 readiness 失败。
		if isNilChecker(checker) {
			return
		}
		if err := checker.Ping(ctx); err != nil {
			// 响应中只暴露 error 状态，不把内部错误详情直接返回给调用方。
			checks[name] = "error"
			ok = false
			return
		}
		checks[name] = "ok"
	}

	check("postgres", h.deps.Postgres)
	check("redis", h.deps.Redis)
	check("kafka", h.deps.Kafka)

	status := http.StatusOK
	if !ok {
		// readiness 失败使用 503，方便负载均衡器或部署脚本判断服务暂不可接流量。
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{
		"ok":     ok,
		"checks": checks,
	})
}

// isNilChecker 同时处理 interface nil 和 typed nil，避免可选 Kafka checker 被误判为可用依赖。
func isNilChecker(checker Checker) bool {
	if checker == nil {
		return true
	}
	// Go 的 interface 可能装着一个 typed nil，例如 var c *Checker = nil。
	// 此时 checker != nil，但真正调用方法会有风险，所以需要用反射再判断一次。
	value := reflect.ValueOf(checker)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// writeJSON 统一健康检查响应格式；编码失败时只能忽略，因为响应头已经写出。
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
