// Package convert provides shared conversion helpers used across adapters.
package convert

import "github.com/moximxxx/shifter/canonical"

// NewLossWarning creates a canonical LossWarning with common fields.
// Reduces duplicated struct construction across 7 adapters.
func NewLossWarning(feature, sourceAgent, targetAgent, reason, severity string) canonical.LossWarning {
	return canonical.LossWarning{
		Feature:     feature,
		Field:       feature,
		SourceAgent: sourceAgent,
		TargetAgent: targetAgent,
		Reason:      reason,
		Severity:    severity,
	}
}
