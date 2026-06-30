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

### 脚本安装（推荐）

**macOS / Linux：**
```bash
curl -fsSL https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.sh | sh
```

**Windows PowerShell：**
```powershell
iwr https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.ps1 -useb | iex
```

**Windows CMD：**
```cmd
powershell -ExecutionPolicy Bypass -Command "iwr https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.ps1 -useb | iex"
```

安装后可直接运行：
```cmd
forgecrew
```

脚本安装到用户目录（`~/.forgecrew/bin/` 或 `%USERPROFILE%\.forgecrew\bin\`），无需管理员权限。
如果 GitHub Releases 尚未发布，脚本会提示如何通过 `go build` 本地构建。

### 手动安装

```bash
# go install
go install github.com/1424772/ForgeCrew/cmd/forgecrew@latest

# 或从源码构建
git clone https://github.com/1424772/ForgeCrew.git
cd ForgeCrew
go build -o forgecrew ./cmd/forgecrew
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
├── checkpoints/               # 修改检查点
└── runs/                      # 运行/任务状态
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

### 7. 执行任务

```bash
# Dry-run（默认）：打印状态序列，不做实际调用
forgecrew task "add login validation"

# 显式 dry-run
forgecrew task "add login validation" --dry-run

# 真实 LLM 调用：生成执行计划（不写文件，不执行 shell）
forgecrew task "add login validation" --execute
```

`--execute` 需要先在 `models.yaml` 中配置模型并通过环境变量设置 API Key。例如：
```bash
export OPENAI_API_KEY="sk-..."
forgecrew task "为项目添加登录功能" --execute
```

### 8. Agent 状态与 CTO 审查

ForgeCrew 支持记录 agent 工作状态并让 CTO 通过结构化审查 workflow。

```bash
# 提交 agent handoff（自动收集 git diff，创建 task 状态和 handoff 文件）
forgecrew handoff submit backend-001 \
  --agent backend_engineer \
  --goal "实现 provider 层" \
  --summary "新增 internal/llm，task --execute 可调用 OpenAI-compatible API"

# 列出所有 run
forgecrew runs list

# 查看 run 详情和 task 状态
forgecrew runs status <run_id>

# CTO 审查 task（规则审计，不接 LLM）
forgecrew cto review backend-001
```

**CTO review 当前是本地规则审计**，不接 LLM。审查依据：
- task 状态必须为 `ready_for_review`
- handoff 文件必须存在
- changed_files 不能为空
- `go test ./...` 和 `go vet ./...` 必须通过

所有状态文件存储在 `.forgecrew/runs/<run_id>/` 下。

### 当前 CTO review 限制
- 不接 LLM，不做代码质量语义分析
- 不自动 git add/commit/push
- 不自动修改代码
- 仅做事实审计：状态、handoff、git diff、测试结果

## CLI 命令

| 命令 | 说明 |
|------|------|
| `forgecrew` | 进入交互模式 |
| `forgecrew version` | 打印版本号 |
| `forgecrew init` | 初始化 ForgeCrew 配置 |
| `forgecrew scan` | 扫描项目并输出画像 |
| `forgecrew agents list` | 列出所有 Agent |
| `forgecrew models list` | 列出所有模型 |
| `forgecrew team suggest` | 推荐团队配置 |
| `forgecrew diff` | 显示 git diff |
| `forgecrew checkpoint list` | 列出所有检查点 |
| `forgecrew task <goal>` | 通过 Loop Engineering 状态机执行任务 |
| `forgecrew runs list` | 列出所有 run |
| `forgecrew runs status <id>` | 查看 run 和 task 状态 |
| `forgecrew handoff submit <id>` | 提交 agent handoff |
| `forgecrew cto review <id>` | CTO 审查 task |
| `forgecrew lang` | 切换语言 (zh/en) |

## 国际化

ForgeCrew 支持中文和英文界面语言。默认跟随系统 locale，也可手动切换：

```bash
forgecrew lang zh   # 切换到中文
forgecrew lang en   # 切换到英文
```

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
├── provider/          # LLM Provider 客户端（OpenAI-compatible API）
├── aci/               # Agent-Computer Interface
├── gitops/            # Git 受控操作
├── checkpoint/        # 检查点系统
├── runs/              # 运行/任务状态管理
├── memory/            # 记忆存储 (JSONL)
├── eval/              # 评测记录
├── i18n/              # 国际化支持
└── settings/          # 设置管理
templates/             # 模板文件
configs/               # 示例配置
scripts/               # 安装脚本
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

- **MVP（当前）**：Go CLI 骨架、配置系统、Scanner、Team Architect、Orchestrator 骨架、ACI、Checkpoint、Memory/Eval 基础结构、Provider 集成、Runs 管理、CTO Review
- **V1**：真实 LLM 调用增强、Repo Map、Repo RAG、MCP Client/Server、完整 Loop Engineering
- **V2**：Windows App（Tauri）、Graph Memory、自修改沙箱实验

## 许可

MIT
