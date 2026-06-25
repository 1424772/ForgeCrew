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
}
