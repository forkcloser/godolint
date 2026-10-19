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

	"github.com/alecthomas/kong"
	"github.com/mycophonic/primordium/app/logger"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/sdk"
)

// errViolations is how Run tells main that the lint found something: the
// failures are already on stdout, and the exit status is 1 without a log line.
var errViolations = errors.New("violations found")

// CLI is the kong grammar: the flags and the Dockerfile arguments. Paths stay
// exactly as given — no kong file type, whose tilde and absolute-path
// expansion would rewrite the File field of every reported failure.
type CLI struct {
	// A blank line ends a tag-alignment block (tagalign pads keys per block),
	// so the one flag with more tag keys than help stands alone.
	LogLevel string `default:"info" enum:"debug,info,warn,error" env:"LOG_LEVEL" help:"Log verbosity (stderr)."`

	DisableIgnorePragma bool     `help:"Disable the ignore pragmas (# hadolint ignore=DLxxxx, stage ignore=, global ignore=)."`
	WithoutShellcheck   bool     `help:"Disable shellcheck integration for RUN instruction validation."`
	Ignore              []string `help:"Rule code to ignore (repeatable: --ignore DL3006 --ignore SC2050)."                                                placeholder:"CODE"`
	ShellcheckRcfile    string   `help:"Shellcheckrc forwarded to shellcheck (--rcfile) when validating RUN instructions (requires shellcheck >= 0.10.0)." placeholder:"FILE"`

	Dockerfiles []string `arg:"" help:"Dockerfile(s) to lint." name:"dockerfile"`
}

// Run lints every Dockerfile, writes the failures to stdout as JSON, and
// returns errViolations when there were any.
func (c *CLI) Run() error {
	rules, err := c.buildRules()
	if err != nil {
		return err
	}

	// One processor, every rule, reused across files.
	processor := process.NewProcessor(rules).WithDisableIgnorePragmas(c.DisableIgnorePragma)

	allFailures, err := lintFiles(processor, c.Dockerfiles)
	if err != nil {
		return err
	}

	allFailures = dropIgnored(allFailures, c.Ignore)

	if err := json.NewEncoder(os.Stdout).Encode(allFailures); err != nil {
		return fmt.Errorf("failed to encode failures: %w", err)
	}

	if len(allFailures) > 0 {
		return errViolations
	}

	return nil
}

// buildRules assembles the rule set, wiring in the shellcheck integration
// unless it is disabled or the binary is missing from PATH.
func (c *CLI) buildRules() ([]rule.Rule, error) {
	rules := sdk.AllRules()

	if c.WithoutShellcheck {
		return rules, nil
	}

	// Missing shellcheck degrades gracefully (matching hadolint): warn and
	// lint without the integration rather than failing the run.
	//nolint:nilerr // intentional, see above.
	if _, err := exec.LookPath("shellcheck"); err != nil {
		slog.Warn("shellcheck binary not found in PATH, shellcheck integration disabled")

		return rules, nil
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

	return append(rules, shell.NewShellcheckRule(checker)), nil
}

// lintFiles runs the processor over each Dockerfile and returns the collected
// failures, each tagged with the file it came from.
func lintFiles(processor *process.Processor, paths []string) ([]rule.CheckFailure, error) {
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

		failures := process.WithoutIgnored(processor.Run(instructions))
		for i := range failures {
			failures[i].File = dockerfilePath
		}

		allFailures = append(allFailures, failures...)
	}

	return allFailures, nil
}

// dropIgnored filters out failures whose rule code was --ignore'd.
func dropIgnored(failures []rule.CheckFailure, ignoredRules []string) []rule.CheckFailure {
	if len(ignoredRules) == 0 {
		return failures
	}

	filtered := []rule.CheckFailure{}

	for _, failure := range failures {
		if !slices.Contains(ignoredRules, string(failure.Code)) {
			filtered = append(filtered, failure)
		}
	}

	return filtered
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
