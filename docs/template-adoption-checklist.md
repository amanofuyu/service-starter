# 模板落地检查清单

将基础项目克隆为真实服务后，使用此检查清单。

## 克隆前判断

适合直接使用本模板的情况：

- 业务服务以 HTTP API 为主。
- 本地开发需要 PostgreSQL 和 Redis。
- Kafka、链路追踪和日志聚合是可选能力。
- 团队希望从小而清晰的服务骨架开始，而不是从完整平台开始。

不建议直接扩展本模板的情况：

- 第一版就需要多服务 monorepo。
- 已经确定要运行在 Kubernetes 并且需要完整 Helm/Kustomize 结构。
- 已有统一认证、网关、CI/CD 或基础设施模板必须接入。
- 服务核心不是 Go HTTP 后端。

## 重命名面向业务的值

- `POSTGRES_DB`
- `OTEL_SERVICE_NAME`
- README 标题
- Go 模块路径
- `/api/ping` 服务值
- 如已自定义备份文件名前缀，也一并修改

建议搜索旧模板标识：

```sh
rg "service-starter|service"
```

注意不要机械替换 Compose 服务名 `service`。它是模板的基础约定，只有存在多服务或明确运维命名要求时才改。

## 非必要不改默认值

- 保持 Compose 服务名为 `service`。
- 保持 PostgreSQL 和 Redis 只在 Compose 网络内可访问。
- 保持 Kafka 可选。
- 保持可观测性可选。
- 不要为单个服务添加 API 网关。
- 在项目有明确需求前，不要添加 Kubernetes、认证、OpenAPI、CI 或迁移工具。

## 业务代码接入

推荐顺序：

1. 在 `internal/httpserver` 新增业务路由。
2. 为业务逻辑创建独立 package，例如 `internal/user`、`internal/order` 或 `internal/usecase`。
3. 需要配置时先扩展 `internal/config.Config`。
4. 需要新外部依赖时创建小型 adapter，并在 `internal/app` 装配。
5. 需要 readiness 检查时实现 `health.Checker` 并加入 `health.Dependencies`。

避免一开始就把所有业务逻辑写进路由闭包。路由层应保持薄，便于测试和后续拆分。

## 环境与密钥

- `.env.example` 可以保留开发默认值，但要改成真实项目语义。
- `.env` 不提交。
- 生产环境不要使用 `.env.example` 中的开发密码。
- Compose 中可以派生 `DATABASE_URL` 和 `REDIS_URL`，应用代码只读取最终 URL。
- 新增 secret 时优先考虑部署环境的 secret 管理能力，不要硬编码到 Compose 文件。

## 文档同步

克隆后至少同步以下文档：

- `README.md`：项目名、适用场景、接口示例和启动方式。
- `docs/developer-guide.md`：业务模块结构、本地依赖和测试策略。
- `docs/operations-guide.md`：环境差异、备份恢复、日志和监控入口。
- `docs/template-design.md`：如果继续保留它，注明它是基础模板设计说明，而不是业务项目设计文档。

## 自定义后验证

```sh
cd service && go test ./...
docker compose config
docker compose up -d
set -a
. ./.env
set +a
HOST_PORT="${SERVICE_PORT:-8081}"
curl "http://127.0.0.1:${HOST_PORT}/healthz"
curl "http://127.0.0.1:${HOST_PORT}/readyz"
curl "http://127.0.0.1:${HOST_PORT}/api/ping"
```

`SERVICE_PORT` 控制宿主机端口映射；容器内服务仍固定监听 `8081`。如果 `.env` 中写的是 `SERVICE_PORT=18081`，上面的 `curl` 应访问 `18081`。

如果启用 Kafka：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml config
docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d
docker compose -f docker-compose.yml -f docker-compose.kafka.yml --profile tools run --rm kafka-init
curl "http://127.0.0.1:${HOST_PORT}/readyz"
```

如果启用可观测性：

```sh
docker compose -f docker-compose.yml -f docker-compose.obs.yml config
docker compose -f docker-compose.yml -f docker-compose.obs.yml up -d
curl "http://127.0.0.1:${HOST_PORT}/api/ping"
```

然后打开：

- Grafana: `http://127.0.0.1:3000`
- Jaeger: `http://127.0.0.1:16686`
