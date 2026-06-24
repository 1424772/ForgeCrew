# ForgeCrew 架构文档

> 版本：MVP 0.1.0  
> 日期：2026-06-24

## 1. 模块图

```text
┌─────────────────────────────────────────────────┐
│                  CLI (cmd/forgecrew)              │
│  version  init  scan  agents  models  team        │
│  diff  checkpoint                                 │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────┼───────────────────────────────┐
│            Internal Core                         │
│                                                  │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  config  │  │  scanner  │  │ teamarchitect│  │
│  └──────────┘  └───────────┘  └──────────────┘  │
│                                                  │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  models  │  │  agents   │  │ orchestrator │  │
│  │ (registry)│  │ (registry)│  │ (state mach) │  │
│  └──────────┘  └───────────┘  └──────────────┘  │
│                                                  │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │   aci    │  │  gitops   │  │  checkpoint  │  │
│  └──────────┘  └───────────┘  └──────────────┘  │
│                                                  │
│  ┌──────────┐  ┌───────────┐                    │
│  │  memory  │  │   eval    │                    │
│  └──────────┘  └───────────┘                    │
└─────────────────────────────────────────────────┘
```

## 2. CLI 流程

```
用户输入 → Cobra CLI → 命令处理函数
                          ↓
              ┌───────────┴───────────┐
              │                       │
         forgecrew init          forgecrew scan
              │                       │
    config.EnsureDir()         scanner.Scan()
    生成 AGENTS.md               ↓
    生成 .forgecrew/*.yaml    Profile.Format()
              │                       │
              ↓                       ↓
          输出成功信息            输出 YAML/JSON
```

```
forgecrew team suggest:
  scanner.Scan() → Profile
      ↓
  teamarchitect.Suggest(Profile) → TeamConfig
      ↓
  TeamConfig.FormatYAML() → 输出 YAML

forgecrew agents/models list:
  registry.Load(.forgecrew/*.yaml)
      ↓
  registry.List() → 逐行输出

forgecrew diff:
  gitops.ReadDiff(root)
      ↓
  调用 git diff → 输出 diff 或友好错误

forgecrew checkpoint list:
  checkpoint.NewStore().List()
      ↓
  读取 .forgecrew/checkpoints/*.json → 逐条输出
```

## 3. Registry 设计

### Model Registry

```go
type ModelDefinition struct {
    ModelID        string  // yaml key
    Provider       string  // openai, anthropic, zhipu, ...
    Model          string  // gpt-5.5, claude-opus-4-8, ...
    APIKeyEnv      string  // 环境变量名
    Role           string  // reasoning, coding, review, ...
    CostTier       string  // high, medium, low
    SupportsTools  bool
    SupportsVision bool
}

type Registry struct {
    models map[string]ModelDefinition
}
// Methods: Load(path), Get(id), List()
```

### Agent Registry

```go
type AgentDefinition struct {
    AgentID         string
    Name            string
    Description     string
    DefaultModel    string    // 引用 models.yaml 中的 model id
    FallbackModels  []string
    PermissionLevel string    // read_only, patch_only, write_with_approval, ...
    RequireApproval bool
    Tools           []string  // 可用 ACI 动作
    Modes           []string  // code, debug, test, review, ...
}

type Registry struct {
    agents map[string]AgentDefinition
}
// Methods: Load(path), Get(id), List()
```

### 验证规则

- Agent 必须: agent_id, name, default_model, 有效 permission_level
- Model 必须: provider, model, api_key_env, 有效 cost_tier (high/medium/low)
- 加载时对缺字段、引用不存在的模型等错误给出清晰错误信息

## 4. Scanner 设计

```
scanner.Scan(root) → Profile
  ├── detectFromFiles: 检查 go.mod, package.json, pyproject.toml 等
  ├── detectTests: 检查 *_test.go, *.test.ts, tests/ 等
  ├── detectDocker: 检查 Dockerfile, docker-compose.yml 等
  ├── detectAgentsMD: 检查 AGENTS.md
  ├── detectGitHubActions: 检查 .github/workflows/
  └── guessProjectType:
        go+cobra → cli_tool
        go+gin/fiber → backend_api
        ts only → frontend_app
        ts+go/python → fullstack_saas
        python+tests → agent_app
        docker only → devops
```

## 5. Team Architect 设计

```
teamarchitect.Suggest(Profile) → TeamConfig
  ├── cli_tool → pm, architect, repo_analyst, backend_engineer, qa_engineer, code_reviewer
  ├── backend_api → pm, architect, repo_analyst, backend_engineer, qa_engineer, code_reviewer
  ├── frontend_app → pm, architect, repo_analyst, frontend_engineer, qa_engineer, code_reviewer
  ├── fullstack_saas → pm, architect, repo_analyst, backend_engineer, frontend_engineer, qa_engineer, code_reviewer
  ├── agent_app → pm, architect, repo_analyst, backend_engineer, qa_engineer, code_reviewer + memory_agent (optional)
  └── generic → pm, repo_analyst, backend_engineer, qa_engineer, code_reviewer
```

**权限默认保守：**

| Agent Role | Permission Level |
|---|---|
| team_architect | read_only |
| pm | read_only |
| architect | read_only |
| repo_analyst | read_only |
| code_reviewer | read_only |
| backend_engineer | patch_only |
| frontend_engineer | patch_only |
| qa_engineer | patch_only |
| devops | write_with_approval |

## 6. Orchestrator 状态机

### Loop Engineering 循环

```text
Goal → Plan → Retrieve → Act → Observe → Reflect → Improve → Review → CommitMemory
  ↑                                                                    │
  └─────────────────── 循环迭代 (max 3) ───────────────────────────────┘
```

### 状态机接口

```go
type StateMachine struct {
    CurrentStep Step
    Iteration   int
    MaxIter     int
    DryRun      bool
    History     []StateRecord
}
// Methods: New(maxIter, dryRun), Next(), Execute(taskID), Current(), IsDone()
```

- MVP 仅支持 dry-run 模式（不接真实模型）
- 第一轮从 Goal 开始，后续迭代从 Plan 开始（跳过 Goal）
- 最多迭代 MaxIter 次

## 7. Agent-Computer Interface (ACI) 安全边界

ACI 层是关键安全边界。Agent **不直接**访问 shell、文件系统、git。
所有操作必须通过 ACI 受控接口：

| 动作 | 风险 | MVP 状态 |
|------|------|---------|
| ReadFile | 低 | ✅ 已实现 |
| SearchCode | 低 | ⚠️ ErrNotImplemented |
| ReadGitDiff | 低 | ✅ 已实现 (via gitops) |
| CreateCheckpoint | 低 | ✅ 已实现 |
| GeneratePatch | 中 | ⚠️ ErrNotImplemented |
| RunTest | 中 | ⚠️ ErrNotImplemented |
| Rollback | 高 | ⚠️ ErrNotImplemented |

```go
// ACI 接口模式
func ReadFile(root, path string) (string, error)
func SearchCode(root, pattern string) (string, error)
func GeneratePatch(root string, changes []FileChange) (string, error)
func RunTest(root string) (string, error)
func ReadGitDiff(root string) (string, error)
func Rollback(checkpointID string) error
```

后续版本将实现完整 ACI 和权限审批策略。

## 8. Checkpoint 系统

```
CreateCheckpoint → .forgecrew/checkpoints/{id}.json
  ├── id
  ├── timestamp
  ├── task_id
  ├── agent_id
  ├── model_id
  ├── changed_files
  └── git_hash_before

checkpoint list → 读取所有 JSON 文件，按时间排序输出
```

未来将支持：
- `forgecrew rollback <checkpoint_id>`
- 部分接受/拒绝修改
- 查看每个 Agent 的修改历史

## 9. Memory / Eval 设计

### Memory (JSONL)

```
.forgecrew/memory/
├── task.jsonl     # 任务记忆
└── project.jsonl  # 项目记忆

Entry: {id, type, scope, content, tags[], timestamp}
```

### Eval (JSON)

```
.forgecrew/evals/
└── eval_{timestamp}.json

Record: {
  id, task_id, success, test_passed, review_passed,
  rounds, cost_estimate, human_intervention_count, timestamp
}
```

## 10. 后续路线

### V1 (计划中)
- 真实 LLM 调用集成
- Repo Map Engine (tree-sitter + ripgrep + 符号图)
- Repo RAG (SQLite FTS5 + 向量检索)
- 完整 ACI 实现 (SearchCode, GeneratePatch, RunTest, Rollback)
- MCP Client/Server
- 完整 Loop Engineering (对接真实 Agent)
- AGENTS.md 深度兼容

### V2 (远期)
- Windows Desktop App (Tauri + React)
- Graph Memory / Decision Graph
- 自修改沙箱实验 (用户审批 + 安全边界)
- 多人协作支持

## 11. 架构决策记录 (ADR)

### ADR-001: Go-native Core Runtime

**决定**: 使用 Go 构建核心 Runtime。  
**原因**: 单二进制分发、跨平台、并发能力强、适合 CLI 工具。  
**结果**: 不使用 Python/TypeScript 做核心调度。

### ADR-002: ACI 层作为安全边界

**决定**: 不把 shell、git、文件系统直接暴露给 Agent，通过 ACI 受控接口。  
**原因**: 防止 Agent 执行危险操作（删除文件、读取密钥、直接 push）。  
**结果**: 所有 Agent 工具调用必须经过 ACI 权限检查。

### ADR-003: JSONL 做 MVP 存储

**决定**: MVP 使用 JSONL 做 memory，JSON 文件做 checkpoint/eval。  
**原因**: 避免 SQLite 依赖，简化 MVP 实现。后续可迁移。  
**结果**: 简单、透明、可手动查看和编辑。

### ADR-004: 配置目录 .forgecrew

**决定**: 统一使用 .forgecrew 作为配置目录。  
**原因**: 产品名统一为 ForgeCrew，CLI 命令统一为 forgecrew。  
**结果**: 与 AGENTS.md 生态兼容，项目级配置清晰。
