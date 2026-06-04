# 0001. 服务架构采用轻量分层装配并预留业务切片演进

## 状态

已接受

## 背景

本仓库是轻量 Go 微服务基础模板，当前只有健康检查、示例 ping、PostgreSQL、Redis、可选 Kafka 和可观测性装配。此阶段业务规则很少，主要目标是让新服务快速启动、容易理解、容易测试。

如果现在强行套用完整洋葱模型、Clean Architecture、DDD 分层或 repository 框架，目录会先于业务复杂度膨胀。相反，如果所有新增业务都继续写进 `internal/httpserver`，后续又会让路由层承载过多业务逻辑。

## 决策

当前保持轻量分层装配式结构：

```text
cmd/service
  -> internal/app
      -> internal/config
      -> internal/observability
      -> internal/database / internal/redis / internal/kafka
      -> internal/httpserver
          -> internal/health
```

短期内继续让 `internal/app` 负责启动装配，`internal/httpserver` 负责路由和 middleware，基础设施包负责连接、客户端和 readiness adapter。

当项目出现多个真实业务能力，并且每个能力的 handler、usecase、store 或 model 开始成组变化时，再按业务切片演进：

```text
internal/httpserver       路由注册和 middleware
internal/<feature>        业务切片，包含 handler/usecase/store/model
internal/database         数据库连接和通用基础设施
internal/health           横切健康检查
```

业务切片内部可以根据复杂度拆分 handler、usecase、store 和 model；简单能力可以先保持更少文件，不为形式完整而拆层。

## 后果

这个决策让模板阶段保持低认知成本，同时给真实业务增长留出演进路径。新增业务时，开发者应避免把复杂业务逻辑写在路由闭包里，而应放进对应业务 package。

这个项目暂时不承诺严格洋葱模型：没有强制 domain/usecase/repository 内外圈，也不要求所有依赖都倒置到接口。只有当业务复杂度证明需要时，才增加更明确的领域层或接口边界。

## 替代方案

- 严格洋葱模型：当前业务太少，会引入过多目录和接口，模板使用者需要先理解架构仪式。
- 完整面向切片架构：适合已有多个业务能力的服务，但当前模板只有基础能力，立即切片会制造空目录或示例噪音。
- 继续按技术层无限扩展：短期简单，但真实业务增多后容易让 handler、store 和 model 分散在不同技术目录，跨文件跳转成本变高。
