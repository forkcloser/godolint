package sdk

import (
	"context"
	"fmt"
	"os"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
)

// Linter performs Dockerfile linting.
type Linter struct {
	parser parser.Parser
	rules  []rule.Rule
}

// Option configures a Linter.
type Option func(*Linter)

// WithParser sets a custom parser implementation.
// By default, uses the buildkit parser.
func WithParser(p parser.Parser) Option {
	return func(l *Linter) {
		l.parser = p
	}
}

// WithRules sets custom rules to use for linting.
// By default, uses AllRules().
func WithRules(rules []rule.Rule) Option {
	return func(l *Linter) {
		l.rules = rules
	}
}

// WithRuleSet uses a predefined rule set.
func WithRuleSet(set RuleSet) Option {
	return func(l *Linter) {
		l.rules = GetRuleSet(set)
	}
}

// WithDisabledRules returns an option that disables specific rules by their codes.
// Example: WithDisabledRules("DL3000", "DL3007").
func WithDisabledRules(codes ...string) Option {
	return func(l *Linter) {
		l.rules = FilterRules(l.rules, codes)
	}
}

// shellcheckConfig collects the shellcheck integration settings.
type shellcheckConfig struct {
	rcFile string
}

// ShellcheckOption configures the shellcheck integration enabled by WithShellcheck.
type ShellcheckOption func(*shellcheckConfig)

// WithShellcheckRCFile makes shellcheck use the given configuration file
// (forwarded as --rcfile, requires shellcheck >= 0.10.0). Without it,
// shellcheck never sees a repository's .shellcheckrc: the checked script runs
// from a temp dir, outside the rcfile search path.
func WithShellcheckRCFile(path string) ShellcheckOption {
	return func(c *shellcheckConfig) {
		c.rcFile = path
	}
}

// WithShellcheck enables shellcheck integration for RUN instruction validation.
// Requires the shellcheck binary to be available in PATH.
func WithShellcheck(scOpts ...ShellcheckOption) Option {
	return func(linter *Linter) {
		cfg := shellcheckConfig{}
		for _, opt := range scOpts {
			opt(&cfg)
		}

		checker := shell.NewBinaryShellchecker()
		checker.RCFile = cfg.rcFile
		scRule := shell.NewShellcheckRule(checker)
		linter.rules = append(linter.rules, scRule)
	}
}

// New creates a new Linter with the given options.
// By default, uses all implemented rules and the buildkit parser.
func New(opts ...Option) *Linter {
	lint := &Linter{
		parser: parser.NewBuildkitParser(),
		rules:  AllRules(),
	}

	for _, opt := range opts {
		opt(lint)
	}

	return lint
}

// Lint lints the given Dockerfile content.
// The context can be used for cancellation.
func (l *Linter) Lint(ctx context.Context, dockerfile []byte) (*Result, error) {
	// Check context cancellation before parsing
	select {
	case <-ctx.Done():
		// context.Canceled/DeadlineExceeded are already the canonical
		// sentinels: callers match them with errors.Is, like the stdlib.
		//nolint:wrapcheck // bare ctx.Err() is the idiomatic cancellation contract.
		return nil, ctx.Err()
	default:
	}

	// Parse Dockerfile
	instructions, err := l.parser.Parse(dockerfile)
	if err != nil {
		return nil, &ParseError{Err: err}
	}

	// Check context cancellation before processing
	select {
	case <-ctx.Done():
		//nolint:wrapcheck // bare ctx.Err(), see the pre-parse check above.
		return nil, ctx.Err()
	default:
	}

	// Run rules
	processor := process.NewProcessor(l.rules)
	failures := process.WithoutIgnored(processor.Run(instructions))

	// Convert to SDK violations
	violations := make([]Violation, len(failures))
	for i, f := range failures {
		violations[i] = Violation{
			Code:     string(f.Code),
			Severity: convertSeverity(f.Severity),
			Message:  f.Message,
			Line:     f.Line,
		}
	}

	result := &Result{
		Violations: violations,
		Passed:     len(violations) == 0,
	}

	return result, nil
}

// LintFile reads the Dockerfile at path and lints it as Lint does; a path
// that cannot be read is an error wrapping the file system's.
func (l *Linter) LintFile(ctx context.Context, path string) (*Result, error) {
	dockerfile, err := os.ReadFile(path) // #nosec G304 -- the path is the caller's, this method's one argument
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return l.Lint(ctx, dockerfile)
}

func convertSeverity(s rule.Severity) Severity {
	// Info is also the deliberate fallback for severities this switch does not
	// know, so the default branch intentionally mirrors the Info case.
	//revive:disable:identical-switch-branches intentional, see above.
	switch s {
	case rule.Error:
		return SeverityError
	case rule.Warning:
		return SeverityWarning
	case rule.Info:
		return SeverityInfo
	case rule.Style:
		return SeverityStyle
	case rule.Ignore:
		// Unreachable in practice: Lint drops Ignore-severity
		// failures before they reach the SDK conversion.
		return SeverityInfo
	default:
		return SeverityInfo
	}
}
