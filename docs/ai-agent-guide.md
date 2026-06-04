# AI 开发指南

本文档面向使用 AI 编码代理维护本模板或基于本模板派生业务项目的场景。它补充 `AGENTS.md`，用于说明常见改动应该怎样落地和验证。

## 默认工作流

1. 先阅读 `README.md`、`docs/index.md`、`docs/developer-guide.md`、相关 Go package 和测试。
2. 明确本次改动属于文档、配置、Go 行为、Compose 拓扑还是依赖升级。
3. 保持改动范围最小，只修改与目标直接相关的文件。
4. 修改后运行对应验证命令。
5. 最终汇报中写清楚改了什么、跑了什么、还有什么没验证。

常规验证入口：

```bash
make check
```

## 新增 HTTP 接口

推荐步骤：

1. 在合适的 `internal/<业务名>` package 中实现业务逻辑。
2. 为业务逻辑补单元测试。
3. 在 `internal/httpserver` 注册路由。
4. 用 `httptest` 覆盖 HTTP 状态码、响应 JSON 和错误分支。
5. 运行 `make fmt-check && make test && make vet`。

不要把数据库查询、Redis 操作和复杂业务判断直接写进路由闭包。

## 新增配置项

推荐步骤：

1. 在 `service/internal/config.Config` 中新增字段。
2. 在 `config.Load` 中读取、设置默认值并校验必填项。
3. 更新 `.env.example`、`README.md` 和相关 docs。
4. 如果 Compose 需要传入该变量，同步更新 Compose 文件。
5. 补配置加载测试，运行 `make test` 和 `make compose-check`。

应用代码应读取最终配置值。数据库和 Redis 仍只通过 `DATABASE_URL` 与 `REDIS_URL` 进入 Go 服务。

## 新增外部依赖

添加 Go module、Docker 镜像或其他工具前必须联网查询当前最新稳定版本和兼容性要求。确认后再固定具体版本。

推荐记录：

- 为什么需要这个依赖；
- 可选替代方案是什么；
- 为什么不使用标准库或已有依赖；
- 固定的版本或镜像 tag；
- 验证命令和结果。

不要因为“以后可能用到”提前加入依赖。

## 新增 readiness 检查

推荐步骤：

1. 为外部依赖创建小型 adapter。
2. adapter 暴露 `Ping(context.Context) error`。
3. 在 `internal/app` 装配依赖。
4. 在 `health.Dependencies` 中加入可选 checker。
5. 补 `/readyz` 成功、失败和未启用分支测试。

不要让 `/healthz` 检查外部依赖。`/healthz` 只表示进程还能响应 HTTP。

## 修改 Compose

修改 Compose 后至少运行：

```bash
make compose-check
```

`make compose-check` 使用 `.env.example` 展开 Compose 配置，用于验证模板仓库的默认配置是否完整。它不启动容器，也不证明本机真实 `.env`、端口占用或依赖运行态都可用。需要验证本机真实 `.env` 时，可以直接运行 `docker compose config`。

如果改动影响服务启动、端口、环境变量或依赖健康检查，还需要执行本地冒烟验证：

```bash
set -a
. ./.env
set +a
HOST_PORT="${SERVICE_PORT:-8081}"
docker compose up -d --build
curl "http://127.0.0.1:${HOST_PORT}/healthz"
curl "http://127.0.0.1:${HOST_PORT}/readyz"
curl "http://127.0.0.1:${HOST_PORT}/api/ping"
```

`SERVICE_PORT` 控制宿主机端口映射；容器内服务仍固定监听 `8081`。如果 `.env` 中写的是 `SERVICE_PORT=18081`，冒烟验证也必须访问 `18081`。

在 WSL 中，如果当前非交互进程没有继承 `docker` 组，但 `getent group docker` 已显示当前用户属于该组，可以用下面的形式临时执行 Docker 命令：

```bash
sg docker -c 'docker compose ps'
```

这只解决当前 shell 的 Unix socket 组权限问题。若用户本身不在 `docker` 组，或 Docker Desktop/daemon 未运行，应先修复 Docker 环境，不要把权限问题归因于项目配置。

默认不要将 PostgreSQL、Redis 或 Kafka 暴露到宿主机。确需暴露时绑定到 `127.0.0.1` 并更新文档。

## 暂不默认加入的能力

除非业务项目已有明确需求，否则不要默认加入：

- API 网关；
- Kubernetes、Helm 或 Kustomize；
- 认证和权限框架；
- OpenAPI 代码生成；
- 数据库迁移工具；
- 多服务 monorepo 结构；
- 复杂 CI/CD 发布流程。

这些能力会显著增加模板假设。作为 AI 主开发基座，优先保证结构清楚、验证便宜、失败可定位。
