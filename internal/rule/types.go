// Package rule defines the core types for hadolint rules.
// Ported from Hadolint/Rule.hs
package rule

import (
	"encoding/json"
	"fmt"

	"github.com/forkcloser/godolint/internal/syntax"
)

// Severity is ported from DLSeverity in Hadolint/Rule.hs.
type Severity int

// Severity levels for rule violations.
const (
	Error Severity = iota
	Warning
	Info
	Style
	Ignore
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Info:
		return "info"
	case Style:
		return "style"
	case Ignore:
		return "ignore"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler to output severity as string.
// Matches hadolint's severityText function in Hadolint/Formatter/Format.hs.
func (s Severity) MarshalJSON() ([]byte, error) {
	// Marshaling a plain string cannot fail, and encoding/json wraps
	// json.Marshaler errors in *json.MarshalerError itself.
	//nolint:wrapcheck // unreachable error path, and the caller wraps.
	return json.Marshal(s.String())
}

// Code is ported from RuleCode in Hadolint/Rule.hs.
type Code string

// Meta contains metadata extracted from hadolint rule definitions.
// Used by generated code to separate metadata from implementation.
type Meta struct {
	Code     Code
	Severity Severity
	Message  string
}

// CheckFailure is ported from CheckFailure in Hadolint/Rule.hs.
// JSON field order and names match hadolint's Json.hs formatter.
type CheckFailure struct {
	File     string   `json:"file,omitempty"` // File path (optional, for multi-file linting)
	Line     int      `json:"line"`
	Column   int      `json:"column"` // Always 1 for lint violations (matches hadolint)
	Severity Severity `json:"level"`  // Outputs as string: "error", "warning", "info", "style"
	Code     Code     `json:"code"`
	Message  string   `json:"message"`
}

// Fail is the failure a rule reports about itself at a line: the describing
// triple, with the column hadolint always emits.
func Fail(r Rule, line int) CheckFailure {
	return CheckFailure{
		Code:     r.Code(),
		Severity: r.Severity(),
		Message:  r.Message(),
		Line:     line,
		Column:   1, // Hardcoded to 1 (matches hadolint)
	}
}

// State holds failures and custom state for a rule.
// Ported from State a in Hadolint/Rule.hs.
type State struct {
	Failures []CheckFailure
	Data     any // Custom state data
	// Err is why the rule could not do its work on the Dockerfile, nil when
	// it could. Its failures are then no verdict: the run stops and returns
	// the error.
	Err error
}

// EmptyState creates a new state with no failures and the given initial data.
func EmptyState(data any) State {
	return State{
		Failures: nil,
		Data:     data,
	}
}

// AddFailure adds a failure to the state.
func (s State) AddFailure(failure CheckFailure) State {
	return State{
		Failures: append(s.Failures, failure),
		Data:     s.Data,
		Err:      s.Err,
	}
}

// ReplaceData replaces the custom state data.
func (s State) ReplaceData(data any) State {
	return State{
		Failures: s.Failures,
		Data:     data,
		Err:      s.Err,
	}
}

// Stop records that the rule could not do its work, and why.
func (s State) Stop(err error) State {
	return State{
		Failures: s.Failures,
		Data:     s.Data,
		Err:      err,
	}
}

// Data returns the custom state data as T. The processor threads each rule's
// state only through that same rule's Check/Finalize, so Data always holds
// what the rule's own InitialState/ReplaceData put there — a mismatch is an
// invariant violation (a bug in the rule itself) and panics loudly rather
// than silently degrading into wrong lint results.
func Data[T any](s State) T {
	data, ok := s.Data.(T)
	if !ok {
		panic(fmt.Sprintf("rule state corrupted: %T is not the expected state type", s.Data))
	}

	return data
}

// Rule is ported from the concept of Rule in Hadolint/Rule.hs.
// All rules are stateful - simple rules just use empty state.
//
//nolint:interfacebloat // what a rule is (code, severity, message) and what it does (initial state, check, finalize), which every rule implements together
type Rule interface {
	// Code returns the unique rule identifier
	Code() Code

	// Severity returns the severity level of violations
	Severity() Severity

	// Message returns the human-readable description of the rule
	Message() string

	// InitialState returns the initial state for this rule
	InitialState() State

	// Check examines an instruction with the current state and returns updated state.
	// Ported from the step function in customRule.
	Check(line int, state State, instruction syntax.Instruction) State

	// Finalize processes the final state after all instructions have been checked.
	// Returns the updated state with any final failures.
	// Ported from markFailures in veryCustomRule.
	// Most rules don't need this and can just return the state unchanged.
	Finalize(state State) State
}

// SimpleRule is ported from simpleRule in Hadolint/Rule.hs.
type SimpleRule struct {
	code     Code
	severity Severity
	message  string
	checker  func(syntax.Instruction) bool
}

// NewSimpleRule creates a new simple rule.
func NewSimpleRule(
	code Code,
	severity Severity,
	message string,
	checker func(syntax.Instruction) bool,
) *SimpleRule {
	return &SimpleRule{
		code:     code,
		severity: severity,
		message:  message,
		checker:  checker,
	}
}

// Code returns the rule code.
func (r *SimpleRule) Code() Code {
	return r.code
}

// Severity returns the rule severity level.
func (r *SimpleRule) Severity() Severity {
	return r.severity
}

// Message returns the rule violation message.
func (r *SimpleRule) Message() string {
	return r.message
}

// InitialState returns an empty state for stateless rules.
func (*SimpleRule) InitialState() State {
	return EmptyState(nil)
}

// Check executes the rule checker function against the instruction.
func (r *SimpleRule) Check(line int, state State, instruction syntax.Instruction) State {
	if !r.checker(instruction) {
		// Checker failed, add failure to state
		return state.AddFailure(CheckFailure{
			Code:     r.code,
			Severity: r.severity,
			Message:  r.message,
			Line:     line,
			Column:   1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*SimpleRule) Finalize(state State) State {
	return state // Simple rules don't need finalization
}

// StatefulRuleBase provides default implementations for stateful rules.
// Embed this in stateful rule structs to avoid boilerplate.
type StatefulRuleBase struct {
	meta Meta
}

// NewStatefulRuleBase creates a base for stateful rules.
func NewStatefulRuleBase(meta Meta) StatefulRuleBase {
	return StatefulRuleBase{meta: meta}
}

// Code returns the rule code.
func (b *StatefulRuleBase) Code() Code {
	return b.meta.Code
}

// Severity returns the rule severity level.
func (b *StatefulRuleBase) Severity() Severity {
	return b.meta.Severity
}

// Message returns the rule violation message.
func (b *StatefulRuleBase) Message() string {
	return b.meta.Message
}

// Finalize performs final checks after processing all instructions.
// Default implementation does nothing. Override in rules that need custom finalization.
func (*StatefulRuleBase) Finalize(state State) State {
	return state
}
