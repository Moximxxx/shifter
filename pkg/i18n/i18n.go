// Package i18n provides internationalization support for Shifter.
//
// Locale files are JSON key-value pairs in pkg/i18n/locales/.
// The active locale is determined by user settings (~/.shifter/settings.json).
//
// Usage:
//
//	T("menu.title")                    // simple key
//	T("menu.found_agents", 3)          // with count replacement {count}
//	T("port.done", "claude", "codex")  // with positional replacements
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	mu       sync.RWMutex
	current  = "en"
	locales  = map[string]map[string]string{}
)

// Init loads all locale files and sets the active language.
func Init(lang string) {
	mu.Lock()
	defer mu.Unlock()

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".json")
		data, err := localeFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			continue
		}
		var kv map[string]string
		if err := json.Unmarshal(data, &kv); err != nil {
			continue
		}
		locales[name] = kv
	}

	if _, ok := locales[lang]; ok {
		current = lang
	}
}

// SetLang switches the active language.
func SetLang(lang string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := locales[lang]; ok {
		current = lang
	}
}

// Lang returns the current language code.
func Lang() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T translates a key with optional format arguments.
// Arguments replace {count}, {0}, {1}, etc. in the translated string.
// Named args like {name} are replaced if a matching key=value pair is found.
func T(key string, args ...interface{}) string {
	mu.RLock()
	msg := getMsg(current, key)
	mu.RUnlock()

	if msg == "" {
		// Fallback to English
		mu.RLock()
		msg = getMsg("en", key)
		mu.RUnlock()
	}
	if msg == "" {
		return key
	}

	// Simple positional replacement
	for i, arg := range args {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%d}", i), fmt.Sprint(arg))
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%v}", arg), fmt.Sprint(arg))
	}

	// Handle common named placeholders
	if len(args) == 1 {
		msg = strings.ReplaceAll(msg, "{count}", fmt.Sprint(args[0]))
	}

	return msg
}

// Tf translates with named arguments via a map.
func Tf(key string, vars map[string]string) string {
	mu.RLock()
	msg := getMsg(current, key)
	mu.RUnlock()

	if msg == "" {
		mu.RLock()
		msg = getMsg("en", key)
		mu.RUnlock()
	}
	if msg == "" {
		return key
	}

	for k, v := range vars {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	return msg
}

func getMsg(lang, key string) string {
	if l, ok := locales[lang]; ok {
		if msg, ok := l[key]; ok {
			return msg
		}
	}
	return ""
}

// Supported returns the list of supported language codes.
func Supported() []string {
	mu.RLock()
	defer mu.RUnlock()
	var langs []string
	for k := range locales {
		langs = append(langs, k)
	}
	return langs
}
