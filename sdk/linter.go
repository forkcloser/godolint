package sdk

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/shell"
)

// Linter performs Dockerfile linting.
type Linter struct {
	rules []rule.Rule

	// The shellcheck integration WithShellcheck asked for, nil for none.
	shellcheck *shellcheckConfig

	// What the options say of the rules a configuration reaches, and whether
	// any option said anything.
	cfg        *config.Config
	configured bool

	disableIgnorePragma bool
	overrides           map[rule.Code]rule.Severity
	ignored             map[rule.Code]bool

	// The first option given a value it does not take; Lint returns it.
	err error
}

// Option configures a Linter.
type Option func(*Linter)

// WithRuleSet uses a predefined rule set, in place of every rule: the
// error and warning rules for RuleSetRecommended, every rule for RuleSetAll
// and for any other value. The shellcheck integration is no rule of a set:
// WithShellcheck adds it to either, wherever it stands among the options.
func WithRuleSet(set RuleSet) Option {
	return func(l *Linter) {
		l.rules = ruleSet(set)
	}
}

// WithDisabledRules returns an option that disables specific rules by their codes.
// Example: WithDisabledRules("DL3000", "DL3007").
//
// A rule with one of the codes is not run, and no violation with one of
// them reaches a Result, wherever the option stands among the others. A code
// shellcheck reports (SC####) is one too: shellcheck still runs, and its
// findings under that code are dropped.
func WithDisabledRules(codes ...string) Option {
	return func(l *Linter) {
		for _, code := range codes {
			l.ignored[rule.Code(code)] = true
		}
	}
}

// WithSeverity gives every violation with one of the codes the severity, in
// place of its rule's: one of SeverityError, SeverityWarning, SeverityInfo
// and SeverityStyle, as hadolint's override lists. A code given several
// severities gets the gravest. A code shellcheck reports (SC####) is one
// too, and a rule that is off (SeverityIgnore: DL1001, DL3057) is reported
// once it is given one. To silence a rule, WithDisabledRules.
func WithSeverity(severity Severity, codes ...string) Option {
	return func(linter *Linter) {
		grave, ok := map[Severity]rule.Severity{
			SeverityError:   rule.Error,
			SeverityWarning: rule.Warning,
			SeverityInfo:    rule.Info,
			SeverityStyle:   rule.Style,
		}[severity]
		if !ok {
			linter.refuse("WithSeverity", string(severity))

			return
		}

		for _, code := range codes {
			// A graver severity is a smaller one.
			if current, given := linter.overrides[rule.Code(code)]; !given || grave < current {
				linter.overrides[rule.Code(code)] = grave
			}
		}
	}
}

// WithLabelSchema names the labels an image must carry and how the value of
// each is written: DL3049 reports a stage that lacks one, and DL3051 to
// DL3056 and DL3058 a value that is empty or not of its type. The map is
// read when the option is made. Given again, it adds to the schema.
func WithLabelSchema(schema map[string]LabelType) Option {
	types := map[LabelType]config.LabelType{
		LabelText:    config.LabelTypeRawText,
		LabelURL:     config.LabelTypeURL,
		LabelSPDX:    config.LabelTypeSPDX,
		LabelHash:    config.LabelTypeGitHash,
		LabelRFC3339: config.LabelTypeRFC3339,
		LabelSemVer:  config.LabelTypeSemVer,
		LabelEmail:   config.LabelTypeEmail,
	}

	labels := maps.Clone(schema)

	return func(linter *Linter) {
		for _, label := range slices.Sorted(maps.Keys(labels)) {
			labelType, ok := types[labels[label]]
			if !ok {
				linter.refuse("WithLabelSchema", string(labels[label]))

				return
			}

			linter.cfg.LabelSchema[label] = labelType
		}

		linter.configured = true
	}
}

// WithStrictLabels makes a label the schema does not name a violation
// (DL3050).
func WithStrictLabels() Option {
	return func(l *Linter) {
		l.cfg.StrictLabels = true
		l.configured = true
	}
}

// WithTrustedRegistries names the registries an image may come from: DL3026
// reports the image of a FROM, of a COPY --from and of a bind or cache mount
// from any other. A leading `*` matches any registry ending as the rest does
// (`*.example.com`), a trailing `*` any starting as it does, `*` alone any.
// An image with no registry in its name is Docker Hub's, trusted as
// `docker.io` or `hub.docker.com`. Given again, it adds to the list; with no
// list, every registry is trusted.
func WithTrustedRegistries(registries ...string) Option {
	return func(l *Linter) {
		l.cfg.AllowedRegistries = append(l.cfg.AllowedRegistries, registries...)
		l.configured = true
	}
}

// WithDisableIgnorePragma makes the linter read no ignore pragma
// (`# hadolint ignore=`, `stage ignore=`, `global ignore=`): what they would
// silence is reported.
func WithDisableIgnorePragma() Option {
	return func(l *Linter) {
		l.disableIgnorePragma = true
	}
}

// shellcheckConfig collects the shellcheck integration settings.
type shellcheckConfig struct {
	rcFile  string
	exclude []string
}

// ShellcheckOption configures the shellcheck integration enabled by WithShellcheck.
type ShellcheckOption func(*shellcheckConfig)

// WithShellcheckRCFile makes shellcheck use the given configuration file
// (forwarded as --rcfile, requires shellcheck >= 0.10.0). Without it,
// shellcheck never sees a repository's .shellcheckrc: the checked script runs
// from a temp dir, outside the rcfile search path. A file that cannot be
// opened is a value the option does not take: Lint returns a *ConfigError.
func WithShellcheckRCFile(path string) ShellcheckOption {
	return func(c *shellcheckConfig) {
		c.rcFile = path
	}
}

// WithShellcheckExclusions names the checks shellcheck leaves out, in place
// of the three it leaves out by default as hadolint has it: SC2187 (ash is
// not a shell it supports), SC1090 and SC1091 (a sourced file it cannot
// follow). With no code, it leaves out none.
func WithShellcheckExclusions(codes ...string) ShellcheckOption {
	return func(c *shellcheckConfig) {
		c.exclude = slices.Clone(codes)
	}
}

// WithShellcheck enables shellcheck integration for RUN instruction validation.
// Requires the shellcheck binary to be available in PATH. Given again, the
// last one holds, with its options alone.
func WithShellcheck(scOpts ...ShellcheckOption) Option {
	return func(linter *Linter) {
		cfg := shellcheckConfig{exclude: shell.DefaultExclude()}
		for _, opt := range scOpts {
			opt(&cfg)
		}

		// shellcheck warns of a file it cannot read on its standard error,
		// which is not its report, and lints on without it.
		if cfg.rcFile != "" && !opens(cfg.rcFile) {
			linter.refuse("WithShellcheckRCFile", cfg.rcFile)
		}

		linter.shellcheck = &cfg
	}
}

// opens reports whether the file at path can be opened for reading.
func opens(path string) bool {
	file, err := os.Open(
		path,
	) // #nosec G304 G703 -- the configuration file the caller names, as shellcheck would open it
	if err != nil {
		return false
	}

	return file.Close() == nil
}

// New creates a new Linter with the given options.
// By default, it has every implemented rule.
func New(opts ...Option) *Linter {
	lint := &Linter{
		rules:     rules.All(),
		cfg:       config.Default(),
		overrides: map[rule.Code]rule.Severity{},
		ignored:   map[rule.Code]bool{},
	}

	for _, opt := range opts {
		opt(lint)
	}

	// Once every option has spoken: the rules a configuration reaches are
	// made with it, whichever option chose the rules, and shellcheck joins
	// them once.
	if lint.configured {
		for i, r := range lint.rules {
			lint.rules[i] = rules.WithConfig(r, lint.cfg)
		}
	}

	if lint.shellcheck != nil {
		checker := shell.NewBinaryShellchecker()
		checker.RCFile = lint.shellcheck.rcFile
		checker.Exclude = lint.shellcheck.exclude
		lint.rules = append(lint.rules, shell.NewShellcheckRule(checker))
	}

	return lint
}

// Lint lints the given Dockerfile content.
// The context can be used for cancellation. A linter made with an option
// given a value it does not take lints nothing and returns a *ConfigError.
func (l *Linter) Lint(ctx context.Context, dockerfile []byte) (*Result, error) {
	if l.err != nil {
		return nil, l.err
	}

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
	instructions, err := parser.NewBuildkitParser().Parse(dockerfile)
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
	processor := process.NewProcessor(l.rules).
		WithDisableIgnorePragmas(l.disableIgnorePragma).
		WithIgnored(l.ignored)
	failures := process.Settle(processor.Run(instructions), l.overrides)

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

// refuse records the first option given a value it does not take.
func (l *Linter) refuse(option, value string) {
	if l.err == nil {
		l.err = &ConfigError{Option: option, Value: value}
	}
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
		// A rule's own severity only: Lint drops Ignore-severity failures
		// before they reach the conversion.
		return SeverityIgnore
	default:
		return SeverityInfo
	}
}
