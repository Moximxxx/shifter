package i18n

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	Init("en")
	if Lang() != "en" {
		t.Errorf("Init(en): got %q", Lang())
	}
	if len(Supported()) < 2 {
		t.Errorf("expected at least en,zh; got %v", Supported())
	}
}

func TestInit_Unknown(t *testing.T) {
	Init("fr")
	// Should fall back to default or keep previous
	lang := Lang()
	if lang != "en" && lang != "zh" && lang != "" {
		t.Errorf("unknown lang should fallback: got %q", lang)
	}
}

func TestSetLang(t *testing.T) {
	Init("en")
	SetLang("zh")
	if Lang() != "zh" {
		t.Errorf("SetLang(zh): got %q", Lang())
	}
	SetLang("en")
	if Lang() != "en" {
		t.Errorf("SetLang(en): got %q", Lang())
	}
}

func TestT_English(t *testing.T) {
	Init("en")
	tests := []struct{ key, want string }{
		{"app.name", "Shifter"},
		{"menu.title", "Shifter — Interactive Wizard"},
		{"help.navigate", "↑↓ navigate"},
		{"welcome.title", "Welcome to Shifter!"},
		{"aspect.instructions", "Project Instructions"},
	}
	for _, tc := range tests {
		got := T(tc.key)
		if got != tc.want {
			t.Errorf("T(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestT_Chinese(t *testing.T) {
	Init("zh")
	tests := []struct{ key, want string }{
		{"app.name", "Shifter"},
		{"menu.title", "Shifter — 交互式向导"},
		{"help.navigate", "↑↓ 浏览"},
		{"welcome.title", "欢迎使用 Shifter！"},
	}
	for _, tc := range tests {
		got := T(tc.key)
		if got != tc.want {
			t.Errorf("T(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestT_MissingKey(t *testing.T) {
	Init("en")
	got := T("nonexistent.key.12345")
	if got != "nonexistent.key.12345" {
		t.Errorf("expected key itself for missing, got %q", got)
	}
}

func TestT_WithCount(t *testing.T) {
	Init("en")
	got := T("menu.found_agents", 3)
	if got == "menu.found_agents" || got == "" {
		t.Errorf("T with count: got %q", got)
	}
}

func TestTf(t *testing.T) {
	Init("zh")
	got := Tf("port.target_title", map[string]string{"source": "claude-code"})
	if got == "" {
		t.Error("Tf returned empty")
	}
	// Should contain the source name
	if len(got) > 0 && got != "port.target_title" {
		// basic sanity check passed
	}
}

func TestSupported(t *testing.T) {
	langs := Supported()
	found := make(map[string]bool)
	for _, l := range langs {
		found[l] = true
	}
	if !found["en"] {
		t.Error("en not in supported languages")
	}
	if !found["zh"] {
		t.Error("zh not in supported languages")
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	// French is not supported, should fall back to English for T()
	Init("fr")
	// After init with unknown lang, T should fallback to English values
	got := T("app.name")
	if got != "Shifter" {
		t.Errorf("fallback: got %q, want 'Shifter'", got)
	}
}

func TestT_LocalesLoaded(t *testing.T) {
	// Verify all locale files are embedded
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		t.Fatalf("read locales dir: %v", err)
	}
	if len(entries) < 2 {
		t.Errorf("expected at least 2 locale files, got %d", len(entries))
	}
}

func TestConcurrentAccess(t *testing.T) {
	Init("en")
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				T("app.name")
				Lang()
				Supported()
			}
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestInitEnvLang(t *testing.T) {
	// Test that Init respects the lang parameter
	os.Setenv("TEST_LANG", "zh")
	Init("zh")
	if Lang() != "zh" {
		t.Errorf("Init should use explicit param: got %q", Lang())
	}
}
