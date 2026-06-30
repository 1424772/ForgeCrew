# ForgeCrew

> **Not just one AI writing code — assemble an entire AI software company to collaborate and deliver.**

ForgeCrew is a Go-native, CLI-first **AI Software Company Runtime**.
It's not a single coding agent, but a dynamically team-assembling multi-agent engineering runtime.

## Positioning

- **Dynamic Teaming**: Automatically analyzes each project and recommends a tailored AI software team
- **User-Controlled**: Every agent's model, permissions, and working style are customizable
- **Go-native Runtime**: Stable, distributable, cross-platform
- **Agent-Computer Interface**: Controlled tool invocation with clear security boundaries
- **Checkpoint & Rollback**: All changes are reviewable and reversible
- **Loop Engineering**: Engineering feedback loop — test, observe, reflect, correct
- **Eval-driven**: Optimize model routing, prompts, and team composition with real metrics
- **AGENTS.md Compatible**: Works alongside Codex, Claude Code, Gemini CLI, Cursor, Cline, and other ecosystems

## Installation

### Script Install (Recommended)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.sh | sh
```

**Windows PowerShell:**
```powershell
iwr https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.ps1 -useb | iex
```

**Windows CMD:**
```cmd
powershell -ExecutionPolicy Bypass -Command "iwr https://raw.githubusercontent.com/1424772/ForgeCrew/main/scripts/install.ps1 -useb | iex"
```

Run directly after installation:
```cmd
forgecrew
```

Scripts install to the user directory (`~/.forgecrew/bin/` or `%USERPROFILE%\.forgecrew\bin\`), no admin privileges required.
If GitHub Releases are not yet published, the script will guide you through local builds via `go build`.

### Manual Install

```bash
# go install
go install github.com/1424772/ForgeCrew/cmd/forgecrew@latest

# or build from source
git clone https://github.com/1424772/ForgeCrew.git
cd ForgeCrew
go build -o forgecrew ./cmd/forgecrew
```

Requires Go 1.21+.

## Quick Start

### 1. Initialize a Project

```bash
forgecrew init
```

Generates in the current directory:

```
AGENTS.md                      # Project documentation for AI Agents
.forgecrew/
├── agents.yaml                # Agent definitions (role, model, permissions)
├── models.yaml                # Model definitions (provider, cost, capabilities)
├── workflows.yaml             # Workflow definitions
├── memory/                    # Task/project memory
├── evals/                     # Evaluation records
├── checkpoints/               # Change checkpoints
└── runs/                      # Run/task state
```

Running `init` again will not overwrite existing files unless `--force` is used.

### 2. Scan a Project

```bash
forgecrew scan
forgecrew scan -o json
```

Outputs a project profile (language, framework, tests, Docker, etc.).

### 3. Get Team Suggestions

```bash
forgecrew team suggest
```

Recommends a team composition based on project type (PM, Architect, Backend, Frontend, QA, Reviewer, etc.).

### 4. View Agents and Models

```bash
forgecrew agents list
forgecrew models list
```

### 5. View Changes

```bash
forgecrew diff
```

### 6. Manage Checkpoints

```bash
forgecrew checkpoint list
```

### 7. Execute Tasks

```bash
# Dry-run (default): prints state sequence without making real calls
forgecrew task "add login validation"

# Explicit dry-run
forgecrew task "add login validation" --dry-run

# Real LLM call: generates an execution plan (no file writes, no shell execution)
forgecrew task "add login validation" --execute
```

`--execute` requires models to be configured in `models.yaml` and API keys set via environment variables. Example:
```bash
export OPENAI_API_KEY="sk-..."
forgecrew task "add login validation" --execute
```

### 8. Agent Handoff & CTO Review

ForgeCrew supports recording agent work status and CTO review through structured workflows.

```bash
# Submit an agent handoff (auto-collects git diff, creates task state and handoff file)
forgecrew handoff submit backend-001 \
  --agent backend_engineer \
  --goal "Implement provider layer" \
  --summary "Added internal/llm, task --execute can call OpenAI-compatible APIs"

# List all runs
forgecrew runs list

# View run details and task status
forgecrew runs status <run_id>

# CTO review of a task (rule-based audit, no LLM)
forgecrew cto review backend-001
```

**CTO review is currently a local rule-based audit**, not connected to an LLM. Audit criteria:
- Task status must be `ready_for_review`
- Handoff file must exist
- `changed_files` must not be empty
- `go test ./...` and `go vet ./...` must pass

All state files are stored under `.forgecrew/runs/<run_id>/`.

### Current CTO Review Limitations
- No LLM integration — no semantic code quality analysis
- No automatic `git add`/`commit`/`push`
- No automatic code modification
- Factual audit only: state, handoff, git diff, test results

## CLI Commands

| Command | Description |
|---------|-------------|
| `forgecrew` | Enter interactive mode |
| `forgecrew version` | Print version number |
| `forgecrew init` | Initialize ForgeCrew configuration |
| `forgecrew scan` | Scan project and output profile |
| `forgecrew agents list` | List all agents |
| `forgecrew models list` | List all models |
| `forgecrew team suggest` | Suggest team configuration |
| `forgecrew diff` | Show git diff |
| `forgecrew checkpoint list` | List all checkpoints |
| `forgecrew task <goal>` | Execute task via Loop Engineering state machine |
| `forgecrew runs list` | List all runs |
| `forgecrew runs status <id>` | View run and task status |
| `forgecrew handoff submit <id>` | Submit agent handoff |
| `forgecrew cto review <id>` | CTO review of a task |
| `forgecrew lang` | Switch language (zh/en) |

## Internationalization

ForgeCrew supports both Chinese and English UI languages. By default it follows the system locale, but you can switch manually:

```bash
forgecrew lang zh   # Switch to Chinese
forgecrew lang en   # Switch to English
```

## Project Structure

```
cmd/forgecrew/         # CLI entry point
internal/
├── config/            # Configuration constants and templates
├── models/            # Model registry
├── agents/            # Agent registry
├── scanner/           # Project scanner
├── teamarchitect/     # Team architect rule engine
├── orchestrator/      # Loop Engineering state machine
├── provider/          # LLM Provider client (OpenAI-compatible API)
├── aci/               # Agent-Computer Interface
├── gitops/            # Controlled git operations
├── checkpoint/        # Checkpoint system
├── runs/              # Run/task state management
├── memory/            # Memory storage (JSONL)
├── eval/              # Evaluation records
├── i18n/              # Internationalization support
└── settings/          # Settings management
templates/             # Template files
configs/               # Example configurations
scripts/               # Installation scripts
docs/                  # Documentation
```

## Development

```bash
# Run tests
go test ./...

# Code vetting
go vet ./...
```

## Roadmap

- **MVP (Current)**: Go CLI skeleton, configuration system, Scanner, Team Architect, Orchestrator skeleton, ACI, Checkpoint, Memory/Eval infrastructure, Provider integration, Runs management, CTO Review
- **V1**: Enhanced real LLM calls, Repo Map, Repo RAG, MCP Client/Server, full Loop Engineering
- **V2**: Windows App (Tauri), Graph Memory, self-modifying sandbox experiments

## License

MIT
