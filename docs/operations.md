# 运维

本项目是本地开发基础项目，不是生产运维平台。

## 核心栈

启动：

```sh
docker compose up -d
```

停止：

```sh
docker compose down
```

日志：

```sh
docker compose logs -f --tail=200
```

默认只有 Go 服务暴露到宿主机，并绑定到 `127.0.0.1:${SERVICE_PORT}`。PostgreSQL 和 Redis 只能通过 Compose 网络访问。

查看状态：

```sh
docker compose ps
```

重建服务镜像：

```sh
docker compose build service
docker compose up -d service
```

清理已停止容器但保留数据卷：

```sh
docker compose down --remove-orphans
```

## 健康状态

- `/healthz`：仅检查进程健康状态
- `/readyz`：检查 PostgreSQL 和 Redis；设置 `KAFKA_BROKERS` 时也检查 Kafka
- `/api/ping`：简单的服务可访问性响应

建议使用方式：

- 容器编排或进程监控使用 `/healthz` 判断进程是否仍在响应。
- 流量入口或部署检查使用 `/readyz` 判断服务是否可以接收业务请求。
- 人工验证和示例脚本使用 `/api/ping`。

`/readyz` 的失败响应会保留已检查依赖的状态，便于定位问题：

```json
{"checks":{"postgres":"error","redis":"ok"},"ok":false}
```

后端新手可以按下面方式理解：

- `healthz` 像是在问“程序有没有活着”。
- `readyz` 像是在问“程序能不能开始接业务流量”。
- `ping` 像是人工检查接口路由和 JSON 响应是否正常。

如果 `/healthz` 正常但 `/readyz` 失败，通常说明 Go 进程还活着，但数据库、Redis 或 Kafka 其中一个依赖不可用。

## 备份

运行：

```sh
docker compose --profile tools run --rm pg-backup
```

备份会写入 `backup_data` 卷。真实部署中需要增加保留策略和异机存储。

查看备份卷内容可以临时运行 PostgreSQL 镜像：

```sh
docker compose --profile tools run --rm pg-backup sh -c 'ls -lh /backup'
```

恢复数据不在默认模板中实现。真实项目需要根据环境补充恢复演练、权限控制和保留策略。

## Kafka 栈

启动：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml up -d
```

引导示例主题：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml --profile tools run --rm kafka-init
```

停止：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml down
```

Kafka 以单节点 KRaft 模式运行。默认不暴露到宿主机。

Kafka 覆盖文件会给 `service` 注入 `KAFKA_BROKERS=kafka:9092`。因此启用后，`/readyz` 会把 Kafka 纳入依赖检查；未启用时不会检查 Kafka。

查看主题：

```sh
docker compose -f docker-compose.yml -f docker-compose.kafka.yml run --rm kafka-init \
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --list
```

## 可观测性栈

启动：

```sh
docker compose -f docker-compose.yml -f docker-compose.obs.yml up -d
```

停止：

```sh
docker compose -f docker-compose.yml -f docker-compose.obs.yml down
```

本地界面：

- Grafana: `http://127.0.0.1:3000`
- Jaeger: `http://127.0.0.1:16686`

Alloy 使用 Docker socket 发现本地容器日志，并将日志发送到 Loki。应把它视为开发便利能力，而不是生产安全模型。

追踪启用逻辑：

- 默认核心栈设置 `OTEL_TRACES_EXPORTER=none`，不创建 OTLP exporter。
- 可观测性覆盖文件设置 `OTEL_TRACES_EXPORTER=otlp` 和 `OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318`。
- 服务关闭时会给 tracing provider 5 秒时间 flush。

日志聚合逻辑：

- Docker 容器仍使用 `json-file` 日志驱动。
- Alloy 读取 Docker socket 中的容器日志元数据。
- Loki 保存日志，Grafana 预置 Loki 数据源。

如果 Grafana 中看不到日志，先检查：

```sh
docker compose -f docker-compose.yml -f docker-compose.obs.yml logs -f --tail=200 alloy
docker compose -f docker-compose.yml -f docker-compose.obs.yml logs -f --tail=200 loki
```

## 数据卷

默认命名卷：

- `postgres_data`：PostgreSQL 数据
- `redis_data`：Redis AOF 数据
- `backup_data`：手动备份结果
- `kafka_data`：Kafka 数据，仅 Kafka 覆盖文件创建
- `loki_data`、`grafana_data`：可观测性数据，仅可观测性覆盖文件创建

执行 `docker compose down` 不会删除命名卷。只有明确执行 `docker compose down -v` 才会删除数据卷，开发时使用前应确认不再需要这些数据。

## 故障排查

### 服务启动失败

先查看最终 Compose 配置和服务日志：

```sh
docker compose config
docker compose logs -f --tail=200 service
```

常见原因：

- `.env` 缺少 `POSTGRES_DB`、`POSTGRES_USER`、`POSTGRES_PASSWORD`、`REDIS_PASSWORD` 或 `SERVICE_PORT`。
- Redis 密码与 `REDIS_URL` 派生值不一致。
- 宿主机 `SERVICE_PORT` 已被占用。
- 启用了 Kafka 覆盖文件但 Kafka 尚未 healthy。

### 数据库或 Redis 不健康

查看对应日志：

```sh
docker compose logs -f --tail=200 postgres
docker compose logs -f --tail=200 redis
```

如果是首次启动，先等待健康检查完成。若是配置改动后失败，检查 `.env` 和命名卷中的旧数据是否仍使用旧密码或旧数据库名。

### 端口被占用

如果 `docker compose up -d` 提示 `8081` 端口无法绑定，说明宿主机已有进程占用该端口。可以修改 `.env` 中的 `SERVICE_PORT`，例如：

```text
SERVICE_PORT=18081
```

然后重新启动：

```sh
docker compose up -d
curl http://127.0.0.1:18081/api/ping
```

注意容器内服务仍然监听 `8081`，这里修改的是宿主机映射端口。

### 改了 `.env` 但没有生效

Compose 会在创建或更新容器时读取环境变量。修改 `.env` 后，重新执行：

```sh
docker compose up -d
```

如果仍不确定最终值，可以用：

```sh
docker compose config
```

查看 Compose 展开后的配置。

## 生产说明

投入生产前，需要补充真实 secret 管理、TLS、备份保留、监控告警、访问控制以及面向具体部署环境的加固。

不要直接把当前 Compose 文件当作生产部署声明使用。它优先服务本地开发：默认开发密码、匿名 Grafana、Docker socket 日志采集和单节点 Kafka 都不满足生产安全与可靠性要求。
