package i18n

import (
	"strings"
	"testing"
)

func TestValidLocale(t *testing.T) {
	tests := []struct {
		input   string
		wantOK  bool
		wantLoc Locale
	}{
		{"zh", true, ZH},
		{"en", true, EN},
		{"fr", false, ""},
		{"", false, ""},
		{"ZH", false, ""},
		{"EN", false, ""},
	}
	for _, tt := range tests {
		loc, err := ValidLocale(tt.input)
		if tt.wantOK && err != nil {
			t.Errorf("ValidLocale(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.wantOK && err == nil {
			t.Errorf("ValidLocale(%q) expected error", tt.input)
		}
		if tt.wantOK && loc != tt.wantLoc {
			t.Errorf("ValidLocale(%q) = %q, want %q", tt.input, loc, tt.wantLoc)
		}
	}
}

func TestT_KnownKey(t *testing.T) {
	zh := T("lang.current", ZH)
	if !strings.Contains(zh, "当前") {
		t.Errorf("ZH translation for lang.current should contain Chinese, got: %s", zh)
	}
	en := T("lang.current", EN)
	if !strings.Contains(en, "Current") {
		t.Errorf("EN translation for lang.current should contain English, got: %s", en)
	}
}

func TestT_MissingKey(t *testing.T) {
	result := T("nonexistent.key.xyz", ZH)
	if result != "nonexistent.key.xyz" {
		t.Errorf("missing key should return key itself, got: %s", result)
	}
}

func TestSupportedLocales(t *testing.T) {
	locales := SupportedLocales()
	if len(locales) != 2 {
		t.Errorf("expected 2 supported locales, got %d", len(locales))
	}
}

func TestT_TaskHeader(t *testing.T) {
	zh := T("task.header", ZH)
	if !strings.Contains(zh, "任务") {
		t.Errorf("ZH task header should contain 任务, got: %s", zh)
	}
	en := T("task.header", EN)
	if !strings.Contains(en, "Task") {
		t.Errorf("EN task header should contain Task, got: %s", en)
	}
}

func TestT_Iteration(t *testing.T) {
	zh := T("task.iteration", ZH)
	if !strings.Contains(zh, "第") {
		t.Errorf("ZH iteration should contain 第, got: %s", zh)
	}
	en := T("task.iteration", EN)
	if !strings.Contains(en, "Iteration") {
		t.Errorf("EN iteration should contain Iteration, got: %s", en)
	}
}

func TestT_ValidateMessages(t *testing.T) {
	if s := T("validate.all_ok", ZH); !strings.Contains(s, "通过") {
		t.Errorf("ZH validate.all_ok should contain 通过, got: %s", s)
	}
	if s := T("validate.all_ok", EN); !strings.Contains(s, "passed") {
		t.Errorf("EN validate.all_ok should contain passed, got: %s", s)
	}
}
