// Package i18n provides simple internationalization for ForgeCrew CLI.
// It supports zh (Chinese) and en (English) locales with a minimal
// key-based translation map.
package i18n

import "fmt"

// Locale represents a supported language.
type Locale string

const (
	ZH Locale = "zh"
	EN Locale = "en"
)

// ValidLocale checks whether s is a supported locale and returns it.
func ValidLocale(s string) (Locale, error) {
	switch s {
	case "zh":
		return ZH, nil
	case "en":
		return EN, nil
	default:
		return "", fmt.Errorf("unsupported language: %q (supported: zh, en)", s)
	}
}

// T returns the translation for key in the given locale.
// If the key or locale is missing, it returns the key itself as a fallback.
func T(key string, locale Locale) string {
	if m, ok := translations[key]; ok {
		if s, ok := m[locale]; ok {
			return s
		}
	}
	return key
}

// SupportedLocales returns the list of supported locale codes.
func SupportedLocales() []string {
	return []string{"zh", "en"}
}

// translations maps message keys to locale-specific strings.
var translations = map[string]map[Locale]string{
	// ── version ──
	"version.header": {
		ZH: "forgecrew 版本 ",
		EN: "forgecrew version ",
	},
	"version.commit": {
		ZH: "  提交: ",
		EN: "  commit: ",
	},
	"version.built": {
		ZH: "  构建: ",
		EN: "  built:  ",
	},

	// ── init ──
	"init.success": {
		ZH: "ForgeCrew 初始化成功。",
		EN: "ForgeCrew initialized successfully.",
	},
	"init.created": {
		ZH: " 已创建",
		EN: " created",
	},
	"init.skipping": {
		ZH: "  跳过（已存在，使用 --force 覆盖）: ",
		EN: "  skipping (already exists, use --force to overwrite): ",
	},

	// ── lang ──
	"lang.current": {
		ZH: "当前语言: ",
		EN: "Current language: ",
	},
	"lang.set_to": {
		ZH: "语言已设置为: ",
		EN: "Language set to: ",
	},
	"lang.available": {
		ZH: "可用语言:",
		EN: "Available languages:",
	},

	// ── task ──
	"task.header": {
		ZH: "任务: ",
		EN: "Task: ",
	},
	"task.dry_run_tag": {
		ZH: " [演习]",
		EN: " [dry-run]",
	},
	"task.iteration": {
		ZH: "第 ",
		EN: "Iteration ",
	},
	"task.iteration_suffix": {
		ZH: " 轮",
		EN: "",
	},
	"task.no_steps": {
		ZH: "(未执行任何步骤)\n",
		EN: "(no steps executed)\n",
	},
	"task.dry_run_note": {
		ZH: "[演习] ",
		EN: "[dry-run] ",
	},
	"task.dry_run_format": {
		ZH: "将为任务 %s 执行 %s",
		EN: "would execute %s for task %s",
	},

	// ── agents validate ──
	"validate.model_ok": {
		ZH: "Agent-Model 交叉校验通过。",
		EN: "Agent-Model cross-validation passed.",
	},
	"validate.tool_ok": {
		ZH: "Agent-Tool 校验通过。",
		EN: "Agent-Tool validation passed.",
	},
	"validate.all_ok": {
		ZH: "所有校验通过。",
		EN: "All validations passed.",
	},

	// ── agents list ──
	"agents.empty": {
		ZH: "没有注册的 Agent。请先运行 'forgecrew init'。",
		EN: "No agents registered. Run 'forgecrew init' first.",
	},

	// ── models list ──
	"models.empty": {
		ZH: "没有注册的模型。请先运行 'forgecrew init'。",
		EN: "No models registered. Run 'forgecrew init' first.",
	},

	// ── interactive mode ──
	"interactive.banner": {
		ZH: "ForgeCrew %s — AI Software Company Runtime",
		EN: "ForgeCrew %s — AI Software Company Runtime",
	},
	"interactive.project": {
		ZH: "项目: %s",
		EN: "Project: %s",
	},
	"interactive.language": {
		ZH: "语言: %s",
		EN: "Language: %s",
	},
	"interactive.mode": {
		ZH: "模式: ",
		EN: "Mode: ",
	},
	"interactive.prompt": {
		ZH: "> ",
		EN: "> ",
	},
	"interactive.help": {
		ZH: `可用命令:
  /help           显示帮助
  /exit           退出
  /lang show      显示当前语言
  /lang set <zh|en>  切换语言
  /mode show      显示当前模式
  /mode plan      切换到计划模式 (dry-run)
  /mode act       切换到执行模式 (dry-run, 写入未启用)
  /mode review    切换到审查模式 (未接入 reviewer)
  /scan           扫描当前项目
  /team           显示团队建议
  /validate       运行配置交叉校验
  直接输入任务目标即可通过 Loop Engineering 状态机执行。`,
		EN: `Available commands:
  /help           Show this help
  /exit           Exit
  /lang show      Show current language
  /lang set <zh|en>  Switch language
  /mode show      Show current mode
  /mode plan      Switch to plan mode (dry-run)
  /mode act       Switch to act mode (dry-run, writes disabled)
  /mode review    Switch to review mode (no reviewer yet)
  /scan           Scan current project
  /team           Show team suggestion
  /validate       Run config cross-validation
  Type a task goal to run through the Loop Engineering state machine.`,
	},
	"interactive.plan_mode_note": {
		ZH: "[计划模式] 分析任务目标，生成执行计划...",
		EN: "[plan mode] Analyzing task goal, generating execution plan...",
	},
	"interactive.dry_run_warning": {
		ZH: "[演习模式] 所有操作均为模拟，不会修改任何文件。",
		EN: "[dry-run] All operations are simulated. No files will be modified.",
	},
	"interactive.goodbye": {
		ZH: "再见！",
		EN: "Goodbye!",
	},
	"interactive.no_init": {
		ZH: "提示: 当前项目尚未初始化。运行 'forgecrew init' 以生成配置。",
		EN: "Hint: project not initialized. Run 'forgecrew init' to generate config.",
	},
	"interactive.start_hint": {
		ZH: "输入 /help 查看命令，输入 /exit 退出。",
		EN: "Type /help for commands, /exit to quit.",
	},
	"interactive.unknown_cmd": {
		ZH: "未知命令: %s。输入 /help 查看可用命令。",
		EN: "Unknown command: %s. Type /help for available commands.",
	},
	"interactive.review_disabled": {
		ZH: "[审查模式] 当前阶段尚未接入真实 reviewer。建议使用 agents validate 和 diff 命令检查代码。",
		EN: "[review mode] Real reviewer is not yet connected. Use 'agents validate' and 'diff' to inspect code.",
	},
	"interactive.act_disabled": {
		ZH: "[执行模式] 当前阶段仍然为演习模式，真实写入能力尚未启用。任务仅通过 dry-run 模拟。",
		EN: "[act mode] Real write capability is not yet enabled. Tasks run in dry-run simulation only.",
	},
	"interactive.scan_hint": {
		ZH: "扫描当前项目... 请使用 'forgecrew scan' 查看完整项目画像。",
		EN: "Scanning project... Use 'forgecrew scan' for full project profile.",
	},
	"interactive.team_hint": {
		ZH: "获取团队建议... 请使用 'forgecrew team suggest' 查看完整团队配置。",
		EN: "Getting team suggestion... Use 'forgecrew team suggest' for full team config.",
	},
	"interactive.validate_hint": {
		ZH: "运行交叉校验... 请使用 'forgecrew agents validate' 执行完整校验。",
		EN: "Running validation... Use 'forgecrew agents validate' for full cross-validation.",
	},

	// ── task execute ──
	"task.execute_mode": {
		ZH: " [执行]",
		EN: " [execute]",
	},
	"task.execute_missing_config": {
		ZH: "未找到模型配置文件。请先运行 'forgecrew init'。",
		EN: "Model configuration not found. Run 'forgecrew init' first.",
	},
	"task.execute_no_model": {
		ZH: "models.yaml 中没有配置任何模型。请在 models.yaml 中添加至少一个模型并设置对应的 API key 环境变量。",
		EN: "No models configured in models.yaml. Add at least one model and set its API key environment variable.",
	},
	"task.execute_provider_error": {
		ZH: "LLM 调用失败",
		EN: "LLM call failed",
	},
	"task.execute_plan_header": {
		ZH: "\n--- LLM 生成的计划 ---\n",
		EN: "\n--- LLM Generated Plan ---\n",
	},
}
