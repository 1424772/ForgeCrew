package config

// DefaultAgentsMD is the default content for AGENTS.md.
const DefaultAgentsMD = `# AGENTS.md

This file provides guidance to AI coding agents working on this project.

## Project Overview

<!-- Describe your project here -->

## Setup

<!-- How to install dependencies and build -->

## Testing

<!-- How to run tests -->

## Code Style

<!-- Coding conventions -->

## Commit Guidelines

<!-- Commit message format -->

## Agent Restrictions

- Do NOT delete files without explicit approval.
- Do NOT read secret/credential files.
- Do NOT push to main directly.
- Do NOT deploy to production without manual approval.
- Always create a checkpoint before making changes.

## ForgeCrew Configuration

This project uses ForgeCrew for AI-assisted development.
Configuration is in .forgecrew/ — do not modify manually unless you know what you are doing.
`

// DefaultAgentsYAML is the default agents configuration.
const DefaultAgentsYAML = `# ForgeCrew Agent Definitions
# Each agent is a virtual team member with a role, model, and permissions.

agents:
  team_architect:
    agent_id: team_architect
    name: Team Architect
    description: Analyzes the project and recommends team composition.
    default_model: gpt_5_5
    fallback_models:
      - claude_opus_4_8
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - scan_project
      - list_agents
      - list_models
    modes:
      - analyze
      - recommend

  orchestrator:
    agent_id: orchestrator
    name: Orchestrator
    description: Coordinates multi-agent workflows and manages task state.
    default_model: gpt_5_5
    fallback_models:
      - claude_opus_4_8
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - read_team_config
      - dispatch_agent
      - check_state
    modes:
      - coordinate
      - monitor

  pm:
    agent_id: pm
    name: Product Manager
    description: Clarifies requirements, writes PRDs, defines acceptance criteria.
    default_model: claude_opus_4_8
    fallback_models:
      - gpt_5_5
      - gemini_3_1_pro
    permission_level: read_only
    require_approval: false
    tools:
      - read_file
      - search_code
    modes:
      - ask
      - prd
      - acceptance

  architect:
    agent_id: architect
    name: Software Architect
    description: Designs technical solutions, modules, interfaces, and data flows.
    default_model: gpt_5_5
    fallback_models:
      - claude_opus_4_8
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - read_file
      - search_code
      - read_project_profile
    modes:
      - architecture
      - review
      - refactor

  repo_analyst:
    agent_id: repo_analyst
    name: Repository Analyst
    description: Analyzes repository structure, finds relevant files, builds code maps.
    default_model: devstral_2
    fallback_models:
      - glm_5_2
      - qwen3_coder
      - deepseek_v4_pro
    permission_level: read_only
    require_approval: false
    tools:
      - scan_project
      - search_code
      - read_file
      - build_repo_map
    modes:
      - analyze
      - map
      - impact

  backend_engineer:
    agent_id: backend_engineer
    name: Backend Engineer
    description: Implements backend logic, APIs, database access, and bug fixes.
    default_model: glm_5_2
    fallback_models:
      - kimi_k2_7_code
      - qwen3_coder
      - deepseek_v4_pro
    permission_level: patch_only
    require_approval: true
    tools:
      - read_file
      - search_code
      - generate_patch
      - run_test
    modes:
      - code
      - debug
      - patch
      - test_fix

  frontend_engineer:
    agent_id: frontend_engineer
    name: Frontend Engineer
    description: Implements UI components, pages, API integration, and styles.
    default_model: kimi_k2_7_code
    fallback_models:
      - qwen3_coder
      - glm_5_2
    permission_level: patch_only
    require_approval: true
    tools:
      - read_file
      - search_code
      - generate_patch
      - run_test
    modes:
      - code
      - debug
      - patch
      - ui

  qa_engineer:
    agent_id: qa_engineer
    name: QA Engineer
    description: Writes tests, analyzes failures, and generates test reports.
    default_model: deepseek_v4_pro
    fallback_models:
      - qwen3_coder
      - glm_5_2
      - kimi_k2_7_code
    permission_level: patch_only
    require_approval: true
    tools:
      - read_file
      - search_code
      - generate_patch
      - run_test
    modes:
      - test
      - analyze
      - report

  code_reviewer:
    agent_id: code_reviewer
    name: Code Reviewer
    description: Reviews code quality, security, performance, and architecture consistency.
    default_model: claude_opus_4_8
    fallback_models:
      - gpt_5_5
      - deepseek_v4_pro
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - read_file
      - search_code
      - read_git_diff
    modes:
      - strict_review
      - security_review
      - performance_review

  devops:
    agent_id: devops
    name: DevOps Engineer
    description: Handles Docker, CI/CD, build failures, and deployment scripts.
    default_model: qwen3_7
    fallback_models:
      - deepseek_v4_pro
      - qwen3_coder
      - glm_5_2
    permission_level: write_with_approval
    require_approval: true
    tools:
      - read_file
      - search_code
      - run_test
      - generate_patch
    modes:
      - build
      - deploy
      - debug

  memory_agent:
    agent_id: memory_agent
    name: Memory Agent
    description: Summarizes tasks, cleans memory, decides what to persist.
    default_model: gpt_5_5
    fallback_models:
      - claude_opus_4_8
      - deepseek_v4_pro
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - read_memory
      - write_memory
    modes:
      - summarize
      - clean
      - index

  self_iteration:
    agent_id: self_iteration
    name: Self-Iteration Agent
    description: Analyzes failures, extracts skills, suggests optimizations.
    default_model: gpt_5_5
    fallback_models:
      - claude_opus_4_8
      - deepseek_v4_pro
      - glm_5_2
    permission_level: read_only
    require_approval: true
    tools:
      - read_eval
      - read_memory
      - suggest
    modes:
      - reflect
      - skill_extract
      - router_optimize
`

// DefaultModelsYAML is the default models configuration.
const DefaultModelsYAML = `# ForgeCrew Model Definitions
# Each model defines a provider, capabilities, and cost tier.

models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true

  claude_opus_4_8:
    provider: anthropic
    model: claude-opus-4-8
    api_key_env: ANTHROPIC_API_KEY
    role: review
    cost_tier: high
    supports_tools: true
    supports_vision: true

  gemini_3_1_pro:
    provider: google
    model: gemini-3.1-pro
    api_key_env: GEMINI_API_KEY
    role: multimodal
    cost_tier: high
    supports_tools: true
    supports_vision: true

  glm_5_2:
    provider: zhipu
    model: glm-5.2
    api_key_env: ZHIPU_API_KEY
    role: coding
    cost_tier: medium
    supports_tools: true
    supports_vision: false

  kimi_k2_7_code:
    provider: moonshot
    model: kimi-k2.7-code
    api_key_env: MOONSHOT_API_KEY
    role: coding
    cost_tier: medium
    supports_tools: true
    supports_vision: false

  deepseek_v4_pro:
    provider: deepseek
    model: deepseek-v4-pro
    api_key_env: DEEPSEEK_API_KEY
    role: coding
    cost_tier: low
    supports_tools: true
    supports_vision: false

  qwen3_coder:
    provider: dashscope
    model: qwen3-coder
    api_key_env: DASHSCOPE_API_KEY
    role: coding
    cost_tier: low
    supports_tools: true
    supports_vision: false

  qwen3_7:
    provider: dashscope
    model: qwen3.7
    api_key_env: DASHSCOPE_API_KEY
    role: tool_use
    cost_tier: low
    supports_tools: true
    supports_vision: false

  devstral_2:
    provider: devstral
    model: devstral-2
    api_key_env: DEVSTRAL_API_KEY
    role: coding
    cost_tier: medium
    supports_tools: true
    supports_vision: false
`

// DefaultWorkflowsYAML is the default workflows configuration.
const DefaultWorkflowsYAML = `# ForgeCrew Workflow Definitions
# Workflows define how agents collaborate on tasks.

workflows:
  loop_engineering:
    name: Loop Engineering
    description: The core iterative engineering loop (Goal → Plan → Retrieve → Act → Observe → Reflect → Improve → Review → CommitMemory).
    steps:
      - id: goal
        name: Goal
        description: Clarify the task goal and acceptance criteria.
        agent: pm
        mode: ask
      - id: plan
        name: Plan
        description: Generate a task plan with dependencies.
        agent: architect
        mode: architecture
      - id: retrieve
        name: Retrieve
        description: Retrieve relevant context from repo, memory, and docs.
        agent: repo_analyst
        mode: analyze
      - id: act
        name: Act
        description: Execute code changes or tool calls.
        agent: backend_engineer
        mode: code
      - id: observe
        name: Observe
        description: Observe test results, logs, and diffs.
        agent: qa_engineer
        mode: test
      - id: reflect
        name: Reflect
        description: Reflect on failures and identify root causes.
        agent: qa_engineer
        mode: analyze
      - id: improve
        name: Improve
        description: Apply fixes based on reflection.
        agent: backend_engineer
        mode: debug
      - id: review
        name: Review
        description: Review code quality and approve or reject changes.
        agent: code_reviewer
        mode: strict_review
      - id: commit_memory
        name: Commit Memory
        description: Persist lessons learned and project knowledge.
        agent: memory_agent
        mode: summarize
    max_iterations: 3
    require_approval_at:
      - act
      - improve

  quick_fix:
    name: Quick Fix
    description: A lightweight loop for simple bug fixes.
    steps:
      - id: retrieve
        name: Retrieve
        description: Find relevant code and context.
        agent: repo_analyst
        mode: analyze
      - id: act
        name: Act
        description: Apply the fix.
        agent: backend_engineer
        mode: patch
      - id: observe
        name: Observe
        description: Run tests and verify.
        agent: qa_engineer
        mode: test
      - id: review
        name: Review
        description: Quick review.
        agent: code_reviewer
        mode: strict_review
    max_iterations: 2
    require_approval_at:
      - act
`
