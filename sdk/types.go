package sdk

// Severity represents the severity level of a violation.
type Severity string

const (
	// SeverityError indicates a critical issue that should be fixed.
	SeverityError Severity = "error"
	// SeverityWarning indicates a potential issue.
	SeverityWarning Severity = "warning"
	// SeverityInfo indicates informational feedback.
	SeverityInfo Severity = "info"
	// SeverityStyle indicates a style preference.
	SeverityStyle Severity = "style"
	// SeverityIgnore is the severity of a rule that is off: it runs, and its
	// violations never reach a Result. Only a RuleInfo carries it.
	SeverityIgnore Severity = "ignore"
)

// RuleInfo describes a rule, as Rules and Rule return it.
type RuleInfo struct {
	// Code is the rule's code, DL and four digits (e.g., "DL3006"). No two
	// rules share one.
	Code string `json:"code"`
	// Severity is the severity the rule declares, and the Severity of every
	// Violation it reports.
	Severity Severity `json:"severity"`
	// Message is the rule's description, in hadolint's words, and the Message
	// of every Violation it reports, but for a rule about a label: its Message
	// says `<label>` where a Violation's names the label ("Label `<label>` is
	// missing."). Any other name in angle brackets (`<package>`, `<version>`)
	// is the Violation's text too.
	Message string `json:"message"`
}

// Violation represents a single linting violation.
type Violation struct {
	// Code is the rule code (e.g., "DL3000", "SC2086").
	Code string `json:"code"`
	// Severity is the violation severity level.
	Severity Severity `json:"severity"`
	// Message is the human-readable description.
	Message string `json:"message"`
	// Line is the line number in the Dockerfile (1-indexed).
	Line int `json:"line"`
}

// Result contains the linting results.
type Result struct {
	// Violations contains all detected violations.
	Violations []Violation
	// Passed indicates whether the Dockerfile passed linting (no violations).
	Passed bool
}

// HasErrors returns true if any violations have Error severity.
func (r *Result) HasErrors() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityError {
			return true
		}
	}

	return false
}

// HasWarnings returns true if any violations have Warning severity.
func (r *Result) HasWarnings() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityWarning {
			return true
		}
	}

	return false
}

// CountBySeverity returns the count of violations for each severity level.
func (r *Result) CountBySeverity() map[Severity]int {
	counts := make(map[Severity]int)
	for _, v := range r.Violations {
		counts[v.Severity]++
	}

	return counts
}
