# 微服务基础项目

面向中小型系统的可复用后端服务基础项目。

默认技术栈刻意保持精简：

- 监听 `8081` 端口的 Go 服务
- PostgreSQL
- Redis
- 可选 Kafka 栈
- 可选本地可观测性栈

应用在 Compose 中的服务名始终为 `service`。

## 适用场景

这个项目适合作为新后端服务的起点，尤其适合需要快速获得以下能力的团队：

- 一个可本地启动、可测试、可复制的 Go 服务骨架
- PostgreSQL 与 Redis 的 Compose 开发环境
- 可按需启用的 Kafka、日志聚合和链路追踪
- 清晰的配置边界，便于后续接入业务代码

它不是生产平台模板。生产部署前仍需要根据目标环境补充 secret 管理、TLS、访问控制、备份保留、监控告警和发布流程。

## 给后端新手的阅读路线

第一次接触这个项目时，不建议一上来阅读所有 Compose 和 Go 代码。推荐按下面顺序理解：

1. 先读本 README，知道项目能启动哪些组件、有哪些接口。
2. 再读 `docs/index.md`，确认每份文档的定位。
3. 然后读 `service/cmd/service/main.go`，确认 Go 程序入口很薄。
4. 接着读 `service/internal/app/app.go`，理解服务启动时依次装配哪些依赖。
5. 然后读 `service/internal/httpserver/router.go` 和 `service/internal/health/health.go`，理解 HTTP 请求怎么被处理。
6. 最后读 `docker-compose.yml`，把环境变量、容器名和连接地址对应起来。

可以把这个项目理解成三层：

```text
Docker Compose 层：提供 postgres、redis、service 等容器
应用装配层：读取配置，创建数据库/Redis/Kafka/HTTP server
HTTP 路由层：处理 /healthz、/readyz、/api/ping 等请求
```

后续新增业务功能时，通常只需要在应用装配层和 HTTP 路由层之间增加自己的业务模块，不需要先改 Docker 或可观测性配置。

## 快速开始

```sh
cp .env.example .env
docker compose up -d
```

验证服务：

```sh
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/api/ping
```

预期的 ping 响应：

```json
{"ok":true,"service":"service"}
```

## 运行模型

默认核心栈只启动 `postgres`、`redis` 和 `service`。PostgreSQL 与 Redis 不暴露到宿主机，只能由 Compose 网络内的服务访问；Go 服务通过 `127.0.0.1:${SERVICE_PORT}` 暴露给本机。

```text
host curl/browser
  |
  v
127.0.0.1:${SERVICE_PORT}
  |
  v
service:8081
  |
  +-- postgres:5432
  |
  `-- redis:6379
```

Kafka 和可观测性组件都放在独立 Compose 覆盖文件中，只有需要相关能力时才启动。

一次 `/api/ping` 请求的大致链路：

```text
curl 127.0.0.1:8081/api/ping
  |
  v
Docker 端口映射 127.0.0.1:${SERVICE_PORT} -> service:8081
  |
  v
http.Server
  |
  v
chi router
  |
  v
/api/ping handler
  |
  v
JSON 响应 {"ok":true,"service":"service"}
```

一次 `/readyz` 请求会多一步依赖检查：handler 会调用 PostgreSQL、Redis 和可选 Kafka 的 `Ping` 方法，任何已配置依赖失败都会返回 `503`。

## 目录结构

```text
.
|-- docker-compose.yml
|-- docker-compose.kafka.yml
|-- docker-compose.obs.yml
|-- .env.example
|-- Makefile
|-- loki-config.yml
|-- alloy-config.alloy
|-- grafana/
|   `-- provisioning/
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
    |-- cmd/service/main.go
    |-- internal/
    `-- migrations/
```

核心目录职责：

- `service/cmd/service`：进程入口，只负责调用应用启动函数。
- `service/internal/app`：装配配置、日志、追踪、数据库、Redis、Kafka 和 HTTP server。
- `service/internal/config`：从环境变量读取配置并做必要校验。
- `service/internal/health`：实现 `/healthz` 和 `/readyz`。
- `service/internal/httpserver`：集中注册 HTTP 路由。
- `docs/`：文档总览、开发、运维、模板设计、克隆检查和 AI 协作说明。

## 文档导航

先从 `docs/index.md` 查看完整文档地图。常用入口：

- `docs/developer-guide.md`：本地开发、代码结构、扩展规则和测试策略。
- `docs/operations-guide.md`：本地运行、备份、可观测性和故障排查。
- `docs/template-adoption-checklist.md`：把模板克隆为真实项目时的检查清单。
- `docs/template-design.md`：基础模板的设计背景、边界和初始验收标准。
- `docs/ai-agent-guide.md`：AI 编码代理执行改动时的操作指南。
- `AGENTS.md`：AI 编码代理必须遵守的仓库规则。

## 环境变量

Compose 必需变量：

```text
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD
REDIS_PASSWORD
SERVICE_PORT
```

Go 服务读取的变量：

```text
APP_ENV
DATABASE_URL
REDIS_URL
SERVICE_PORT
OTEL_SERVICE_NAME
OTEL_TRACES_EXPORTER
OTEL_EXPORTER_OTLP_ENDPOINT
KAFKA_BROKERS
KAFKA_TOPIC_PREFIX
```

在 Compose 中，`DATABASE_URL` 和 `REDIS_URL` 会从数据库与 Redis 变量派生。若在非 Compose 的本地环境运行，请直接设置这两个变量。

配置要点：

- `APP_ENV` 默认是 `development`，主要用于日志和环境标识。
- `SERVICE_PORT` 控制宿主机端口映射；容器内服务仍监听 `8081`。
- `OTEL_TRACES_EXPORTER=none` 时不发送链路追踪。
- `OTEL_TRACES_EXPORTER=otlp` 时使用 `OTEL_EXPORTER_OTLP_ENDPOINT`。
- `KAFKA_BROKERS` 为空时不会创建 Kafka checker，`/readyz` 也不会检查 Kafka。
- `KAFKA_TOPIC_PREFIX` 用于区分本地、预发和生产主题前缀。

不要把真实密码提交到仓库。`.env.example` 只放开发默认值，真实 `.env` 应保持忽略。

新手容易混淆的是：`POSTGRES_DB`、`POSTGRES_USER`、`POSTGRES_PASSWORD` 是给 Compose 创建数据库容器用的；Go 服务真正读取的是 Compose 派生出来的 `DATABASE_URL`。Redis 也是同理，Go 服务读取 `REDIS_URL`，而不是自己重新拼接 Redis 密码和地址。

这样做的好处是：无论服务运行在 Compose、本地终端还是生产环境，只要给出最终连接 URL，应用代码都不用关心这些 URL 是怎么来的。

## HTTP 接口

| 路径 | 用途 | 成功响应 | 失败行为 |
| --- | --- | --- | --- |
| `/healthz` | 进程存活检查 | `200 {"ok":true}` | 当前不检查外部依赖 |
| `/readyz` | 依赖就绪检查 | `200 {"ok":true,"checks":...}` | 任一已配置依赖失败时返回 `503` |
| `/api/ping` | 服务可访问性验证 | `200 {"ok":true,"service":"service"}` | 无额外依赖检查 |

`/readyz` 默认检查 PostgreSQL 和 Redis。只有启用 Kafka 覆盖文件或显式设置 `KAFKA_BROKERS` 后才会检查 Kafka。

## 核心命令

```sh
make up
make logs
make down
make backup
make test
```

备份命令会把自定义格式的 PostgreSQL dump 写入 `backup_data` 卷。

命令含义：

- `make up`：启动核心栈。
- `make logs`：持续查看所有服务日志。
- `make down`：停止核心栈，但不删除命名数据卷。
- `make backup`：运行一次 PostgreSQL 备份容器。
- `make test`：进入 `service` 目录运行 Go 测试。

常用排查命令：

```sh
docker compose ps
docker compose logs -f --tail=200 service
docker compose config
docker compose down --remove-orphans
```

## Kafka

Kafka 是可选组件，不属于默认启动内容。

```sh
make kafka-up
docker compose -f docker-compose.yml -f docker-compose.kafka.yml --profile tools run --rm kafka-init
make kafka-down
```

Kafka 只能在 Compose 网络内通过 `kafka:9092` 访问。只有配置了 `KAFKA_BROKERS` 时，服务才会检查 Kafka 就绪状态。

示例主题由 `kafka-init` 创建：

- `${KAFKA_TOPIC_PREFIX}.service.events`
- `${KAFKA_TOPIC_PREFIX}.service.commands`

## 可观测性

链路追踪和日志聚合都是可选能力。

```sh
make obs-up
make obs-down
```

本地界面：

- Grafana: http://127.0.0.1:3000
- Jaeger: http://127.0.0.1:16686

默认通过 `OTEL_TRACES_EXPORTER=none` 禁用链路追踪。可观测性 Compose 覆盖文件会把它设置为 `otlp`。

Alloy 通过 Docker socket 发现本地容器日志并发送到 Loki。这个配置是本地开发便利能力，不应直接当作生产安全模型。

## 克隆为新项目

只修改面向业务的值：

- `POSTGRES_DB`
- `OTEL_SERVICE_NAME`
- README 标题
- Go 模块路径
- 默认 JSON 响应中的服务名
- 以后如新增备份文件前缀，也一并修改

在真实项目需要之前，不要添加 API 网关、Kubernetes 文件、认证或迁移工具。

推荐克隆步骤：

1. 复制仓库并替换 README 标题和业务描述。
2. 修改 `service/go.mod` 模块路径。
3. 修改 `.env.example` 中的数据库名、服务名和开发密码。
4. 修改 `/api/ping` 默认服务名或改为读取业务服务标识。
5. 运行 `make test` 和 `docker compose config`。
6. 启动核心栈并验证三个 HTTP 接口。

## 验证

本地 Go 验证：

```sh
cd service && go test ./...
```

安装 Docker 后的 Docker 验证：

```sh
docker compose config
docker compose up -d
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/api/ping
docker compose -f docker-compose.yml -f docker-compose.kafka.yml config
docker compose -f docker-compose.yml -f docker-compose.obs.yml config
```

如果只修改文档或注释，至少运行：

```sh
cd service && go test ./...
```
