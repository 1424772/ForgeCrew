# AI Software Company Runtime 项目企划书 V1.0

> 版本：V1.0  
> 日期：2026-06-24  
> 目标用途：交给 Codex / Coding Agent / 后续开发团队作为项目开发输入文档  
> 项目形态：CLI 优先，后续扩展 Windows Desktop App  
> 底层方向：Go-native Core Runtime + Python/TypeScript Plugin Layer  
> 核心定位：动态组队的 AI 软件公司 Runtime，而不是单一 Coding Agent

---

## 0. 版本说明

### V1.0 相比 v0.2 的主要升级

1. **底层架构正式确定为 Go-native Core Runtime**  
   Go 负责 CLI、TUI、任务调度、权限控制、沙箱、Git、文件系统、MCP、Checkpoint、Eval Harness 和本地数据库。

2. **正式加入 DeepAgents-Inspired Agent Harness 思想**  
   借鉴 deepagents 的 sub-agents、filesystem、shell、context management、persistent memory、skills、MCP、human-in-the-loop，但不绑定 LangChain 技术栈。

3. **正式加入开源项目思想归档**  
   吸收 OpenHands、SWE-agent、mini-SWE-agent、Aider、Cline、Roo Code、Gemini CLI、AGENTS.md、Mem0、Letta/MemGPT、Graphiti/Zep、agentmemory、Eino、MCP Go SDK、SWE-bench、OpenHands Benchmarks / HAL Harness 等项目的工程思想。

4. **新增 Agent-Computer Interface Layer**  
   不把 shell、git、文件系统直接裸露给模型，而是包装为 inspect_project、search_code、open_file、edit_file、run_test、generate_patch、rollback 等受控动作。

5. **新增 Repo Map Engine**  
   借鉴 Aider 的 repo map 思想，结合 tree-sitter、ripgrep、import graph、symbol graph、dependency graph、BM25、向量检索，形成更适合代码任务的 Repo RAG。

6. **新增 Checkpoint & Rollback System**  
   借鉴 Cline 的 diff review、checkpoint、revert 思想，保证所有 Agent 修改都可审查、可回滚、可追踪。

7. **新增 AGENTS.md Compatibility Layer**  
   初始化项目时生成 AGENTS.md，同时维护 `.ai-company/` 私有配置，使项目可兼容 Codex、Claude Code、Gemini CLI、Cursor、Copilot、Cline、Roo Code 等外部 Agent 工具。

8. **新增 Eval Harness 与自迭代闭环**  
   用评测指标驱动 Self-Iteration Agent，而不是凭感觉优化 Prompt 或模型路由。

9. **自迭代分层明确化**  
   MVP 只做任务级自修复和经验沉淀；V1 做 Prompt / Workflow / Model Router 优化建议；V2 再做安全沙箱内的系统自修改实验。

---

## 1. 项目一句话定位

**AI Software Company Runtime 是一个“虚拟软件公司”式多 Agent 协作开发工具。用户提出一个项目目标或开发任务后，系统先自动判断该项目需要什么团队，再动态组建 PM、架构师、后端、前端、Agent 工程师、RAG 工程师、QA、代码审查、DevOps、OCR/视觉、记忆管理、自迭代等 Agent，并通过 Loop Engineering 执行闭环持续规划、检索、执行、测试、审查、修正和沉淀经验，最终输出可审查的代码 Diff、测试报告、文档、提交信息或 PR。**

产品 slogan：

> **不是让一个 AI 写代码，而是组建一支 AI 软件公司协作交付。**

更完整的技术定位：

> **Go-native AI Software Company Runtime，融合 DeepAgents Harness 思想、MetaGPT 软件公司 SOP、AutoGen 多 Agent 协作、Hermes 记忆与技能沉淀、Loop Engineering 自迭代闭环。**

---

## 2. 项目背景与痛点

当前主流 Coding Agent / 多 Agent 工具存在几个明显问题。

### 2.1 单 Agent 模式过于粗糙

一个模型同时负责需求、架构、编码、测试、审查和部署，会导致：

- 角色混乱
- 上下文污染
- 质量不可控
- 高风险操作缺少审批
- 失败后不会沉淀经验

### 2.2 多 Agent 常常只是“多个模型聊天”

很多多 Agent 框架看似角色很多，但缺少真正的软件工程约束：

- 没有明确 SOP
- 没有任务状态机
- 没有 Diff 审批
- 没有 Git/Checkpoint/rollback
- 没有测试闭环
- 没有成本统计
- 没有模型表现评估

### 2.3 不同项目需要不同团队

普通后端 API、前端页面、RAG 平台、AI Agent 项目、爬虫自动化、桌面 App、DevOps 项目需要的 Agent 队伍完全不同。

所以本项目的第一原则是：

> **不要固定启用所有 Agent，而是先由 Team Architect Agent 判断项目需要什么团队。**

### 2.4 用户需要可控，而不是全自动黑盒

工程任务必须允许用户：

- 查看计划
- 查看相关文件
- 查看 Agent 执行动作
- 查看代码 Diff
- 确认写入文件
- 确认执行 shell
- 确认 commit / PR
- 一键回滚

### 2.5 真正有价值的是持续循环，而不是一次性提示词

软件开发不是“一问一答”，而是：

```text
计划 → 检索 → 修改 → 测试 → 报错 → 反思 → 修复 → 审查 → 再测试 → 交付 → 沉淀经验
```

因此本项目必须内置 Loop Engineering 和自迭代闭环。

---

## 3. 核心设计原则

### 3.1 Go-native Core Runtime

底层 Runtime 使用 Go 构建，原因：

- 单二进制分发，适合 CLI 工具
- 跨平台能力强，Windows / macOS / Linux 都友好
- 文件系统、进程管理、Git、Shell、网络请求能力稳定
- 并发和长任务调度能力强
- 适合做权限控制和沙箱边界
- 适合后续作为 Windows App 的本地 sidecar

Go 负责稳定底座，Python/TypeScript 作为插件层接 AI 生态。

### 3.2 Agent 职责边界清晰

每个 Agent 都必须定义：

- 职责
- 输入
- 输出
- 默认模型
- 备选模型
- 工具权限
- 是否允许写文件
- 是否允许执行命令
- 是否需要人工确认

### 3.3 用户可自定义每个 Agent 的模型

系统提供默认模型配置，但用户必须可以自定义：

- 每个 Agent 用什么模型
- fallback 模型
- temperature
- max rounds
- context budget
- cost tier
- 工具权限
- 执行模式

核心理念：

> **每个 Agent 都是一名虚拟员工，每名员工都可以配置自己的模型、权限和工作方式。**

### 3.4 Team Architect Agent 优先级最高

系统启动任务时，不直接进入 Orchestrator，而是：

```text
用户需求 / 项目仓库
  ↓
Project Scanner
  ↓
Team Architect Agent
  ↓
生成 Team Config
  ↓
Orchestrator Agent
  ↓
专业 Agent 执行
```

Team Architect 决定“需要哪些员工”，Orchestrator 决定“这些员工怎么协作”。

### 3.5 所有危险动作都必须由 Go Runtime 强制控制

不能依赖模型自觉。

高风险动作必须由 Runtime 拦截：

- 写文件
- 删除文件
- 执行 shell
- 安装依赖
- 修改配置
- commit
- push
- 部署
- 操作数据库
- 读取密钥

### 3.6 RAG + 多层级记忆是核心竞争力

没有 RAG 和多层级记忆，它只是一个多 Agent 编码工具；加入 RAG 和记忆后，它才像一个会成长的软件公司。

### 3.7 自迭代必须有评测，不做玄学优化

自迭代必须建立在 Eval Harness 上：

- 成功率
- 测试通过率
- Review 通过率
- 平均轮数
- token 成本
- shell 调用次数
- 回滚率
- 人工介入次数

---

## 4. 融合思想来源与落地方式

本项目不是简单套壳任何框架，而是吸收优秀开源项目的工程思想。

### 4.1 AutoGen：多 Agent 对话与协作编排

吸收点：

- Agent 之间可以互相发送结构化消息
- Orchestrator 负责多 Agent 协作
- 人类用户可以插入确认、纠偏和审批

本项目改造方式：

```text
AutoGen-style conversation
  → Agent Registry + Team Config + Workflow State Machine
```

即：Agent 可以对话，但不能自由乱聊，必须受任务状态机约束。

### 4.2 MetaGPT：软件公司 SOP

吸收点：

- 软件开发可以拆成 PM、Architect、Engineer、QA 等角色
- Code = SOP(Team)
- 多 Agent 协作应当围绕明确流程展开

本项目改造方式：

```text
MetaGPT-style software company SOP
  → Dynamic Team Builder + Configurable Agent Workforce
```

即：不是固定软件公司，而是按项目动态组队。

### 4.3 Hermes Agent：记忆、技能和自学习

吸收点：

- Agent 应该沉淀经验
- 成功任务可以变成 skill
- 记忆不是聊天记录，而是可复用上下文和经验

本项目改造方式：

```text
Hermes-style memory & skills
  → Memory Agent + Skill Memory + Self-Iteration Agent
```

### 4.4 DeepAgents：Agent Harness

吸收点：

- Sub-agents
- Filesystem abstraction
- Shell access
- Context management
- Persistent memory
- Human-in-the-loop
- Skills
- MCP tools

本项目改造方式：

```text
DeepAgents-inspired harness
  → Go-native Agent Harness + Plugin Layer
```

本项目不绑定 LangChain / LangGraph，但借鉴其 harness 思想。

### 4.5 OpenHands：Agent Control Center / Always-on Engineering Team

吸收点：

- Agent 工作台
- 长期在线的工程团队视角
- 任务、日志、文件、命令、diff 可视化

本项目改造方式：

```text
OpenHands-style control center
  → CLI/TUI first, Windows Agent Workspace later
```

### 4.6 SWE-agent：Agent-Computer Interface

吸收点：

- 关键不是直接暴露 shell，而是设计适合 Agent 使用的计算机接口
- 让模型通过结构化动作浏览仓库、编辑文件、运行测试

本项目改造方式：

```text
SWE-agent ACI
  → Agent-Computer Interface Layer
```

### 4.7 mini-SWE-agent：极简工具层和 Bash Expert Mode

吸收点：

- 工具层可以很简洁
- bash 作为强大统一接口
- 关键在沙箱和反馈格式

本项目改造方式：

```text
Safe Tools Mode + Bash Expert Mode
```

默认 Safe Tools Mode，高级用户可开启 Bash Expert Mode，但所有 shell 必须经过 sandbox 和 approval policy。

### 4.8 Aider：Repo Map + Git-native 工作流

吸收点：

- Repo Map 比普通 RAG 更适合代码仓库
- Git-native 修改、diff、commit、rollback 很重要

本项目改造方式：

```text
Repo Map Engine + Repo RAG + Git Checkpoint
```

### 4.9 Cline：Plan/Act、Diff Review、Checkpoint

吸收点：

- Plan/Act 模式
- 修改前后差异可视化
- Checkpoint / revert
- IDE 内透明交互

本项目改造方式：

```text
Checkpoint & Rollback System
```

### 4.10 Roo Code：多模式 Agent

吸收点：

- Code / Architect / Ask / Debug / Custom Modes
- MCP 兼容
- 同一 Agent 可根据任务进入不同模式

本项目改造方式：

```text
Agent Role + Agent Mode
```

### 4.11 Gemini CLI：ReAct Loop + MCP + 本地终端

吸收点：

- Reason-and-Act loop
- 本地终端工具
- MCP server
- 修 bug、加功能、补测试的循环过程

本项目改造方式：

```text
Loop Engineering Runtime
```

### 4.12 AGENTS.md：项目级 Agent 指令标准

吸收点：

- AGENTS.md 作为给 coding agent 看的 README
- 可描述安装、测试、编码规范、提交规范
- 逐级目录优先生效

本项目改造方式：

```text
AGENTS.md Compatibility Layer
```

初始化项目时生成：

```text
/AGENTS.md
/.ai-company/agents.yaml
/.ai-company/models.yaml
/.ai-company/workflows.yaml
/.ai-company/memory/
/.ai-company/skills/
/.ai-company/evals/
```

### 4.13 Mem0：Universal Memory Layer

吸收点：

- 用户偏好记忆
- Agent 长期记忆
- 可插拔记忆服务

本项目改造方式：

```text
Memory Provider Layer
```

### 4.14 Letta / MemGPT：自管理记忆

吸收点：

- Agent 自己判断什么应该写入长期记忆
- 压缩上下文
- 管理过期信息

本项目改造方式：

```text
Memory Agent
```

### 4.15 Zep / Graphiti：Temporal Knowledge Graph

吸收点：

- 时间感知知识图谱
- 上下文关系检索
- 生产级 Agent 记忆治理

本项目改造方式：

```text
V2: Graph Memory / Decision Graph / Bug-Fix Graph
```

### 4.16 agentmemory：共享记忆服务

吸收点：

- 多个外部 coding agent 共享记忆
- 支持 MCP / REST / hooks

本项目改造方式：

```text
Local Memory Daemon
ai-company memory serve
```

### 4.17 Eino：Go-native LLM 应用框架

吸收点：

- Go 语言习惯的 LLM 应用组件
- ChatModel / Tool / Retriever / Chain / Graph / Callback / Workflow

本项目改造方式：

```text
Go Runtime 可参考 Eino 组件化思想，但核心 Runtime 自研。
```

### 4.18 MCP Go SDK / mcp-go：工具协议层

吸收点：

- MCP 是连接 LLM 应用和外部工具/数据源的开放协议
- Go 可同时做 MCP Client 和 MCP Server

本项目改造方式：

```text
MCP Client / Server 原生支持
```

### 4.19 SWE-bench / OpenHands Benchmarks / HAL Harness：评测和自迭代

吸收点：

- 用真实软件工程任务评估 Agent
- 记录成功率、成本、日志、复现过程
- 让模型和团队配置优化有数据依据

本项目改造方式：

```text
Eval Harness + Self-Iteration Agent
```

---

## 5. 总体架构

```text
AI Software Company Runtime
│
├── Go Core Runtime
│   ├── CLI / TUI
│   ├── Agent Orchestrator
│   ├── Team Architect Runtime
│   ├── Agent Registry
│   ├── Model Registry
│   ├── Agent-Computer Interface
│   ├── Tool Permission Policy
│   ├── Sandbox Shell
│   ├── Git Diff / Patch / Checkpoint
│   ├── MCP Client / Server
│   ├── Eval Harness
│   └── Telemetry / Cost Tracker
│
├── Context Engineering Layer
│   ├── Project Scanner
│   ├── Repo Map Engine
│   ├── Repo RAG
│   ├── Document RAG
│   ├── Decision RAG
│   ├── Task RAG
│   ├── Skill RAG
│   ├── Memory Retrieval
│   └── Context Compression
│
├── Memory Layer
│   ├── Session Memory
│   ├── Task Memory
│   ├── Project Memory
│   ├── Agent Memory
│   ├── Team Memory
│   ├── User Preference Memory
│   ├── Skill Memory
│   └── Graph Memory
│
├── Agent Layer
│   ├── Team Architect Agent
│   ├── Orchestrator Agent
│   ├── PM Agent
│   ├── Architect Agent
│   ├── Planner Agent
│   ├── Repo Analyst Agent
│   ├── Backend Engineer Agent
│   ├── Frontend Engineer Agent
│   ├── Agent Engineer Agent
│   ├── RAG Engineer Agent
│   ├── QA Engineer Agent
│   ├── Code Reviewer Agent
│   ├── DevOps Agent
│   ├── Document/OCR Agent
│   ├── Memory Agent
│   └── Self-Iteration Agent
│
└── Plugin Layer
    ├── Python Plugin Host
    ├── Node/TypeScript Plugin Host
    ├── OCR Tools
    ├── Browser Tools
    ├── External Agent Bridges
    ├── MCP Tools
    └── Model Provider Adapters
```

---

## 6. 技术栈建议

### 6.1 Go Core Runtime

建议使用：

```text
语言：Go
CLI：Cobra
TUI：Bubble Tea / Lipgloss
配置：YAML + Viper
本地数据库：SQLite
并发任务：goroutine + context + worker pool
Git：go-git 或调用 git CLI
Patch：diff/patch library + git apply
进程执行：os/exec + sandbox policy
日志：zerolog / slog
API：net/http / chi / fiber 可选
```

### 6.2 Python Plugin Layer

用于 AI 生态能力：

```text
OCR：PaddleOCR / RapidOCR
文档解析：MarkItDown / unstructured
Embedding：bge-m3 / qwen embedding / jina embedding / OpenAI embedding
Rerank：bge-reranker / qwen reranker
向量库：Chroma / LanceDB / Qdrant
代码解析：tree-sitter bindings
实验性框架桥接：deepagents / AutoGen / MetaGPT
```

### 6.3 TypeScript Plugin Layer

用于：

```text
MCP server/client 扩展
前端生态工具
Node 项目分析
Playwright browser automation
VS Code / Cursor 插件桥接
```

### 6.4 Windows App 第二阶段

推荐：

```text
Tauri + React + Go Sidecar
```

Windows App 页面：

```text
Dashboard：任务状态、成本、Agent 活动
Agents：角色、模型、权限、Prompt
Models：API Key、Base URL、连通性测试
Workspace：文件树、diff、终端输出
Task Board：任务拆解、状态、负责人 Agent
Memory：项目记忆、架构决策、历史 Bug、技能库
Eval：模型表现、团队表现、任务成功率
```

---

## 7. 核心模块设计

### 7.1 Project Scanner

职责：

- 扫描目录树
- 识别技术栈
- 读取 README / AGENTS.md
- 读取 package.json / pyproject.toml / go.mod / pom.xml
- 检查 Dockerfile / compose / CI 配置
- 识别前端、后端、测试、文档目录
- 生成 project_profile

输出示例：

```yaml
project_profile:
  language:
    - go
    - typescript
  framework:
    - cobra
    - sqlite
  has_frontend: false
  has_backend: true
  has_tests: true
  has_docker: false
  has_agents_md: true
  project_type_guess: "cli_tool"
```

### 7.2 Agent Registry

保存系统内置 Agent 定义：

- agent_id
- name
- description
- default_model
- fallback_models
- tools
- permissions
- system_prompt
- modes
- output_schema

### 7.3 Model Registry

保存模型配置：

- provider
- model id
- base url
- api key env
- supports_vision
- supports_tools
- supports_json
- max context
- cost tier
- default role

### 7.4 Team Architect Runtime

Team Architect Agent 是 P0 优先级。

职责：

- 判断项目类型
- 判断复杂度
- 判断风险等级
- 判断需要哪些能力
- 从 Agent Registry 中选择 Agent
- 为每个 Agent 推荐模型
- 设置权限等级
- 生成 Team Config

执行链路：

```text
Project Scanner → Team Architect Agent → Team Config → Orchestrator
```

### 7.5 Orchestrator Runtime

职责：

- 读取 Team Config
- 读取 Workflow
- 调度 Agent
- 管理任务状态
- 控制循环次数
- 触发用户确认
- 聚合最终结果

### 7.6 Agent-Computer Interface Layer

借鉴 SWE-agent。

不要让模型直接操作底层系统，而是通过 Action API：

```text
inspect_project
search_code
open_file
read_file
edit_file
generate_patch
apply_patch
run_command
run_test
read_log
create_checkpoint
rollback
read_git_diff
create_commit_message
```

每个 action 都有：

- 输入 schema
- 输出 schema
- 权限要求
- 风险等级
- 是否需要审批
- 日志记录

### 7.7 Tool Permission Policy

权限等级：

```text
Level 0: read_only
Level 1: patch_only
Level 2: write_with_approval
Level 3: auto_write_in_sandbox
Level 4: commit_with_approval
Level 5: deploy_with_manual_approval，MVP 禁用
```

默认策略：

- 不允许删除文件
- 不允许读取密钥文件
- 不允许直接 push main
- 不允许直接部署生产
- shell 命令默认需要风险分类
- npm install / pip install 等需要用户确认

### 7.8 Repo Map Engine

借鉴 Aider。

目标：让 Agent 理解代码仓库结构，而不是盲目向量检索。

组成：

```text
- 文件树
- 符号表
- 类 / 函数 / 方法索引
- import graph
- dependency graph
- call graph，后期
- git history relevance，后期
- test mapping
- token budget renderer
```

Repo RAG 最佳形态：

```text
Repo RAG = ripgrep + BM25 + vector search + tree-sitter symbol graph + repo map ranking
```

### 7.9 Checkpoint & Rollback System

借鉴 Cline。

每次 Agent 修改前创建 checkpoint：

```yaml
checkpoint:
  id: "ckpt_001"
  task_id: "task_001"
  agent_id: "backend_engineer"
  model_id: "glm_5_2"
  timestamp: "2026-06-24T00:00:00Z"
  before_git_hash: "..."
  changed_files:
    - "internal/runtime/orchestrator.go"
  command_logs:
    - "go test ./..."
  test_result: "failed / passed / skipped"
  review_result: "pending / approved / rejected"
```

用户必须能：

- 查看 diff
- 部分接受
- 全部接受
- 一键回滚
- 查看每个 Agent 做了什么

### 7.10 MCP Client / Server

Go Runtime 应原生支持：

```text
MCP Client：调用外部工具
MCP Server：把 ai-company 的能力暴露给其他 Agent
```

ai-company 作为 MCP Server 可暴露：

```text
project_scan
repo_map
memory_search
task_create
run_test
read_diff
checkpoint_create
rollback
```

### 7.11 AGENTS.md Compatibility Layer

初始化时生成：

```text
AGENTS.md
.ai-company/agents.yaml
.ai-company/models.yaml
.ai-company/team.yaml
.ai-company/workflows.yaml
.ai-company/memory/
.ai-company/skills/
.ai-company/evals/
```

AGENTS.md 内容包含：

- 项目简介
- 安装命令
- 测试命令
- Lint 命令
- 代码规范
- 提交规范
- Agent 注意事项
- 禁止操作

### 7.12 Eval Harness

用于评估模型、Agent、团队配置和工作流。

指标：

```text
success_rate
test_pass_rate
review_pass_rate
cost_per_task
avg_rounds
avg_tokens
shell_call_count
rollback_rate
human_intervention_rate
regression_rate
```

Eval 结果进入：

```text
Model Performance Memory
Team Performance Memory
Self-Iteration Agent
```

---

## 8. Agent 定义

### 8.1 Team Architect Agent，P0

定位：项目团队架构师。

职责：

- 判断项目类型
- 判断复杂度
- 判断需要哪些 Agent
- 推荐模型映射
- 设置权限等级
- 生成 Team Config

默认模型：

```text
GPT-5.5
```

备选：

```text
Claude Opus 4.8 / Gemini 3.1 Pro / GLM 5.2
```

权限：只读，不写代码。

### 8.2 Orchestrator Agent，P0

定位：总控调度。

职责：

- 读取 Team Config
- 调度 Agent
- 管理状态机
- 控制循环次数
- 汇总最终结果

默认模型：GPT-5.5  
备选：Claude Opus 4.8 / GLM 5.2

### 8.3 PM Agent

职责：

- 需求澄清
- PRD
- 用户故事
- 验收标准
- MVP 范围

默认模型：Claude Opus 4.8  
备选：GPT-5.5 / Gemini 3.1 Pro

### 8.4 Architect Agent

职责：

- 技术方案
- 模块拆分
- 接口设计
- 数据流
- 风险评估

默认模型：GPT-5.5  
备选：Claude Opus 4.8 / GLM 5.2

### 8.5 Planner Agent

职责：

- 拆任务
- 排依赖
- 分配 Agent
- 生成执行顺序

默认模型：Qwen3.7 / Qwen Max  
备选：GPT-5.5 / GLM 5.2 / MiniMax

### 8.6 Repo Analyst Agent

职责：

- 分析仓库
- 找相关文件
- 生成代码地图
- 判断影响范围

默认模型：Devstral 2 / GLM 5.2  
备选：Qwen3-Coder / DeepSeek V4 Pro / Kimi K2.7 Code

### 8.7 Backend Engineer Agent

职责：

- 后端业务逻辑
- API
- 数据库访问
- Bug 修复

默认模型：GLM 5.2  
备选：Kimi K2.7 Code / Qwen3-Coder / DeepSeek V4 Pro

### 8.8 Frontend Engineer Agent

职责：

- 页面
- 组件
- API 接入
- 样式
- 根据截图还原 UI

默认模型：Kimi K2.7 Code  
视觉辅助：Gemini 3.1 Pro  
备选：Qwen3-Coder / GLM 5.2

### 8.9 Agent Engineer Agent

职责：

- Agent 工作流
- Prompt
- MCP
- 工具调用
- 模型路由
- AI 功能实现

默认模型：GPT-5.5  
备选：GLM 5.2 / Qwen3-Coder / Kimi K2.7 Code

### 8.10 RAG Engineer Agent

职责：

- 代码切片
- 文档切片
- 向量检索
- BM25
- Rerank
- 上下文压缩
- 检索质量评估

默认模型：GLM 5.2 / Qwen3-Coder / DeepSeek V4 Pro  
方案设计可用：GPT-5.5

### 8.11 QA Engineer Agent

职责：

- 测试用例
- 单元测试
- 集成测试
- 失败日志分析
- 测试报告

默认模型：DeepSeek V4 Pro  
备选：Qwen3-Coder / GLM 5.2 / Kimi K2.7 Code

### 8.12 Code Reviewer Agent

职责：

- 代码质量审查
- 安全风险
- 性能问题
- 架构一致性
- 是否建议合并

默认模型：Claude Opus 4.8  
备选：GPT-5.5 / DeepSeek V4 Pro / GLM 5.2

### 8.13 DevOps Agent

职责：

- 构建失败
- Dockerfile
- CI/CD
- 部署脚本
- 日志分析

默认模型：Qwen3.7  
备选：DeepSeek V4 Pro / Qwen3-Coder / GLM 5.2

### 8.14 Document/OCR Agent

职责：

- PDF / Word / 图片 / 截图解析
- OCR
- 表格提取
- 设计稿理解
- 报错截图理解

默认视觉模型：Gemini 3.1 Pro  
OCR 工具：PaddleOCR / RapidOCR / Mistral OCR

### 8.15 Memory Agent

职责：

- 总结任务
- 清洗记忆
- 判断什么值得保存
- 更新 Project Memory / Task Memory / Agent Memory / Skill Memory
- 防止垃圾记忆污染

默认模型：GPT-5.5 / Claude Opus 4.8  
备选：DeepSeek V4 Pro / GLM 5.2

### 8.16 Self-Iteration Agent

职责：

- 分析任务失败原因
- 总结成功经验
- 提取可复用 skill
- 优化 Agent Prompt 建议
- 优化模型路由建议
- 优化团队配置建议
- 生成 eval 改进建议

权限分级：

```text
MVP：只能提出建议，不能自动改系统配置
V1：可以生成配置 patch，用户确认后应用
V2：可以在安全沙箱中自动实验和评分
```

默认模型：GPT-5.5 / Claude Opus 4.8  
备选：DeepSeek V4 Pro / GLM 5.2 / Qwen3-Coder

---

## 9. Agent Mode System

借鉴 Roo Code。

每个 Agent 支持 mode，而不是只有固定角色。

示例：

```yaml
agent_modes:
  pm:
    - ask
    - prd
    - acceptance
  architect:
    - architecture
    - review
    - refactor
  backend_engineer:
    - code
    - debug
    - patch
    - test_fix
  code_reviewer:
    - strict_review
    - security_review
    - performance_review
  self_iteration:
    - reflect
    - skill_extract
    - router_optimize
```

---

## 10. 团队模板

### 10.1 backend_api

```yaml
required_agents:
  - pm
  - architect
  - repo_analyst
  - backend_engineer
  - qa_engineer
  - code_reviewer
optional_agents:
  - database
  - devops
  - security
```

### 10.2 frontend_app

```yaml
required_agents:
  - pm
  - uiux
  - repo_analyst
  - frontend_engineer
  - qa_engineer
  - code_reviewer
optional_agents:
  - document_ocr
  - backend_engineer
```

### 10.3 fullstack_saas

```yaml
required_agents:
  - pm
  - architect
  - repo_analyst
  - backend_engineer
  - frontend_engineer
  - database
  - qa_engineer
  - code_reviewer
optional_agents:
  - devops
  - security
  - uiux
```

### 10.4 agent_app

```yaml
required_agents:
  - pm
  - architect
  - agent_engineer
  - backend_engineer
  - repo_analyst
  - qa_engineer
  - code_reviewer
optional_agents:
  - memory
  - cost_controller
  - document_ocr
  - devops
```

### 10.5 rag_platform

```yaml
required_agents:
  - pm
  - architect
  - rag_engineer
  - document_ocr
  - data_engineer
  - backend_engineer
  - database
  - qa_engineer
  - code_reviewer
optional_agents:
  - frontend_engineer
  - evaluation
  - security
  - devops
```

### 10.6 crawler_automation

```yaml
required_agents:
  - architect
  - backend_engineer
  - data_engineer
  - qa_engineer
  - code_reviewer
optional_agents:
  - vision_debug
  - document_ocr
  - devops
```

### 10.7 desktop_app

```yaml
required_agents:
  - pm
  - architect
  - desktop_app_engineer
  - frontend_engineer
  - qa_engineer
  - code_reviewer
optional_agents:
  - devops
  - uiux
  - document_ocr
```

---

## 11. RAG Context Layer

RAG 在本项目中不是普通知识库问答，而是 Agent 执行前的上下文检索层。

### 11.1 Repo RAG，P0

用途：

- 根据需求找相关文件
- 根据报错找相关代码
- 找调用链
- 找模块依赖
- 缩小修改范围

组成：

```text
ripgrep + SQLite FTS5 + vector search + tree-sitter + repo map
```

### 11.2 Document RAG，P1

检索：

- README
- PRD
- API 文档
- 数据库文档
- 部署文档
- 用户上传 PDF / Word / 图片 OCR

### 11.3 Decision RAG，P1

保存架构决策：

- 为什么选某个框架
- 为什么用某种目录结构
- 为什么某个 Agent 没有写权限
- 为什么某个模型负责某类任务

### 11.4 Task RAG，P1

检索历史任务：

- 类似 Bug
- 类似需求
- 类似测试失败
- 类似 Review 结论

### 11.5 Skill RAG，P2

检索技能库：

- 常见修复套路
- 项目初始化流程
- 测试修复流程
- 部署排障流程
- 前端截图还原流程

---

## 12. Multi-level Memory Layer

### 12.1 Session Memory

生命周期：一次对话 / 一次任务。

保存：当前目标、临时假设、Agent 讨论内容。

### 12.2 Task Memory

生命周期：一个任务从开始到结束。

保存：任务目标、拆解、参与 Agent、修改记录、测试结果、Review 结论、失败原因。

### 12.3 Project Memory，P0

生命周期：整个项目。

保存：技术栈、目录结构、代码规范、架构决策、API 约定、数据库约定、部署方式、历史 Bug、团队配置。

### 12.4 Agent Memory

保存某个 Agent 在当前项目中的经验。

例如 Backend Engineer 记住：

- 后端使用 FastAPI
- 返回格式统一为 ResponseModel
- 数据库操作必须走 repository 层

### 12.5 Team Memory

保存当前项目团队配置：

- 启用了哪些 Agent
- 每个 Agent 负责什么
- 谁最终审查
- 哪些模型在本项目表现好

### 12.6 User Preference Memory

保存用户偏好：

- 中文输出
- CLI 优先
- 代码修改先给 diff
- 偏好 Go / Python / FastAPI
- 成本优先还是质量优先

### 12.7 Skill Memory

保存可复用技能：

- 修复 Go test 失败
- 生成 CLI 命令
- 初始化 RAG 项目
- 根据截图生成前端
- 分析 Docker 构建失败

### 12.8 Graph Memory，V2

保存关系型经验：

- Bug 与修复方案
- 模块依赖
- 架构决策关系
- Agent 与任务成功率
- 模型与任务类型表现

---

## 13. Loop Engineering Runtime

最终循环：

```text
Goal
  ↓
Plan
  ↓
Retrieve
  ↓
Act
  ↓
Observe
  ↓
Reflect
  ↓
Improve
  ↓
Review
  ↓
Commit Memory
```

说明：

- Goal：明确目标
- Plan：生成任务计划
- Retrieve：从 Repo RAG / Memory 检索上下文
- Act：执行代码修改或工具调用
- Observe：观察测试、日志、diff
- Reflect：反思失败原因
- Improve：修正方案
- Review：代码审查、质量审查
- Commit Memory：沉淀经验

---

## 14. 自迭代 Self-Iteration 设计

### 14.1 L1：任务级自修复，MVP

```text
测试失败 → 读取日志 → 定位原因 → 修改 → 再测试
```

### 14.2 L2：项目级经验沉淀，MVP / V1

Memory Agent 总结：

- 失败原因
- 成功修复方式
- 有效测试命令
- 常见坑
- Reviewer 拒绝原因

### 14.3 L3：Prompt / Workflow / Model Router 自优化，V1

根据 Eval 数据建议：

- 某类任务优先给哪个模型
- 某个 Agent prompt 增加什么约束
- 某类任务是否自动启用 OCR / DevOps / QA
- 某类项目需要什么团队模板

### 14.4 L4：系统自修改，V2

限制：

- 只能新分支
- 只能生成 patch
- 必须测试
- 必须 Reviewer 审查
- 必须用户确认
- 不允许改权限系统、密钥系统、生产部署逻辑

---

## 15. 模型池与默认分工

### 15.1 核心决策层

```text
GPT-5.5：CEO / Team Architect / Architect / Agent Engineer
Claude Opus 4.8：PM / Code Reviewer / Refactor Lead
Gemini 3.1 Pro：多模态、UI、截图、PDF、视觉理解
```

### 15.2 主力工程层

```text
GLM 5.2：后端、长程编码、Agentic workflow
Kimi K2.7 Code：前端、多文件 patch、代码实现
DeepSeek V4 Pro：低成本长上下文、测试、日志分析
Qwen3-Coder：开源 coding agent、本地/私有化、中文工程
Devstral 2：仓库分析、真实软件工程 Agent
```

### 15.3 工具执行和成本优化层

```text
Qwen3.7 / Qwen Max：工具调用、DevOps、自动化
MiniMax：低成本 worker
Grok Build：快速代码草稿，可选
Llama 4：私有化通用 fallback，可选
```

### 15.4 OCR / 文档工具层

```text
PaddleOCR：本地中文 OCR
RapidOCR：轻量 OCR
Mistral OCR：云端复杂文档解析
Gemini 3.1 Pro：视觉语义理解
```

---

## 16. 用户配置设计

示例：`.ai-company/models.yaml`

```yaml
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true

  opus_4_8:
    provider: anthropic
    model: claude-opus-4.8
    api_key_env: ANTHROPIC_API_KEY
    role: review
    cost_tier: high

  glm_5_2:
    provider: zhipu
    model: glm-5.2
    api_key_env: ZHIPU_API_KEY
    role: coding
    cost_tier: medium

  qwen3_coder:
    provider: dashscope
    model: qwen3-coder
    api_key_env: DASHSCOPE_API_KEY
    role: coding
    cost_tier: low
```

示例：`.ai-company/agents.yaml`

```yaml
agents:
  team_architect:
    default_model: gpt_5_5
    fallback_models:
      - opus_4_8
      - glm_5_2
    permission_level: read_only
    require_approval: false

  backend_engineer:
    default_model: glm_5_2
    fallback_models:
      - kimi_k2_7_code
      - qwen3_coder
      - deepseek_v4_pro
    permission_level: patch_only
    require_approval: true

  code_reviewer:
    default_model: opus_4_8
    fallback_models:
      - gpt_5_5
      - deepseek_v4_pro
    permission_level: read_only
    require_approval: false
```

---

## 17. CLI 命令设计

### 17.1 初始化

```bash
ai-company init
ai-company init --profile go-cli
ai-company init --generate-agents-md
```

### 17.2 模型管理

```bash
ai-company models list
ai-company models test gpt_5_5
ai-company models add
ai-company models set backend_engineer qwen3_coder
```

### 17.3 Agent 管理

```bash
ai-company agents list
ai-company agents show backend_engineer
ai-company agents edit backend_engineer
ai-company agents modes backend_engineer
```

### 17.4 项目扫描和团队组建

```bash
ai-company scan
ai-company team suggest
ai-company team apply
ai-company team show
```

### 17.5 任务执行

```bash
ai-company task "帮我给登录接口加验证码"
ai-company run
ai-company run --mode safe
ai-company run --mode semi-auto
```

### 17.6 Diff / 审查 / 回滚

```bash
ai-company diff
ai-company review
ai-company apply
ai-company checkpoint list
ai-company rollback <checkpoint_id>
```

### 17.7 RAG / Memory

```bash
ai-company index
ai-company memory search "认证逻辑"
ai-company memory add
ai-company memory summarize
ai-company memory serve
```

### 17.8 Eval / 自迭代

```bash
ai-company eval run
ai-company eval report
ai-company self-iterate suggest
ai-company self-iterate apply --with-approval
```

---

## 18. 目录结构建议

```text
ai-company/
  cmd/
    ai-company/
      main.go
  internal/
    runtime/
    orchestrator/
    teamarchitect/
    agents/
    models/
    aci/
    tools/
    sandbox/
    gitops/
    checkpoint/
    memory/
    rag/
    repomap/
    workflow/
    approval/
    eval/
    selfiterate/
    telemetry/
    config/
  plugins/
    python/
      ocr/
      embedding/
      rerank/
      docparse/
    node/
      mcp/
      browser/
  configs/
    agents.yaml
    models.yaml
    workflows.yaml
    teams.yaml
  templates/
    AGENTS.md.tmpl
    agents.yaml.tmpl
    models.yaml.tmpl
  docs/
  tests/
```

---

## 19. 数据库设计，MVP

使用 SQLite。

### 19.1 tasks

```sql
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  title TEXT,
  description TEXT,
  status TEXT,
  project_type TEXT,
  team_config_id TEXT,
  created_at TEXT,
  updated_at TEXT
);
```

### 19.2 agent_runs

```sql
CREATE TABLE agent_runs (
  id TEXT PRIMARY KEY,
  task_id TEXT,
  agent_id TEXT,
  model_id TEXT,
  mode TEXT,
  input TEXT,
  output TEXT,
  status TEXT,
  token_usage INTEGER,
  cost_estimate REAL,
  created_at TEXT
);
```

### 19.3 memories

```sql
CREATE TABLE memories (
  id TEXT PRIMARY KEY,
  scope TEXT,
  scope_id TEXT,
  memory_type TEXT,
  content TEXT,
  tags TEXT,
  importance INTEGER,
  created_at TEXT,
  updated_at TEXT
);
```

### 19.4 checkpoints

```sql
CREATE TABLE checkpoints (
  id TEXT PRIMARY KEY,
  task_id TEXT,
  agent_id TEXT,
  git_hash_before TEXT,
  diff TEXT,
  command_logs TEXT,
  test_result TEXT,
  review_result TEXT,
  created_at TEXT
);
```

### 19.5 eval_runs

```sql
CREATE TABLE eval_runs (
  id TEXT PRIMARY KEY,
  task_id TEXT,
  team_config_id TEXT,
  success INTEGER,
  test_passed INTEGER,
  review_passed INTEGER,
  cost REAL,
  rounds INTEGER,
  rollback_count INTEGER,
  human_intervention_count INTEGER,
  created_at TEXT
);
```

---

## 20. MVP 范围

### 20.1 MVP 必做

```text
1. Go CLI 基础框架
2. 配置系统：models.yaml / agents.yaml
3. Project Scanner
4. Team Architect Agent
5. Orchestrator Agent
6. Agent Registry / Model Registry
7. ACI 基础动作：read/search/edit/generate_patch/run_test
8. Git diff + checkpoint + rollback
9. Repo Map 基础版：目录树 + ripgrep + tree-sitter 符号
10. Repo RAG 基础版：SQLite FTS5 + 简单向量检索可选
11. Task Memory + Project Memory
12. Memory Agent 任务总结
13. QA 测试循环
14. Code Reviewer Agent
15. AGENTS.md 生成
16. Eval 基础指标记录
```

### 20.2 MVP 不做

```text
1. 自动生产部署
2. 自动 push main
3. 完整 Windows App
4. 复杂图谱记忆
5. 自修改核心系统代码
6. 企业级权限系统
7. 多人协作 SaaS
8. 深度浏览器自动化
```

---

## 21. 开发路线

### Phase 0：项目骨架

- Go CLI
- 配置系统
- 日志系统
- SQLite 初始化
- 基础目录结构

### Phase 1：Agent Runtime

- Agent 抽象
- Model Adapter
- Orchestrator
- Team Architect
- Agent Registry
- Model Registry

### Phase 2：ACI + Git 安全层

- read/search/open/edit/generate_patch
- run_test
- diff
- checkpoint
- rollback
- approval policy

### Phase 3：Repo Map + Repo RAG

- Project Scanner
- Repo Map
- SQLite FTS5
- tree-sitter
- Context builder

### Phase 4：Memory Layer

- Task Memory
- Project Memory
- Memory Agent
- Memory search

### Phase 5：Loop Engineering

- Plan/Retrieve/Act/Observe/Reflect/Improve/Review/Commit Memory
- 测试失败自动修复循环
- Review 驳回自动修正循环

### Phase 6：Eval + Self-Iteration

- Eval metrics
- model performance log
- team performance log
- self-iteration suggestions

### Phase 7：Plugin Layer

- Python OCR plugin
- embedding plugin
- rerank plugin
- MCP tools

### Phase 8：Windows App

- Tauri + React
- Agent Dashboard
- Diff viewer
- Memory UI
- Eval UI

---

## 22. Codex 开发提示词

后续可直接交给 Codex 的总提示词：

```text
你是一名资深 Go 架构师和 AI Agent 工程师。

请基于《AI Software Company Runtime 项目企划书 V1.0》开发 MVP。

核心要求：
1. 使用 Go 构建 CLI-first 的核心 Runtime。
2. 不要一开始做 Windows App。
3. 不要依赖单一 Agent 框架，核心 Runtime 自研。
4. 必须实现配置驱动的 Agent Registry 和 Model Registry。
5. 必须实现 Team Architect Agent，优先根据项目类型生成 Team Config。
6. 必须实现 Orchestrator Agent，按 Team Config 调度 Agent。
7. 必须实现 Agent-Computer Interface Layer，禁止模型直接无限制操作 shell 和文件系统。
8. 必须实现 Git diff、checkpoint、rollback。
9. 必须实现 Repo Map 基础能力和 Repo RAG 基础能力。
10. 必须实现 Task Memory 和 Project Memory。
11. 必须实现 Loop Engineering 基础循环：Plan → Retrieve → Act → Observe → Reflect → Improve → Review → Commit Memory。
12. 必须实现 AGENTS.md 生成。
13. 必须实现 Eval 基础指标记录，为后续自迭代服务。
14. 所有危险操作必须走 approval policy。
15. 第一版优先保证架构清晰、接口稳定、可测试、可扩展。

请先输出：
- 项目目录结构
- 核心 Go interface 设计
- 配置文件 schema
- MVP 任务拆分
- 第一批可运行代码
```

---

## 23. 风险与控制

### 23.1 技术复杂度过高

控制：MVP 只做 Go CLI + 核心 Agent Runtime，不做完整 UI 和企业版。

### 23.2 Agent 乱改代码

控制：ACI + approval policy + diff + checkpoint + rollback。

### 23.3 记忆污染

控制：Memory Agent 负责清洗和打分，不直接把所有聊天记录写入长期记忆。

### 23.4 成本失控

控制：Model Router + cost tracker + fallback models + batch worker。

### 23.5 自迭代失控

控制：MVP 只允许建议，不允许自动改系统配置；V2 才允许沙箱实验。

### 23.6 框架绑定风险

控制：核心 Runtime 自研，deepagents / AutoGen / MetaGPT 只作为思想或插件桥接。

---

## 24. 最终产品卖点

1. **动态组队**  
   每个项目自动生成专属 AI 软件团队。

2. **用户可自定义每个 Agent 的模型**  
   默认高配，但可以按预算和偏好调整。

3. **Go-native Runtime**  
   稳定、可分发、跨平台、适合 CLI 和桌面 App。

4. **Agent-Computer Interface**  
   工具调用受控，安全边界清晰。

5. **Repo Map + Repo RAG**  
   更懂代码仓库，而不是普通向量检索。

6. **Checkpoint & Rollback**  
   所有修改可审查、可回滚。

7. **多层级记忆**  
   项目越用越懂，团队越用越稳。

8. **Loop Engineering**  
   能测试、观察、反思、修正，不是一锤子买卖。

9. **Eval-driven Self-Iteration**  
   通过真实指标优化模型路由、Prompt、团队配置和技能库。

10. **AGENTS.md + MCP 兼容**  
   可与 Codex、Claude Code、Gemini CLI、Cursor、Cline、Roo Code 等生态协作。

---

## 25. 最终一句话总结

**AI Software Company Runtime 的目标不是成为又一个 Coding Agent，而是成为一个 Go-native 的 AI 软件公司操作系统：先判断项目需要什么团队，再让合适的 Agent 在受控工具层、Repo RAG、多层级记忆、Checkpoint、安全审批和 Eval 驱动的自迭代闭环中完成真实软件工程交付。**
