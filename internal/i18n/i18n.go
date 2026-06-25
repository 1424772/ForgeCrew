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
  /help          显示帮助
  /exit          退出
  /lang show     显示当前语言
  /lang set <zh|en>  切换语言
  /mode plan     切换到计划模式
  /mode act      切换到执行模式
  /mode review   切换到审查模式
  /scan          扫描当前项目
  /team          显示团队建议
  直接输入任务目标即可通过 Loop Engineering 状态机执行 (dry-run)。`,
		EN: `Available commands:
  /help          Show this help
  /exit          Exit
  /lang show     Show current language
  /lang set <zh|en>  Switch language
  /mode plan     Switch to plan mode
  /mode act      Switch to act mode
  /mode review   Switch to review mode
  /scan          Scan current project
  /team          Show team suggestion
  Type a task goal to run through the Loop Engineering state machine (dry-run).`,
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
}
