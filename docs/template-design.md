# 模板设计说明

## 1. 项目目标

构建一个面向中小型系统的可复用后端服务基础项目。

这个基础项目应当可以直接克隆使用。对于新的业务项目，开发者主要修改服务代码、环境名称、数据库名称和文档，而不是从零重新搭建基础设施。

默认技术栈应保持轻量：

- Go 应用服务
- PostgreSQL
- Redis
- 可选 Kafka 消息栈
- 可选可观测性栈

项目应优化 Codex 辅助开发体验。优先采用明确约定、简单结构、可重复命令以及便于验证的代码。

## 2. 非目标

这个基础项目不打算成为：

- Kubernetes 平台
- 默认完整可观测性平台
- 重型微服务框架
- 从第一天开始就是多服务 monorepo
- 生产级安全或合规方案
- 默认 API 网关脚手架

第一版应保持朴素、易理解、易扩展。

## 3. 架构

默认运行拓扑：

```text
client
  |
  v
service :8081
  |
  +--> postgres :5432
  |
  +--> redis :6379
```

可选 Kafka 消息拓扑：

```text
service
  |
  +--> kafka :9092
```

可选可观测性拓扑：

```text
service
  |
  +--> jaeger :4318 / :4317

docker containers
  |
  v
alloy
  |
  v
loki
  |
  v
grafana :3000
```

当真实项目需要跨服务路由、边缘认证、限流、请求转换、聚合或统一外部 API 边界时，可以之后再添加 API 网关。不要为了代理单个服务而在默认基础项目中加入网关。

## 4. 推荐目录结构

```text
.
|-- docker-compose.yml
|-- docker-compose.kafka.yml
|-- docker-compose.obs.yml
|-- .env.example
|-- .env
|-- loki-config.yml
|-- alloy-config.alloy
|-- Makefile
|-- README.md
|-- docs/
|   |-- index.md
|   |-- developer-guide.md
|   |-- operations-guide.md
|   |-- template-adoption-checklist.md
|   |-- template-design.md
|   `-- ai-agent-guide.md
`-- service/
    |-- Dockerfile
    |-- go.mod
    |-- go.sum
    |-- cmd/
    |   `-- service/
    |       `-- main.go
    |-- internal/
    |   |-- app/
    |   |-- config/
    |   |-- database/
    |   |-- redis/
    |   |-- httpserver/
    |   |-- observability/
    |   `-- health/
    `-- migrations/
```

Docker Compose 的服务名保持为 `service`。基础项目中不要命名为 `go-service`、`node-service` 或具体业务域名。

`service/migrations/` 目录可以作为第一版占位存在，但在第一次真实 schema 迁移出现之前，不要添加迁移工具依赖。

克隆此基础项目用于真实项目时，只重命名面向业务的值，例如：

- `POSTGRES_DB`
- `OTEL_SERVICE_NAME`
- README 标题
- Go 模块路径
- 默认 JSON 响应中的服务名
- 如新增备份文件前缀，也一并修改

## 5. Docker Compose 设计

使用三个 Compose 文件：

- `docker-compose.yml`：轻量核心栈
- `docker-compose.kafka.yml`：可选 Kafka 消息栈
- `docker-compose.obs.yml`：可选可观测性栈

核心服务：

- `postgres`
- `redis`
- `service`
- `tools` profile 下的 `pg-backup`

Kafka 服务：

- `kafka`
- 如需主题引导，使用 `tools` profile 下的 `kafka-init`

可观测性服务：

- `jaeger`
- `loki`
- `alloy`
- `grafana`

默认启动：

```sh
docker compose up -d
```

启用 Kafka 消息：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d
```

启用可观测性：

```sh
docker compose -f docker-compose.yml -f docker-compose.obs.yml up -d
```

同时启用 Kafka 和可观测性：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml -f docker-compose.obs.yml up -d
```

数据库备份：

```sh
docker compose --profile tools run --rm pg-backup
```

Compose 规则：

- 为该栈使用专用 Compose 网络。
- 默认将 Go 服务的宿主机端口绑定到 `127.0.0.1`。
- 默认不要将 PostgreSQL 或 Redis 暴露到宿主机。
- PostgreSQL 和 Redis 只能由 Go 服务通过 Compose 服务名访问。
- 为 PostgreSQL 和 Redis 添加健康检查。
- PostgreSQL 健康检查使用 `pg_isready`。
- Redis 健康检查必须使用配置的 Redis 密码认证。
- PostgreSQL 和 Redis 数据使用命名卷。
- 可选工具放在 Compose profiles 后面。
- 默认以 KRaft 模式运行 Kafka；除非存在真实兼容性需求，否则不要添加 ZooKeeper。
- 默认不要将 Kafka 暴露到宿主机，除非本地生产者或消费者工具需要。
- 如果为了本地工具暴露 Kafka，请绑定到 `127.0.0.1`，并记录 advertised listener 行为。

## 6. 配置规则

配置必须来自环境变量。

Compose 必需变量：

```text
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD
REDIS_PASSWORD
SERVICE_PORT
```

Go 服务读取的应用变量：

```text
APP_ENV
DATABASE_URL
REDIS_URL
OTEL_SERVICE_NAME
OTEL_TRACES_EXPORTER
OTEL_EXPORTER_OTLP_ENDPOINT
KAFKA_BROKERS
KAFKA_TOPIC_PREFIX
```

连接字符串规则：

- Compose 可以从 `POSTGRES_DB`、`POSTGRES_USER` 和 `POSTGRES_PASSWORD` 派生 `DATABASE_URL`。
- Compose 可以从 `REDIS_PASSWORD` 派生 `REDIS_URL`。
- Go 服务只读取 `DATABASE_URL` 和 `REDIS_URL`，不要从单独的数据库变量中重新组装连接串。
- 在 Compose 内使用 `postgres` 和 `redis` 作为主机名。
- 在非 Compose 的本地运行中，开发者可以直接覆盖 `DATABASE_URL` 和 `REDIS_URL`。
- 启用 Kafka 的代码应将 `KAFKA_BROKERS` 读取为逗号分隔的 broker 列表。
- `KAFKA_TOPIC_PREFIX` 可用于区分本地、预发和生产主题名。

规则：

- 永远不要在 Compose 中硬编码密码。
- 提交 `.env.example`。
- 真实 `.env` 保持 Git 忽略。
- 固定 Docker 镜像版本。
- 升级镜像 tag 或 Go 依赖前，先联网查询当前上游版本。

## 7. 版本矩阵

第一版实现应在代码和 Compose 中固定具体版本。实现时先联网查询当前上游版本，再填写此表。

| 组件 | 版本 / 镜像 tag | 说明 |
| --- | --- | --- |
| Go | 1.26.4 / `golang:1.26.4-alpine3.23` | 用于本地开发和 Docker 构建阶段 |
| PostgreSQL | `postgres:18.4` | 使用稳定主版本，不使用 `latest` |
| Redis | `redis:8.8.0-alpine` | 使用稳定主版本，不使用 `latest` |
| Kafka | `apache/kafka:4.3.0` | 仅作为可选消息组件；优先使用 KRaft 模式 |
| Jaeger | `jaegertracing/all-in-one:1.76.0` | 仅作为可选可观测性组件 |
| Loki | `grafana/loki:3.7.2` | 仅作为可选可观测性组件 |
| Alloy | `grafana/alloy:v1.16.2` | 仅作为可选可观测性组件 |
| Grafana | `grafana/grafana:13.0.2` | 仅作为可选可观测性组件 |

## 8. Go 服务要求

Go 服务应暴露：

```text
GET /healthz
GET /readyz
GET /api/ping
```

预期行为：

- `/healthz` 只返回进程健康状态。
- `/readyz` 检查 PostgreSQL 和 Redis 连通性。
- `/api/ping` 返回简单 JSON 响应，并证明服务可访问。

示例响应：

```json
{
  "ok": true,
  "service": "service"
}
```

推荐 Go 技术栈：

- HTTP router：标准库 `net/http` 或 `chi`
- PostgreSQL driver：`pgx`
- Redis client：`go-redis`
- 启用消息时的 Kafka client：联网查询当前版本后选择 `franz-go` 或 `segmentio/kafka-go`
- Logging：标准库 `log/slog`
- Config：环境变量
- Tests：标准 `go test`

除非有具体理由，否则避免引入大型 Web 框架。

## 9. 服务生命周期

应用应当：

- 启动时加载配置。
- 初始化 logger。
- 连接 PostgreSQL。
- 连接 Redis。
- 注册 HTTP 路由。
- 启动 HTTP server。
- 处理优雅关闭。
- 关闭时关闭已启用的数据库、Redis 和 Kafka client。

关闭要求：

- 监听 `SIGINT` 和 `SIGTERM`。
- 停止接受新的 HTTP 请求。
- 允许进行中的请求在超时时间内完成。
- 关闭外部 client。

## 10. 健康检查

Compose 健康检查应覆盖：

- PostgreSQL
- Redis

Go 服务应提供应用级 readiness 端点。初始 Compose 文件可以不依赖服务健康状态，除非 Docker 镜像中包含可靠的本地 healthcheck 命令。

不要假设 `depends_on` 等同于“已就绪”。除非使用 `condition: service_healthy`，它只控制启动顺序。

## 11. Kafka 消息

Kafka 默认可选。包含它是因为队列和事件流在服务项目中很常见，但最小核心栈不应依赖它。Kafka 会增加启动成本、存储、advertised listener 复杂度以及额外运维概念，而小型服务未必一开始就需要。

Kafka 模式：

- 使用 `docker-compose.kafka.yml`。
- 本地开发优先使用单节点 KRaft 模式 Kafka。
- broker 数据放入命名卷。
- 主题引导保持显式并写入文档。
- 不要仅因为 Kafka 被禁用就让 `/readyz` 失败。
- 如果服务编译进 Kafka 支持，readiness 只应在本次运行配置了 Kafka 且要求 Kafka 时检查 Kafka。

如需示例，推荐基础主题：

```text
service.events
service.commands
```

在基础服务有具体示例用例前，不要添加 Kafka 业务消费者或生产者。基础项目中最小的管理或连通性检查就足够。

## 12. 可观测性

可观测性默认可选。

默认模式：

- 使用 `docker compose logs`。
- 使用带轮转的 JSON-file logging。
- 保持链路追踪禁用。

可观测性模式：

- Jaeger 通过 OTLP 接收 traces。
- Alloy 收集 Docker 容器日志。
- Loki 存储日志。
- Grafana 读取 Loki。

链路追踪应由以下配置控制：

```text
OTEL_TRACES_EXPORTER=none
OTEL_TRACES_EXPORTER=otlp
```

如果链路追踪被禁用，服务不应启动失败。

Alloy 日志采集可能需要访问 Docker 日志或 Docker socket，具体取决于所选配置。将可观测性 Compose 文件视为本地开发便利能力，而不是生产安全模型。

## 13. 安全基线

最低基线：

- 默认将暴露到宿主机的端口绑定到 `127.0.0.1`。
- 默认不要将 PostgreSQL 或 Redis 暴露到宿主机。
- 使用 Redis 密码。
- 不要把 secret 放入 Git。
- 配置文件使用只读挂载。
- 固定 Docker 镜像。
- 添加日志轮转。
- 可行时让 Go 服务容器以非 root 用户运行。

这不能替代生产加固。真实部署仍需要合适的 secret 管理、TLS、备份保留、监控和访问控制。

## 14. Makefile 命令

推荐命令：

```makefile
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

kafka-up:
	docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d

kafka-down:
	docker compose -f docker-compose.yml -f docker-compose.kafka.yml down

obs-up:
	docker compose -f docker-compose.yml -f docker-compose.obs.yml up -d

obs-down:
	docker compose -f docker-compose.yml -f docker-compose.obs.yml down

backup:
	docker compose --profile tools run --rm pg-backup

test:
	cd service && go test ./...

fmt:
	cd service && gofmt -w .
```

## 15. README 要求

README 应包含：

- 这个基础项目的用途
- 快速开始
- 目录结构
- 必需环境变量
- 核心栈命令
- Kafka 消息栈命令
- 可观测性栈命令
- 如何克隆为新项目
- 如何验证服务

快速验证命令：

```sh
HOST_PORT="${SERVICE_PORT:-8081}"
curl "http://127.0.0.1:${HOST_PORT}/healthz"
curl "http://127.0.0.1:${HOST_PORT}/readyz"
curl "http://127.0.0.1:${HOST_PORT}/api/ping"
```

这里的 `SERVICE_PORT` 是宿主机映射端口；容器内服务仍固定监听 `8081`。

## 16. Codex 开发说明

Codex 开发这个基础项目时，应遵循以下规则：

- 优先使用简单 Go 代码，而不是重框架代码。
- Compose 服务名保持为 `service`。
- 默认栈保持轻量。
- 第一版不要添加网关。
- 将 Kafka 作为可选 Compose 栈添加，不要作为默认核心启动的一部分。
- 默认不要启用可观测性。
- 第一版不要添加 Kubernetes 文件。
- 在基础服务稳定前不要引入认证。
- 为配置加载、健康处理器和 readiness 检查编写测试。
- 声称完成前先运行格式化和测试。
- 添加或升级依赖时，先联网查询当前版本。
- 后续新增或修改的项目文档与代码注释统一使用中文书写；命令、标识符、配置键、协议名和第三方专有名词按原文保留。

当前仓库可以包含轻量的 GitHub Actions 检查，用于运行格式、测试、静态检查和 Compose 配置校验。复杂发布流水线、部署编排和环境推广流程仍不是基础模板默认目标。

## 17. 初始实现任务

任务 1：创建核心 Compose 文件

- 创建 `docker-compose.yml`。
- 创建 `.env.example`。
- 确保核心栈可以通过 `docker compose up -d` 启动。
- 确保默认只将 Go 服务暴露到宿主机。

任务 2：创建 Go 服务骨架

- 在 `service/` 下创建 Go module。
- 添加 HTTP server。
- 添加 `/healthz`、`/readyz`、`/api/ping`。
- 添加配置加载器。
- 添加优雅关闭。

任务 3：添加 PostgreSQL 和 Redis client

- 使用 `DATABASE_URL` 和 `REDIS_URL`。
- 实现 readiness 检查。
- 添加基础测试。

任务 4：添加 Dockerfile

- 使用多阶段构建。
- 可行时构建静态 Go binary。
- 可行时以非 root 用户运行。
- 保持最终镜像小。

任务 5：添加可选 Kafka 消息

- 创建 `docker-compose.kafka.yml`。
- 使用单节点 KRaft 模式 Kafka。
- 添加 `KAFKA_BROKERS` 和 `KAFKA_TOPIC_PREFIX` 示例。
- 添加可选 Kafka 连通性检查，同时不要让核心模式依赖 Kafka。

任务 6：添加可选可观测性

- 创建 `docker-compose.obs.yml`。
- 创建 Loki 和 Alloy 配置。
- 添加可选 OpenTelemetry 设置。
- 默认保持链路追踪禁用。

任务 7：添加项目文档

- 创建 README。
- 创建 `docs/developer-guide.md`。
- 创建 `docs/operations-guide.md`。
- 记录克隆和自定义工作流。

## 18. 验收标准

核心模式通过条件：

- `docker compose config` 成功。
- `docker compose up -d` 启动 PostgreSQL、Redis 和 service。
- PostgreSQL 健康检查变为 healthy。
- Redis 健康检查变为 healthy。
- 按 `.env` 中的 `SERVICE_PORT` 访问宿主机端口时，`/healthz`、`/readyz` 和 `/api/ping` 成功。
- `cd service && go test ./...` 成功。

Kafka 模式通过条件：

- `docker compose -f docker-compose.yml -f docker-compose.kafka.yml config` 成功。
- `docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d` 启动 PostgreSQL、Redis、service 和 Kafka。
- Kafka 接受来自 Compose 网络内的 broker metadata 请求。
- 未启用 Kafka 的核心模式在不设置 Kafka 变量时仍然通过。

可观测性模式通过条件：

- `docker compose -f docker-compose.yml -f docker-compose.obs.yml config` 成功。
- Grafana 可在 `127.0.0.1:3000` 访问。
- Jaeger UI 可在 `127.0.0.1:16686` 访问。
- Alloy 启动时没有配置错误。
- Loki 接受来自 Alloy 的日志。

## 19. 未来扩展

仅在需要时添加：

- API 网关，例如 KrakenD
- Kafka 业务生产者和消费者
- 数据库迁移
- JWT 认证
- 限流
- API 版本化
- 复杂 CI/CD 发布流水线
- OpenAPI 生成
- 后台 worker
- 消息队列
- Kubernetes 清单

不要过早将这些能力加入基础项目。
