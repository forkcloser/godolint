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

// ShellcheckError indicates a RUN that shellcheck, asked for with
// WithShellcheck, did not scan: Lint returns it, and no Result, since a
// Result would say of that RUN what was not checked.
type ShellcheckError struct {
	// Line is the line of the RUN.
	Line int
	// Err is why: no shellcheck in PATH, an exit status that is no scan's
	// (an option shellcheck does not take), no answer in time, no report.
	Err error
}

func (e *ShellcheckError) Error() string {
	return fmt.Sprintf("shellcheck did not scan the RUN at line %d: %v", e.Line, e.Err)
}

func (e *ShellcheckError) Unwrap() error {
	return e.Err
}
