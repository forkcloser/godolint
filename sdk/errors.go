// Package sdk provides a high-level API for linting Dockerfiles with godolint.
package sdk

import "fmt"

// ParseError indicates a Dockerfile parsing failure.
type ParseError struct {
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse Dockerfile: %v", e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// RuleError indicates a rule execution failure.
type RuleError struct {
	RuleCode string
	Err      error
}

func (e *RuleError) Error() string {
	return fmt.Sprintf("rule %s failed: %v", e.RuleCode, e.Err)
}

func (e *RuleError) Unwrap() error {
	return e.Err
}

// ConfigError indicates an option given a value it does not take: Lint
// returns it, and lints nothing.
type ConfigError struct {
	// Option is the option's name, e.g. "WithSeverity".
	Option string
	// Value is what it was given.
	Value string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("%s does not take %q", e.Option, e.Value)
}
