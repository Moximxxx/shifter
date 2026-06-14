// Package env provides environment variable configuration for Shifter.
//
// All CLI flags have corresponding SHIFTER_* environment variables.
// CLI flags take precedence over environment variables.
package env

import (
	"os"
	"strings"
)

// Supported environment variables:
//
//   SHIFTER_SOURCE        Default source agent (e.g., claude-code)
//   SHIFTER_TARGET        Default target agent (e.g., codex)
//   SHIFTER_SCOPE         Default scope: project | global | all
//   SHIFTER_PROJECT_ROOT  Default project root directory
//   SHIFTER_DRY_RUN       Set to "1" or "true" to enable dry-run by default
//   SHIFTER_NO_BACKUP     Set to "1" or "true" to skip automatic backups
//   SHIFTER_FORCE         Set to "1" or "true" to skip confirmation prompts
//   SHIFTER_ASPECTS       Default aspects to port (comma-separated)
//   SHIFTER_PROFILE       Default profile name for save/load
//   SHIFTER_STRATEGY      Default merge strategy: newer | merge | source
//   SHIFTER_PRETTY        Set to "0" or "false" to disable pretty output

// Get returns the value of an environment variable, or the default if not set.
func Get(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Bool returns true if the environment variable is set to a truthy value.
func Bool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

// Source returns the default source agent from SHIFTER_SOURCE.
func Source() string { return Get("SHIFTER_SOURCE", "") }

// Target returns the default target agent from SHIFTER_TARGET.
func Target() string { return Get("SHIFTER_TARGET", "") }

// Scope returns the default scope from SHIFTER_SCOPE.
func Scope() string { return Get("SHIFTER_SCOPE", "project") }

// ProjectRoot returns the default project root from SHIFTER_PROJECT_ROOT.
func ProjectRoot() string { return Get("SHIFTER_PROJECT_ROOT", ".") }

// DryRun returns true if SHIFTER_DRY_RUN is set to a truthy value.
func DryRun() bool { return Bool("SHIFTER_DRY_RUN") }

// NoBackup returns true if SHIFTER_NO_BACKUP is set to a truthy value.
func NoBackup() bool { return Bool("SHIFTER_NO_BACKUP") }

// Force returns true if SHIFTER_FORCE is set to a truthy value.
func Force() bool { return Bool("SHIFTER_FORCE") }

// Aspects returns the default aspects from SHIFTER_ASPECTS.
func Aspects() []string {
	val := Get("SHIFTER_ASPECTS", "")
	if val == "" {
		return nil
	}
	var result []string
	for _, a := range strings.Split(val, ",") {
		a = strings.TrimSpace(a)
		if a != "" {
			result = append(result, a)
		}
	}
	return result
}

// Profile returns the default profile name from SHIFTER_PROFILE.
func Profile() string { return Get("SHIFTER_PROFILE", "") }

// Strategy returns the default merge strategy from SHIFTER_STRATEGY.
func Strategy() string { return Get("SHIFTER_STRATEGY", "newer") }

// Pretty returns true if pretty output is enabled (default: true).
func Pretty() bool {
	val := strings.ToLower(os.Getenv("SHIFTER_PRETTY"))
	if val == "" {
		return true
	}
	return val != "0" && val != "false" && val != "no" && val != "off"
}

// Config returns a summary of all Shifter environment variables and their current values.
func Config() map[string]string {
	vars := map[string]string{
		"SHIFTER_SOURCE":       os.Getenv("SHIFTER_SOURCE"),
		"SHIFTER_TARGET":       os.Getenv("SHIFTER_TARGET"),
		"SHIFTER_SCOPE":        os.Getenv("SHIFTER_SCOPE"),
		"SHIFTER_PROJECT_ROOT": os.Getenv("SHIFTER_PROJECT_ROOT"),
		"SHIFTER_DRY_RUN":      os.Getenv("SHIFTER_DRY_RUN"),
		"SHIFTER_NO_BACKUP":    os.Getenv("SHIFTER_NO_BACKUP"),
		"SHIFTER_FORCE":        os.Getenv("SHIFTER_FORCE"),
		"SHIFTER_ASPECTS":      os.Getenv("SHIFTER_ASPECTS"),
		"SHIFTER_PROFILE":      os.Getenv("SHIFTER_PROFILE"),
		"SHIFTER_STRATEGY":     os.Getenv("SHIFTER_STRATEGY"),
	}
	// Remove empty values
	for k, v := range vars {
		if v == "" {
			delete(vars, k)
		}
	}
	return vars
}
