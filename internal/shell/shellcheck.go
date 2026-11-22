// This file provides the shellcheck integration for validating shell commands
// in RUN instructions. The package godoc lives in parser.go.

package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/forkcloser/godolint/internal/pragma"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// shellcheckTimeout bounds a single shellcheck invocation: one RUN
// instruction's script is tiny, so this is far beyond any legitimate run.
const shellcheckTimeout = 30 * time.Second

// errNoAnswer is why a shellcheck was killed at the timeout.
var errNoAnswer = errors.New("no answer in " + shellcheckTimeout.String())

// ScanError is a RUN shellcheck did not scan. What the rule found before
// it is no verdict on the Dockerfile: the lint stops on it.
type ScanError struct {
	// Line is the line of the RUN.
	Line int
	// Err is what running shellcheck returned.
	Err error
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("shellcheck did not scan the RUN at line %d: %v", e.Line, e.Err)
}

func (e *ScanError) Unwrap() error {
	return e.Err
}

// Opts contains options for running shellcheck.
// Ported from Hadolint.Shell.ShellOpts.
type Opts struct {
	ShellName string            // Shell command (e.g., "/bin/sh -c")
	EnvVars   map[string]string // Environment variables to export
}

// DefaultOpts returns the default shell options.
// Matches hadolint's defaultShellOpts with common proxy variables.
func DefaultOpts() Opts {
	return Opts{
		ShellName: "/bin/sh -c",
		EnvVars: map[string]string{
			"HTTP_PROXY":  "1",
			"http_proxy":  "1",
			"HTTPS_PROXY": "1",
			"https_proxy": "1",
			"FTP_PROXY":   "1",
			"ftp_proxy":   "1",
			"NO_PROXY":    "1",
			"no_proxy":    "1",
		},
	}
}

// Shellchecker defines the interface for shellcheck integration.
type Shellchecker interface {
	// Check runs shellcheck on a shell script with the given options.
	// Returns violations found by shellcheck. Each failure's Line is the
	// 0-based line offset within the script (0 for its first line); the
	// caller anchors it to the Dockerfile line the script starts on.
	Check(script string, opts Opts) ([]rule.CheckFailure, error)
}

// BinaryShellchecker shells out to the shellcheck binary.
type BinaryShellchecker struct {
	// RCFile, when set, is forwarded to shellcheck as --rcfile so the check
	// uses that configuration instead of searching for one. Without it the
	// script runs from a temp dir, so a repository's .shellcheckrc would
	// never be found. Requires shellcheck >= 0.10.0.
	RCFile string

	// Exclude is the checks shellcheck is told to leave out (--exclude).
	// NewBinaryShellchecker sets the three hadolint leaves out.
	Exclude []string
}

// DefaultExclude returns the checks hadolint has shellcheck leave out:
// SC2187 (ash is not a shell it supports), SC1090 and SC1091 (a sourced file
// it cannot follow).
func DefaultExclude() []string {
	return []string{"SC2187", "SC1090", "SC1091"}
}

// NewBinaryShellchecker creates a shellchecker that uses the shellcheck binary.
func NewBinaryShellchecker() *BinaryShellchecker {
	return &BinaryShellchecker{Exclude: DefaultExclude()}
}

// shellcheckOutput represents the JSON output from shellcheck.
type shellcheckOutput struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	EndLine int    `json:"endLine"`
	Column  int    `json:"column"`
	Level   string `json:"level"` // "error", "warning", "info", "style"
	Code    int    `json:"code"`  // SC code number
	Message string `json:"message"`
}

// isPOSIXShell reports whether a RUN's shell is one shellcheck can read.
// pwsh, powershell and cmd are not shells it knows anything about.
func isPOSIXShell(name string) bool {
	lower := strings.ToLower(name)

	return !strings.Contains(lower, "pwsh") &&
		!strings.Contains(lower, "powershell") &&
		!strings.Contains(lower, "cmd")
}

// Check runs shellcheck on the given script.
// Ported from Hadolint.Shell.shellcheck.
func (c *BinaryShellchecker) Check(script string, opts Opts) ([]rule.CheckFailure, error) {
	// A script for another interpreter, or one that declares an unsupported
	// shebang, is not shellcheck's to judge.
	if !isPOSIXShell(opts.ShellName) || hasUnsupportedShebang(script) {
		return nil, nil
	}

	scResults, err := c.run(buildScript(script, opts))
	if err != nil {
		return nil, err
	}

	return convertFindings(scResults, opts), nil
}

// run hands the script to shellcheck as a file and returns its report.
// shellcheck exits 0 for a scan that found nothing and 1 for one that found
// something, and prints a report for either, `[]` for nothing. Anything
// else is a scan that did not happen, and an error: no binary, a status of
// 2 and above (an option it does not take), a kill, and no report or one
// that does not read, which is what a wrapper that fails in shellcheck's
// place leaves, often with status 1.
func (c *BinaryShellchecker) run(fullScript string) ([]shellcheckOutput, error) {
	tmpFile, err := os.CreateTemp("", "shellcheck-*.sh")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Every path below closes the file before this runs, as windows requires
	// to remove it. A temp file that outlives a failed removal is the OS's to
	// reap, so that error is dropped.
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err = tmpFile.WriteString(fullScript); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to write script: %w", err), tmpFile.Close())
	}

	if err = tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	args := []string{
		"--format=json",
		"--severity=style", // Minimum severity (matches hadolint)
	}
	if len(c.Exclude) > 0 {
		args = append(args, "--exclude="+strings.Join(c.Exclude, ","))
	}

	if c.RCFile != "" {
		args = append(args, "--rcfile="+c.RCFile)
	}

	args = append(args, tmpFile.Name())

	// No caller context reaches this point (the rule fold carries none), but
	// a wedged shellcheck must not hang the whole lint run: bound it.
	ctx, cancel := context.WithTimeoutCause(context.Background(), shellcheckTimeout, errNoAnswer)
	defer cancel()

	cmd := exec.CommandContext(ctx, "shellcheck", args...)

	// A shellcheck behind a wrapper is the wrapper's child: killed at the
	// bound, the wrapper leaves it holding the pipes, and without this the
	// wait goes on until the child ends.
	cmd.WaitDelay = time.Second

	// Stdout only: the JSON report. Stderr is anything else — shellcheck's
	// own diagnostics, or the tool manager's (on Windows, aqua's shim warns
	// about a temporary file it could not remove, in color) — and merging
	// it into the report corrupts the JSON. Keep it for the error message.
	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		exitError := &exec.ExitError{}
		if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
			if cause := context.Cause(ctx); cause != nil {
				err = fmt.Errorf("%w: %w", cause, err)
			}

			return nil, said(fmt.Errorf("failed to run shellcheck: %w", err), &stderr)
		}
	}

	var findings []shellcheckOutput

	if err = json.Unmarshal(output, &findings); err != nil {
		return nil, said(fmt.Errorf("failed to parse shellcheck output: %w", err), &stderr)
	}

	return findings, nil
}

// said is the error with what the program wrote on its standard error, when
// it wrote anything.
func said(err error, stderr *bytes.Buffer) error {
	if text := strings.TrimSpace(stderr.String()); text != "" {
		return fmt.Errorf("%w: %s", err, text)
	}

	return err
}

// convertFindings turns shellcheck's report into rule failures. shellcheck
// reports positions within the synthesized script; subtracting the header
// (shebang + exports) gives the 0-based offset within the original script,
// which the rule anchors to the Dockerfile line the script starts on. Only
// a heredoc's body has lines of its own: a continued RUN is one line by the
// time it is read, and its findings are all on the instruction's line.
func convertFindings(scResults []shellcheckOutput, opts Opts) []rule.CheckFailure {
	headerLines := 1 + len(opts.EnvVars)

	var failures []rule.CheckFailure

	for _, finding := range scResults {
		offset := max(finding.Line-headerLines-1,
			// A finding inside the synthesized header (e.g. on the shebang or
			// an export) has no counterpart in the script — pin to its start.
			0)

		// The column is 1, as for every rule and as in hadolint: shellcheck's
		// counts a tab for eight, and knows nothing of the instruction's
		// keyword, of collapsed continuations or of the tabs <<- strips.
		failures = append(failures, rule.CheckFailure{
			Code:     rule.Code(fmt.Sprintf("SC%d", finding.Code)),
			Severity: convertSeverity(finding.Level),
			Message:  finding.Message,
			Line:     offset, // Anchored to the script's first line by the rule
			Column:   1,
		})
	}

	return failures
}

// buildScript constructs the complete script to pass to shellcheck.
// Ported from the script construction in Hadolint.Shell.shellcheck.
func buildScript(runCommand string, opts Opts) string {
	var build strings.Builder

	// The first line is this function's, not the Dockerfile's: shellcheck
	// reports one that is no absolute path (SC2239), so a shell named
	// without its path (SHELL ["bash", "-c"]) is found through env, which
	// tells shellcheck the same shell.
	shebang := extractShell(opts.ShellName)

	switch {
	case shebang == "":
		shebang = "/bin/sh"
	case !strings.HasPrefix(shebang, "/"):
		shebang = "/usr/bin/env " + Basename(shebang)
	}

	_, _ = build.WriteString("#!")
	_, _ = build.WriteString(shebang)
	_, _ = build.WriteString("\n")

	for key := range opts.EnvVars {
		_, _ = build.WriteString("export ")
		_, _ = build.WriteString(key)
		_, _ = build.WriteString("=1\n")
	}

	_, _ = build.WriteString(runCommand)

	return build.String()
}

// extractShell extracts the shell path from shell command.
// Ported from extractShell in Hadolint.Shell.
func extractShell(shellCmd string) string {
	// Take first word from shell command
	// E.g., "/bin/bash -c" -> "/bin/bash"
	parts := strings.Fields(shellCmd)
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// hasUnsupportedShebang checks if script starts with an unsupported shebang.
// Ported from Hadolint.Shell.hasUnsupportedShebang.
func hasUnsupportedShebang(script string) bool {
	if !strings.HasPrefix(script, "#!") {
		return false
	}

	supported := []string{
		"#!/bin/sh",
		"#!/bin/bash",
		"#!/bin/ksh",
		"#!/usr/bin/env sh",
		"#!/usr/bin/env bash",
		"#!/usr/bin/env ksh",
	}

	for _, prefix := range supported {
		if strings.HasPrefix(script, prefix) {
			return false
		}
	}

	return true
}

// convertSeverity converts shellcheck severity to hadolint severity.
// Ported from Hadolint.Rule.Shellcheck.getDLSeverity.
func convertSeverity(scLevel string) rule.Severity {
	switch scLevel {
	case "warning":
		return rule.Warning
	case "info":
		return rule.Info
	case "style":
		return rule.Style
	default:
		return rule.Error
	}
}

// ShellcheckRule is a stateful rule that runs shellcheck on RUN instructions.
// Ported from Hadolint.Rule.Shellcheck.
type ShellcheckRule struct {
	checker Shellchecker
}

// shellState is the shell in force, per stage. Ported from Acc in
// Hadolint.Rule.Shellcheck.
type shellState struct {
	// opts is the current stage's shell and variables.
	opts Opts
	// defaultOpts is what a stage starts from: the defaults, or what a shell
	// pragma before the first instruction set for every stage.
	defaultOpts Opts
	// stageIdx counts the FROMs seen.
	stageIdx int
	// stageOpts is each stage's final opts, by index, for a later FROM that
	// names it.
	stageOpts map[int]Opts
	// stages is each stage's alias, by index.
	stages map[int]string
	// started is false before any instruction has touched the state:
	// hadolint's Empty. A shell pragma seen then is global.
	started bool
}

// NewShellcheckRule creates a new shellcheck rule.
func NewShellcheckRule(checker Shellchecker) *ShellcheckRule {
	return &ShellcheckRule{
		checker: checker,
	}
}

// Code returns the rule code.
func (*ShellcheckRule) Code() rule.Code {
	return "SHELLCHECK"
}

// InOnBuild says that the rule also reads the RUN an ONBUILD wraps, as
// hadolint's does (`rule = scrule <> onbuild scrule`).
func (*ShellcheckRule) InOnBuild() bool {
	return true
}

// Severity returns the rule severity.
func (*ShellcheckRule) Severity() rule.Severity {
	// Shellcheck violations have their own severities
	return rule.Info
}

// Message returns the rule message.
func (*ShellcheckRule) Message() string {
	return "ShellCheck violations in RUN instructions"
}

// InitialState returns the initial state for this rule.
func (*ShellcheckRule) InitialState() rule.State {
	defaultOpts := DefaultOpts()

	return rule.EmptyState(shellState{
		opts:        defaultOpts,
		defaultOpts: defaultOpts,
		stageIdx:    0,
		stageOpts:   map[int]Opts{0: defaultOpts},
		stages:      map[int]string{},
		started:     false,
	})
}

// Check processes each instruction and updates shell state.
// Ported from scrule in Hadolint.Rule.Shellcheck.
func (r *ShellcheckRule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	// Extract current shell state (InitialState always seeds it; a type
	// mismatch would be an invariant violation and panics in rule.Data).
	shState := rule.Data[shellState](state)

	switch instr := instruction.(type) {
	case *syntax.From:
		return state.ReplaceData(shState.newStage(instr.Image))

	case *syntax.Arg:
		return state.ReplaceData(shState.addVars([]string{instr.ArgName}))

	case *syntax.Env:
		names := make([]string, 0, len(instr.Pairs))
		for _, pair := range instr.Pairs {
			names = append(names, pair.Key)
		}

		return state.ReplaceData(shState.addVars(names))

	case *syntax.Shell:
		if len(instr.Arguments) == 0 {
			return state
		}

		return state.ReplaceData(shState.setShell(strings.Join(instr.Arguments, " ")))

	case *syntax.Comment:
		if name, ok := pragma.ParseShell(instr.Text); ok {
			return state.ReplaceData(shState.shellPragma(name))
		}

		return state

	case *syntax.Run:
		// The exec form runs no shell: nothing for shellcheck to read.
		if instr.IsJSON {
			return state
		}

		violations, err := r.checker.Check(instr.Command, shState.opts)
		if err != nil {
			return state.Stop(&ScanError{Line: line, Err: err})
		}

		// Add all shellcheck violations to state, anchored to the line the
		// script starts on (each violation carries its 0-based offset within
		// the command).
		newState := state

		for _, v := range violations {
			v.Line += line + instr.ScriptOffset
			newState = newState.AddFailure(v)
		}

		return newState

	default:
		// Other instructions do not affect the shell state.
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*ShellcheckRule) Finalize(state rule.State) rule.State {
	return state // No finalization needed
}

// newStage opens the stage a FROM starts: from the stage the image names,
// when it is an earlier stage's alias (the first such stage, as hadolint
// takes it), else from the defaults. Ported from newStage.
func (s shellState) newStage(image syntax.BaseImage) shellState {
	next := s

	if s.started {
		next.stageIdx = s.stageIdx + 1
	}

	next.started = true
	next.opts = s.defaultOpts

	if idx, ok := s.stageNamed(image.Image); ok {
		next.opts = s.stageOpts[idx]
	}

	next.stageOpts = cloneMap(s.stageOpts)
	next.stageOpts[next.stageIdx] = next.opts
	next.stages = cloneMap(s.stages)

	if image.Alias != nil {
		next.stages[next.stageIdx] = *image.Alias
	}

	return next
}

// stageNamed is the lowest stage index whose alias is the name.
func (s shellState) stageNamed(name string) (int, bool) {
	found, ok := -1, false

	for idx, alias := range s.stages {
		if alias == name && (!ok || idx < found) {
			found, ok = idx, true
		}
	}

	return found, ok
}

// addVars adds ARG and ENV names to the current stage's variables. hadolint
// resets the shell to the default here as well; that is a drift in its
// addVars, and a SHELL stays in force across an ENV in this port.
func (s shellState) addVars(names []string) shellState {
	opts := s.opts
	opts.EnvVars = cloneMap(s.opts.EnvVars)

	for _, name := range names {
		opts.EnvVars[name] = "1"
	}

	return s.withOpts(opts, false)
}

// setShell is a SHELL instruction: the current stage's shell.
func (s shellState) setShell(name string) shellState {
	opts := s.opts
	opts.ShellName = name

	return s.withOpts(opts, false)
}

// shellPragma is "# hadolint shell=": the current stage's shell, and every
// stage's when nothing has started a stage yet.
func (s shellState) shellPragma(name string) shellState {
	opts := s.opts
	opts.ShellName = name

	return s.withOpts(opts, !s.started)
}

// withOpts records new opts for the current stage, and as the default for
// the stages to come when asked.
func (s shellState) withOpts(opts Opts, asDefault bool) shellState {
	next := s
	next.started = true
	next.opts = opts
	next.stageOpts = cloneMap(s.stageOpts)
	next.stageOpts[s.stageIdx] = opts

	if asDefault {
		next.defaultOpts = opts
	}

	return next
}

// cloneMap copies a map, a nil one included, so a state never shares a map
// with the state it came from.
func cloneMap[K comparable, V any](src map[K]V) map[K]V {
	dst := make(map[K]V, len(src))
	maps.Copy(dst, src)

	return dst
}

// NoopShellchecker is a no-op implementation for when shellcheck is not available.
type NoopShellchecker struct{}

// NewNoopShellchecker creates a new no-op shellchecker.
func NewNoopShellchecker() *NoopShellchecker {
	return &NoopShellchecker{}
}

// Check always returns nil.
func (*NoopShellchecker) Check(_ string, _ Opts) ([]rule.CheckFailure, error) {
	return nil, nil
}
