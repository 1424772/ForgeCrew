# ForgeCrew

> **不是让一个 AI 写代码，而是组建一支 AI 软件公司协作交付。**

ForgeCrew 是一个 Go-native、CLI-first 的 **AI Software Company Runtime**。
它不是单一 Coding Agent，而是一个可动态组队的多 Agent 工程运行时。

## 定位

- **动态组队**：每个项目自动分析并推荐专属 AI 软件团队
- **用户可控**：每个 Agent 的模型、权限、工作方式均可自定义
- **Go-native Runtime**：稳定、可分发、跨平台
- **Agent-Computer Interface**：工具调用受控，安全边界清晰
- **Checkpoint & Rollback**：所有修改可审查、可回滚
- **Loop Engineering**：能测试、观察、反思、修正的工程闭环
- **Eval-driven**：通过真实指标优化模型路由、Prompt 和团队配置
- **AGENTS.md 兼容**：可与 Codex、Claude Code、Gemini CLI、Cursor、Cline 等生态协作

## 安装

```bash
# 从源码安装
git clone https://github.com/1424772/ForgeCrew.git
cd ForgeCrew
go build -o forgecrew ./cmd/forgecrew

# 或直接运行
go run ./cmd/forgecrew
```

要求 Go 1.21+。

## 快速开始

### 1. 初始化项目

```bash
forgecrew init
```

在当前目录生成：

```
AGENTS.md                      # 给 AI Agent 看的项目说明
.forgecrew/
├── agents.yaml                # Agent 定义（角色、模型、权限）
├── models.yaml                # 模型定义（provider、cost、capabilities）
├── workflows.yaml             # 工作流定义
├── memory/                    # 任务/项目记忆
├── evals/                     # 评测记录
└── checkpoints/               # 修改检查点
```

重复执行 `init` 不会覆盖已有文件，除非使用 `--force`。

### 2. 扫描项目

```bash
forgecrew scan
forgecrew scan -o json
```

输出项目画像（语言、框架、测试、Docker 等）。

### 3. 获取团队建议

```bash
forgecrew team suggest
```

基于项目类型推荐团队组成（PM、架构师、后端、前端、QA、Reviewer 等）。

### 4. 查看 Agent 和模型

```bash
forgecrew agents list
forgecrew models list
```

### 5. 查看变更

```bash
forgecrew diff
```

### 6. 管理检查点

```bash
forgecrew checkpoint list
```

### 7. 执行任务（dry-run）

```bash
forgecrew task "add login validation" --dry-run
```

打印完整的 Loop Engineering 状态序列（Goal → Plan → Retrieve → Act → Observe → Reflect → Improve → Review → CommitMemory），不做实际修改。

## CLI 命令

| 命令 | 说明 |
|------|------|
| `forgecrew version` | 打印版本号 |
| `forgecrew init` | 初始化 ForgeCrew 配置 |
| `forgecrew scan` | 扫描项目并输出画像 |
| `forgecrew agents list` | 列出所有 Agent |
| `forgecrew models list` | 列出所有模型 |
| `forgecrew team suggest` | 推荐团队配置 |
| `forgecrew diff` | 显示 git diff |
| `forgecrew checkpoint list` | 列出所有检查点 |
| `forgecrew task <goal>` | 通过 Loop Engineering 状态机执行任务 |

## 项目结构

```
cmd/forgecrew/         # CLI 入口
internal/
├── config/            # 配置常量和模板
├── models/            # 模型注册表
├── agents/            # Agent 注册表
├── scanner/           # 项目扫描器
├── teamarchitect/     # 团队架构师规则引擎
├── orchestrator/      # Loop Engineering 状态机
├── aci/               # Agent-Computer Interface
├── gitops/            # Git 受控操作
├── checkpoint/        # 检查点系统
├── memory/            # 记忆存储 (JSONL)
└── eval/              # 评测记录
templates/             # 模板文件
configs/               # 示例配置
docs/                  # 文档
```

## 开发

```bash
# 运行测试
go test ./...

# 代码检查
go vet ./...
```

## 路线图

- **MVP（当前）**：Go CLI 骨架、配置系统、Scanner、Team Architect、Orchestrator 骨架、ACI、Checkpoint、Memory/Eval 基础结构
- **V1**：真实 LLM 调用、Repo Map、Repo RAG、MCP Client/Server、完整 Loop Engineering
- **V2**：Windows App（Tauri）、Graph Memory、自修改沙箱实验

## 许可

MIT
