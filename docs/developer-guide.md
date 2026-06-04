# 开发指南

## 前置条件

- Go 1.26.x
- 用于容器验证的 Docker 与 Compose 插件
- 用于执行项目统一命令的 Make

当前固定的构建镜像是 `golang:1.26.4-alpine3.23`。

如果在 WSL/nvm 环境中运行 Node 相关命令，先显式加载 nvm：

```sh
export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"
```

当前项目没有 Node 构建步骤；这条规则主要用于以后新增前端、脚本工具或代码生成器时避免非交互 shell 找不到 Node。

## 本地 Go 工作流

运行测试：

```sh
cd service && go test ./...
```

格式化代码：

```sh
cd service && gofmt -w .
```

通过直接提供连接 URL 在 Compose 外运行：

```sh
APP_ENV=development \
DATABASE_URL='postgres://service:service_dev_password@127.0.0.1:5432/service?sslmode=disable' \
REDIS_URL='redis://:redis_dev_password@127.0.0.1:6379/0' \
SERVICE_PORT=8081 \
go run ./cmd/service
```

默认 Compose 栈不会将 PostgreSQL 或 Redis 暴露到宿主机，因此非 Compose 运行需要单独可用的依赖，或临时本地覆盖配置。

## 启动流程

`cmd/service/main.go` 只负责调用 `app.Run`。实际启动顺序集中在 `internal/app`：

1. 读取并校验环境变量。
2. 初始化 JSON 日志。
3. 根据 `OTEL_TRACES_EXPORTER` 初始化或跳过 OpenTelemetry。
4. 创建 PostgreSQL 连接池。
5. 创建 Redis client。
6. 在 `KAFKA_BROKERS` 非空时创建 Kafka checker。
7. 注册 HTTP 路由并启动 server。
8. 收到 `SIGINT` 或 `SIGTERM` 后执行 10 秒优雅关闭。

这种结构让进程入口保持简单，也让配置、依赖装配和路由注册各自有明确边界。

## 请求处理流程

后端新手可以把一个 HTTP 请求拆成下面几步：

```text
客户端请求
  |
  v
net/http.Server 接收请求
  |
  v
chi 路由根据路径和方法选择 handler
  |
  v
handler 读取请求、调用依赖、写出 JSON 响应
```

当前项目里，`httpserver.NewRouter` 是路由总入口：

- `GET /healthz` 交给 `health.Handler.Healthz`。
- `GET /readyz` 交给 `health.Handler.Readyz`。
- `GET /api/ping` 使用一个很小的匿名 handler 返回示例 JSON。

新增接口时，先从 `httpserver.NewRouter` 看起。不要直接在 `main.go` 里写路由，因为入口文件越简单，后续越容易测试和迁移。

## 配置加载流程

`config.Load` 做三件事：

1. 从环境变量读取字符串。
2. 给可选项设置默认值。
3. 对必须存在的配置做校验。

如果 `DATABASE_URL` 或 `REDIS_URL` 为空，服务会在启动阶段失败。这比等到第一次业务请求才失败更容易排查。

`KAFKA_BROKERS` 使用逗号分隔，例如：

```text
kafka-1:9092,kafka-2:9092,kafka-3:9092
```

代码会自动去掉多余空格，并忽略空项。所以 `kafka:9092, ` 这种输入不会生成一个空 broker。

## 依赖检查流程

`/healthz` 和 `/readyz` 的职责不同：

- `/healthz` 只回答“这个进程还能不能响应 HTTP”。
- `/readyz` 回答“这个进程现在能不能处理需要依赖外部资源的请求”。

`/readyz` 通过 `health.Checker` 接口检查依赖。只要某个对象有下面这个方法，就可以被纳入 readiness：

```go
Ping(context.Context) error
```

这种接口很小，原因是 readiness 不需要知道依赖的全部能力，只需要知道“能不能 ping 通”。PostgreSQL 连接池、Redis client 和 Kafka checker 都被适配成了这个形状。

## 服务结构

- `cmd/service`：进程入口
- `internal/app`：启动、生命周期和依赖装配
- `internal/config`：环境变量加载与校验
- `internal/database`：PostgreSQL 连接池创建
- `internal/redis`：Redis client 创建和 readiness adapter
- `internal/kafka`：可选 Kafka readiness adapter
- `internal/httpserver`：路由注册
- `internal/health`：`/healthz` 和 `/readyz`
- `internal/observability`：可选 OpenTelemetry tracing

## 架构演进方向

当前项目是轻量的分层装配式模板，不是严格洋葱模型，也不是完整的面向切片架构。这样设计是有意为之：模板阶段真实业务很少，过早引入 Clean Architecture、DDD 分层、repository 框架或代码生成，会让新项目先承担不必要的结构成本。

后续新增业务时，优先保持 `internal/app` 负责装配、`internal/httpserver` 负责路由和 middleware、基础设施包负责连接和适配。不要把复杂业务逻辑继续堆进路由闭包。

当项目出现多个真实业务能力，并且每个能力的 handler、usecase、store 或 model 开始成组变化时，再逐步演进为按业务切片组织：

```text
internal/httpserver       路由注册和 middleware
internal/<feature>        业务切片，包含 handler/usecase/store/model
internal/database         数据库连接和通用基础设施
internal/health           横切健康检查
```

架构演进的取舍记录放在 `docs/adr/`。修改项目结构、引入新的业务组织方式或添加会影响长期边界的依赖前，先补一条 ADR，说明背景、决策、后果和替代方案。

## 代码扩展规则

新增业务能力时优先保持以下边界：

- 路由注册放在 `internal/httpserver`，不要散落到进程入口。
- 配置字段统一放在 `internal/config.Config`，并在 `Load` 中校验必需项。
- 需要进入 `/readyz` 的外部依赖实现 `health.Checker`。
- 数据库、Redis、Kafka 的具体 client 封装在各自 package 中，不直接泄露到 HTTP handler。
- handler 中只处理协议层逻辑，复杂业务逻辑放到独立 service/usecase package。

新增 API 时建议先补 handler 测试，再接入路由。处理器测试应使用假的依赖对象，避免把单元测试变成 Compose 集成测试。

一个简单的新增接口路径可以按这个顺序做：

1. 在合适的 `internal/<业务名>` package 中写业务函数。
2. 给业务函数写单元测试。
3. 在 `internal/httpserver` 注册路由。
4. 给 handler 写 `httptest` 测试。
5. 如果接口依赖新的外部服务，补 adapter 并决定是否加入 `/readyz`。

不要把数据库查询、Redis 操作和复杂业务判断都塞进路由闭包。短期看写得快，长期会让测试和排错都变困难。

## 配置策略

配置来自环境变量。Compose 中可以从基础变量派生连接串，但 Go 服务只读取最终连接 URL：

- `DATABASE_URL`
- `REDIS_URL`

不要在 Go 代码里重新读取 `POSTGRES_DB`、`POSTGRES_USER`、`POSTGRES_PASSWORD` 或 `REDIS_PASSWORD` 来拼接连接串。这样可以让 Compose、本地直连、测试环境和生产环境使用同一套应用配置入口。

## 依赖策略

添加或升级 Go modules 或 Docker image tags 前，先联网查询当前上游版本，并固定精确版本。

依赖选择原则：

- 优先选择成熟、维护活跃、API 稳定的库。
- 不为了“以后可能用到”提前加入框架。
- 新增依赖后同步说明用途，避免模板项目变成隐式依赖集合。
- Docker image 不使用 `latest`，避免不同机器上拉到不一致版本。

## 测试

测试应避免依赖正在运行的容器。处理器测试使用假的 readiness 检查器，将真实 Postgres、Redis、Kafka、Loki 和 Jaeger 检查留给 Compose 验收测试。

推荐验证层级：

```text
go test ./...
  |
  +-- 配置加载
  +-- 路由响应
  `-- readiness 分支

docker compose config
  |
  `-- Compose 变量、覆盖文件和网络配置

docker compose up -d + curl
  |
  `-- 本地端到端冒烟验证
```

## 常见开发问题

### `/readyz` 返回 503

先查看响应中的 `checks` 字段，确认是 `postgres`、`redis` 还是 `kafka` 失败。随后查看对应容器日志：

```sh
docker compose logs -f --tail=200 postgres
docker compose logs -f --tail=200 redis
docker compose logs -f --tail=200 service
```

如果失败项是 Kafka，确认是否使用 Kafka 覆盖文件启动，以及 `KAFKA_BROKERS` 是否指向 Compose 网络内的 `kafka:9092`。

### 本地 `go run` 连不上数据库

默认 Compose 不把 PostgreSQL 和 Redis 映射到宿主机。要么使用 Compose 运行服务，要么单独准备本地 PostgreSQL/Redis，并显式提供 `DATABASE_URL` 和 `REDIS_URL`。

### 修改 Compose 后服务仍异常

先检查最终配置：

```sh
docker compose config
```

如果覆盖文件组合有变化，使用对应组合执行 `config`，例如：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml config
```

### 看不懂 `depends_on` 和 readiness 的区别

Compose 的 `depends_on` 只影响容器启动顺序，不能代表服务永远可用。比如 PostgreSQL 容器启动并 healthy 后，运行过程中仍可能因为配置、网络或资源问题不可用。

所以应用内部仍然需要 `/readyz`。它是运行时检查，不是启动时检查。

### 为什么测试里用 fake checker

单元测试应该快速、稳定、可重复。如果 handler 测试每次都要启动真实 PostgreSQL 或 Redis，那么测试会变慢，也更容易受本地环境影响。

测试里使用 fake checker，可以专注验证 handler 在依赖成功或失败时的 HTTP 状态码和响应格式。真实依赖连通性留给 Compose 冒烟测试。
