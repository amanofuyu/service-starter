# 文档总览

本目录按“先入口、再操作、再背景”的顺序组织文档。新增文档时优先放在 `docs/` 下，并在本文档登记定位和适用场景。

## 阅读顺序

第一次接触项目时，建议按下面顺序阅读：

1. `README.md`：了解项目目标、运行模型、接口和快速启动命令。
2. `docs/developer-guide.md`：理解本地开发流程、代码结构、配置加载和测试策略。
3. `docs/operations-guide.md`：理解本地运行、备份、可观测性和常见故障处理。
4. `docs/template-adoption-checklist.md`：在把模板克隆为真实项目时逐项检查。
5. `docs/template-design.md`：需要理解模板设计取舍、非目标和初始验收标准时再读。
6. `docs/ai-agent-guide.md`：使用 AI 编码代理维护项目时阅读。
7. `docs/ai-task-prompts.md`：需要把开发任务整理成 AI 可执行提示时阅读。
8. `docs/adr/`：需要追溯架构决策和后续演进方向时阅读。

## 按角色阅读

- 后端新手：先读 `README.md`，再读 `docs/developer-guide.md`。
- 本地运行和排障人员：先读 `README.md`，再读 `docs/operations-guide.md`。
- 模板克隆和项目初始化人员：重点读 `docs/template-adoption-checklist.md`。
- 模板维护者和架构决策者：先读 `docs/template-design.md`，再读 `docs/adr/`。
- 使用 AI 编码代理的开发者：先读 `AGENTS.md`，再读 `docs/ai-agent-guide.md` 和 `docs/ai-task-prompts.md`。

## 文档定位

| 文档 | 定位 | 主要读者 |
| --- | --- | --- |
| `README.md` | 项目入口，说明项目能做什么、如何启动、有哪些核心接口 | 所有人 |
| `AGENTS.md` | AI 编码代理的仓库级工作规则 | AI 编码代理、维护者 |
| `docs/developer-guide.md` | 开发指南，说明代码结构、配置、扩展方式和测试策略 | 开发者、AI 编码代理 |
| `docs/operations-guide.md` | 运维指南，说明本地运行、备份、可观测性和故障排查 | 开发者、运维人员 |
| `docs/template-adoption-checklist.md` | 模板落地检查清单，说明克隆为真实项目时要改什么、不要改什么 | 项目初始化人员 |
| `docs/template-design.md` | 模板设计说明，记录模板目标、非目标、架构和验收标准 | 维护者、架构决策者 |
| `docs/ai-agent-guide.md` | AI 开发指南，说明 AI 修改常见任务时的操作路径和验证命令 | AI 编码代理、审查者 |
| `docs/ai-task-prompts.md` | AI 任务提示模板，帮助用户把需求、不变量、验收条件和验证命令整理成可执行输入 | 使用 AI 编码代理的开发者、审查者 |
| `docs/adr/` | 架构决策记录，记录重要取舍、约束和演进方向 | 维护者、架构决策者、AI 编码代理 |

## 命名规则

- 面向操作流程的文档使用 `*-guide.md`。
- 面向检查清单的文档使用 `*-checklist.md`。
- 面向设计背景和决策的文档使用 `*-design.md`。
- 架构决策记录放在 `docs/adr/`，文件名使用 `NNNN-short-title.md`。
- 仓库根目录只保留项目入口、代理规则和构建运行所需文件；长期文档默认放入 `docs/`。
