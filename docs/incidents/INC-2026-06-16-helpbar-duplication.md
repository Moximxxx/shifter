# INC-2026-06-16: Help Bar Key Prefix Duplication

**Date**: 2026-06-16
**Severity**: LOW (cosmetic)
**Status**: RESOLVED

## Symptom
TUI help bars showed duplicated key prefixes:
`↑↓ ↑↓ 浏览 • Enter Enter 选择 • Esc Esc 返回`

## Root Cause
i18n translation keys already include key prefixes (e.g., `help.select = "Enter 选择"`), 
but Go code also hardcoded prefixes: `"Enter " + i18n.T("help.select")`.

## Affected Areas
- `tui/wizard.go`: viewTemplates (line 879), empty templates state (line 856)
- Multiple view functions had this pattern in earlier versions

## Resolution
- Removed all hardcoded key prefixes (`Enter `, `Esc `, `Space `, `↑↓ `) from Go code
- i18n keys now contain the complete text including key hints
- All help bars use only `i18n.T()` concatenation with ` • ` separators

## Prevention
- CI check: `grep -rn '"Enter\|"Esc\|" ↑↓' tui/` should return empty
- Code review: verify new help bars don't duplicate key names
