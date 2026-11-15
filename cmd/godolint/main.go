// Package main implements the godolint CLI for linting Dockerfiles.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mycophonic/primordium/app/logger"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/sdk"
)

// errViolations is how Run tells main that the lint found something: the
// failures are already on stdout, and the exit status is 1 without a log line.
var errViolations = errors.New("violations found")

// What the command line may say that a setting does not take.
var (
	errLabelType = errors.New("not a label type (text, url, spdx, hash, rfc3339, semver, email)")
	errThreshold = errors.New("not a severity (error, warning, info, style, ignore, none)")
)

// CLI is the kong grammar: the flags and the Dockerfile arguments. Paths stay
// exactly as given — no kong file type, whose tilde and absolute-path
// expansion would rewrite the File field of every reported failure.
//
// The settings are hadolint's, under hadolint's flags; what a flag does not
// say, the configuration file may (see settle).
type CLI struct {
	// A blank line ends a tag-alignment block (tagalign pads keys per block),
	// so the one flag with more tag keys than help stands alone.
	LogLevel string `default:"info" enum:"debug,info,warn,error" env:"LOG_LEVEL" help:"Log verbosity (stderr)."`

	Config string `help:"Path to the configuration file (default: the first of .hadolint.yaml and its other places that exists)." placeholder:"FILENAME" short:"c"`

	FailureThreshold string `help:"Exit with a failure status only when a rule of this severity or a graver one is violated: error, warning, info, style, ignore or none (default: info)." placeholder:"THRESHOLD" short:"t"`

	NoFail              bool     `help:"Do not exit with a failure status when a rule is violated."`
	NoColor             bool     `help:"Do not colour the output (the JSON output has none)."`
	Verbose             bool     `help:"Say on stderr which configuration file is used."                                                                   short:"V"`
	DisableIgnorePragma bool     `help:"Disable the ignore pragmas (# hadolint ignore=DLxxxx, stage ignore=, global ignore=)."`
	StrictLabels        bool     `help:"Do not permit labels other than those of the label schema."`
	WithoutShellcheck   bool     `help:"Disable shellcheck integration for RUN instruction validation."`
	Error               []string `help:"Make the rule's violations errors (repeatable)."                                                                   placeholder:"RULECODE"`
	Warning             []string `help:"Make the rule's violations warnings (repeatable)."                                                                 placeholder:"RULECODE"`
	Info                []string `help:"Make the rule's violations infos (repeatable)."                                                                    placeholder:"RULECODE"`
	Style               []string `help:"Make the rule's violations style remarks (repeatable)."                                                            placeholder:"RULECODE"`
	Ignore              []string `help:"Rule code to ignore (repeatable: --ignore DL3006 --ignore SC2050)."                                                placeholder:"CODE"`
	TrustedRegistry     []string `help:"A registry images may come from (repeatable; e.g. docker.io, *.example.com)."                                      placeholder:"REGISTRY"`
	RequireLabel        []string `help:"A label every image must carry, and how its value is written (repeatable; e.g. maintainer:email)."                 placeholder:"LABEL:TYPE"`
	ShellcheckRcfile    string   `help:"Shellcheckrc forwarded to shellcheck (--rcfile) when validating RUN instructions (requires shellcheck >= 0.10.0)." placeholder:"FILE"`

	Dockerfiles []string `arg:"" help:"Dockerfile(s) to lint." name:"dockerfile"`
}

// settings is what a run obeys: the configuration file's word, then the
// command line's.
type settings struct {
	rules               *config.Config
	overrides           map[rule.Code]rule.Severity
	ignored             map[rule.Code]bool
	threshold           rule.Severity
	noFail              bool
	disableIgnorePragma bool
}

// Run lints every Dockerfile, writes the failures to stdout as JSON, and
// returns errViolations when one of them is at the failure threshold or
// graver, unless failing is off.
func (c *CLI) Run() error {
	settled, err := c.settle()
	if err != nil {
		return err
	}

	ruleSet, err := c.buildRules(settled.rules)
	if err != nil {
		return err
	}

	// One processor, every rule, reused across files.
	processor := process.NewProcessor(ruleSet).WithDisableIgnorePragmas(settled.disableIgnorePragma)

	allFailures, err := lintFiles(processor, c.Dockerfiles, settled)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(os.Stdout).Encode(allFailures); err != nil {
		return fmt.Errorf("failed to encode failures: %w", err)
	}

	if settled.noFail {
		return nil
	}

	// A graver severity is a smaller one.
	if slices.ContainsFunc(allFailures, func(failure rule.CheckFailure) bool {
		return failure.Severity <= settled.threshold
	}) {
		return errViolations
	}

	return nil
}

// settle reads the configuration file, the one --config names or the one
// hadolint would find, and lays the command line over it as hadolint does:
// a list is the file's and the command line's together, a switch is on when
// either turns it on, the failure threshold is the command line's when it
// gives one, and a label both name keeps the file's type.
func (c *CLI) settle() (*settings, error) {
	file, err := c.configFile()
	if err != nil {
		return nil, err
	}

	settled := &settings{
		rules:               config.Default(),
		overrides:           map[rule.Code]rule.Severity{},
		ignored:             map[rule.Code]bool{},
		threshold:           rule.Info,
		noFail:              c.NoFail || isSet(file.NoFail),
		disableIgnorePragma: c.DisableIgnorePragma || isSet(file.DisableIgnorePragma),
	}

	settled.rules.StrictLabels = c.StrictLabels || isSet(file.StrictLabels)
	settled.rules.AllowedRegistries = slices.Concat(file.TrustedRegistries, c.TrustedRegistry)

	// The gravest severity a code is given holds, so the lists go from the
	// lightest to the gravest and each overwrites the one before.
	for _, override := range []struct {
		severity rule.Severity
		codes    []string
	}{
		{rule.Style, slices.Concat(file.Override.Style, c.Style)},
		{rule.Info, slices.Concat(file.Override.Info, c.Info)},
		{rule.Warning, slices.Concat(file.Override.Warning, c.Warning)},
		{rule.Error, slices.Concat(file.Override.Error, c.Error)},
	} {
		for _, code := range override.codes {
			settled.overrides[rule.Code(code)] = override.severity
		}
	}

	for _, code := range slices.Concat(file.Ignored, c.Ignore) {
		settled.ignored[rule.Code(code)] = true
	}

	if err := c.requireLabels(file, settled.rules.LabelSchema); err != nil {
		return nil, err
	}

	threshold := c.FailureThreshold
	if threshold == "" && file.FailureThreshold != nil {
		threshold = *file.FailureThreshold
	}

	if threshold != "" {
		var ok bool
		if settled.threshold, ok = config.ParseSeverity(threshold); !ok {
			return nil, fmt.Errorf("%w: --failure-threshold %s", errThreshold, threshold)
		}
	}

	return settled, nil
}

// requireLabels puts in schema the labels the command line asks for, each
// LABEL:TYPE with text for no type, then the file's: a label both name keeps
// the file's type, as in hadolint, whose schema is a map joined with the
// file's on the left. The file's types were checked when it was read; the
// command line's are here.
func (c *CLI) requireLabels(file *config.File, schema map[string]config.LabelType) error {
	for _, required := range c.RequireLabel {
		label, name, _ := strings.Cut(required, ":")

		labelType, ok := config.ParseLabelType(name)
		if !ok {
			return fmt.Errorf("%w: --require-label %s", errLabelType, required)
		}

		schema[label] = labelType
	}

	for label, name := range file.LabelSchema {
		if name != nil {
			schema[label], _ = config.ParseLabelType(*name)
		}
	}

	return nil
}

// configFile is what the configuration file says: the one --config names,
// which must be readable, or the first hadolint would find, or nothing. With
// --verbose, or verbose in the file, which one is said on stderr, in
// hadolint's words.
func (c *CLI) configFile() (*config.File, error) {
	path, found := c.Config, c.Config != ""
	if !found {
		path, found = config.Find()
	}

	file := &config.File{}

	if found {
		var err error
		if file, err = config.Read(path); err != nil {
			return nil, fmt.Errorf("configuration: %w", err)
		}
	}

	if c.Verbose || isSet(file.Verbose) {
		used := "No configuration was specified. Using default configuration"
		if found {
			used = "Configuration file used: " + path
		}

		fmt.Fprintln(os.Stderr, used)
	}

	return file, nil
}

// isSet reports whether a switch of the configuration file is there and on.
func isSet(value *bool) bool {
	return value != nil && *value
}

// buildRules assembles the rule set, wiring in the shellcheck integration
// unless it is disabled or the binary is missing from PATH.
func (c *CLI) buildRules(cfg *config.Config) ([]rule.Rule, error) {
	ruleSet := sdk.AllRules()
	for i, given := range ruleSet {
		ruleSet[i] = rules.WithConfig(given, cfg)
	}

	if c.WithoutShellcheck {
		return ruleSet, nil
	}

	// Missing shellcheck degrades gracefully (matching hadolint): warn and
	// lint without the integration rather than failing the run.
	//nolint:nilerr // intentional, see above.
	if _, err := exec.LookPath("shellcheck"); err != nil {
		slog.Warn("shellcheck binary not found in PATH, shellcheck integration disabled")

		return ruleSet, nil
	}

	checker := shell.NewBinaryShellchecker()
	// Fail fast on an unreadable rcfile: shellcheck errors are non-fatal per
	// rule (matching hadolint), so a bad path would otherwise silently
	// disable every SC check.
	if c.ShellcheckRcfile != "" {
		if _, err := os.Stat(c.ShellcheckRcfile); err != nil {
			return nil, fmt.Errorf("cannot read shellcheck rcfile: %w", err)
		}

		checker.RCFile = c.ShellcheckRcfile
	}

	return append(ruleSet, shell.NewShellcheckRule(checker)), nil
}

// lintFiles runs the processor over each Dockerfile and returns the collected
// failures, each tagged with the file it came from.
func lintFiles(processor *process.Processor, paths []string, settled *settings) ([]rule.CheckFailure, error) {
	// Non-nil so an all-clean run still encodes as JSON [] rather than null.
	allFailures := []rule.CheckFailure{}

	for _, dockerfilePath := range paths {
		// #nosec G304 -- reading user-supplied Dockerfile paths is this tool's purpose.
		dockerfileContent, err := os.ReadFile(dockerfilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", dockerfilePath, err)
		}

		instructions, err := parser.NewBuildkitParser().Parse(dockerfileContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", dockerfilePath, err)
		}

		slog.Debug("parsed Dockerfile", "file", dockerfilePath, "instructions", len(instructions))

		failures := process.Settle(processor.Run(instructions), settled.overrides, settled.ignored)
		for i := range failures {
			failures[i].File = dockerfilePath
		}

		allFailures = append(allFailures, failures...)
	}

	return allFailures, nil
}

// configureLogger installs the process-wide slog handler through primordium:
// tint on a terminal, JSON when stderr is redirected. The level is the
// --log-level flag, whose enum kong has already validated.
func configureLogger(level string) {
	slogLevel := slog.LevelInfo

	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		// info, the enum's remaining value.
	}

	logger.SetDefaultsForLogger(context.Background(), slogLevel)
}

func main() {
	var cli CLI

	// Parse and usage errors are kong's: printed with the usage. Every
	// failure exits 1 — kong's own usage-error status (80) is folded into
	// the one non-zero status this tool has always had; --help stays 0.
	cmdline := kong.Must(&cli,
		kong.Name("godolint"),
		kong.Description("Dockerfile linter"),
		kong.UsageOnError(),
		kong.Exit(func(code int) {
			if code != 0 {
				code = 1
			}

			os.Exit(code)
		}),
	)

	kctx, err := cmdline.Parse(os.Args[1:])
	if err != nil {
		// The usage goes with the error, to stderr: stdout is the JSON only.
		// --help, handled inside Parse, keeps stdout.
		cmdline.Stdout = os.Stderr
		cmdline.FatalIfErrorf(err)
	}

	configureLogger(cli.LogLevel)

	if err := kctx.Run(); err != nil {
		if errors.Is(err, errViolations) {
			os.Exit(1)
		}

		slog.Error("failed to run godolint", "err", err)
		os.Exit(1)
	}
}
