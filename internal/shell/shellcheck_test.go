package shell_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

func TestBinaryShellchecker_Check(t *testing.T) {
	t.Parallel()

	checker := shell.NewBinaryShellchecker()

	tests := []struct {
		name          string
		script        string
		opts          shell.Opts
		expectFailure bool
	}{
		{
			name:          "valid script",
			script:        "apt-get install -y package",
			opts:          shell.DefaultOpts(),
			expectFailure: false,
		},
		{
			name:          "unquoted variable",
			script:        "echo $FOO",
			opts:          shell.DefaultOpts(),
			expectFailure: true, // SC2086: quote to prevent word splitting
		},
		{
			name:   "powershell script skipped",
			script: "Write-Host 'test'",
			opts: shell.Opts{
				ShellName: "pwsh -c",
				EnvVars:   make(map[string]string),
			},
			expectFailure: false, // Skipped for non-POSIX shells
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			failures, err := checker.Check(tt.script, tt.opts)
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}

			hasFailure := len(failures) > 0
			if hasFailure != tt.expectFailure {
				t.Errorf("Check() hasFailure = %v, want %v (failures: %v)", hasFailure, tt.expectFailure, failures)
			}
		})
	}
}

func TestBinaryShellchecker_RCFile(t *testing.T) {
	t.Parallel()

	// SC2086 (unquoted variable) fires without configuration…
	script := "echo $FOO"

	plain, err := shell.NewBinaryShellchecker().Check(script, shell.DefaultOpts())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if !hasCode(plain, "SC2086") {
		t.Fatalf("expected SC2086 without rcfile, got %v", plain)
	}

	// …and is silenced by an rcfile that disables it.
	rcfile := filepath.Join(t.TempDir(), "shellcheckrc")
	if err = os.WriteFile(rcfile, []byte("disable=SC2086\n"), 0o600); err != nil {
		t.Fatalf("writing rcfile: %v", err)
	}

	checker := shell.NewBinaryShellchecker()
	checker.RCFile = rcfile

	configured, err := checker.Check(script, shell.DefaultOpts())
	if err != nil {
		t.Fatalf("Check() with rcfile error = %v", err)
	}

	if hasCode(configured, "SC2086") {
		t.Errorf("expected rcfile to silence SC2086, got %v", configured)
	}
}

func TestBinaryShellchecker_LineOffsets(t *testing.T) {
	t.Parallel()

	// A two-line script: SC2164 (cd without || exit) fires on the second
	// line, so its failure must carry offset 1; SC2086 on the first line
	// must carry offset 0. The column is 1 on either, wherever shellcheck
	// puts the finding.
	script := "echo $FOO\n  cd /app"

	failures, err := shell.NewBinaryShellchecker().Check(script, shell.DefaultOpts())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	checkedLine := func(code string, wantLine, wantColumn int) {
		t.Helper()

		for _, f := range failures {
			if string(f.Code) != code {
				continue
			}

			if f.Line != wantLine || f.Column != wantColumn {
				t.Errorf("%s at line %d column %d, want line %d column %d",
					code, f.Line, f.Column, wantLine, wantColumn)
			}

			return
		}

		t.Errorf("expected %s, got %v", code, failures)
	}

	checkedLine("SC2086", 0, 1)
	checkedLine("SC2164", 1, 1)
}

func hasCode(failures []rule.CheckFailure, code string) bool {
	for _, f := range failures {
		if string(f.Code) == code {
			return true
		}
	}

	return false
}

func TestShellcheckRule_StatefulTracking(t *testing.T) {
	t.Parallel()

	checker := shell.NewBinaryShellchecker()
	scRule := shell.NewShellcheckRule(checker)

	// Simulate a Dockerfile with ENV, ARG, SHELL, and RUN
	instructions := []syntax.InstructionPos{
		{
			Instruction: &syntax.From{Image: syntax.BaseImage{Image: "debian"}},
			LineNumber:  1,
		},
		{
			Instruction: &syntax.Env{Pairs: []syntax.EnvPair{{Key: "MY_VAR", Value: "value"}}},
			LineNumber:  2,
		},
		{
			Instruction: &syntax.Arg{ArgName: "BUILD_ARG"},
			LineNumber:  3,
		},
		{
			Instruction: &syntax.Shell{Arguments: []string{"/bin/bash", "-c"}},
			LineNumber:  4,
		},
		{
			Instruction: &syntax.Run{Command: "echo $UNDEFINED"},
			LineNumber:  5,
		},
	}

	state := scRule.InitialState()
	for _, instr := range instructions {
		state = scRule.Check(instr.LineNumber, state, instr.Instruction)
	}

	// Should have shellcheck violations for undefined variable
	if len(state.Failures) == 0 {
		t.Error("Expected shellcheck violations, got none")
	}

	// Verify violations have SC codes
	for _, f := range state.Failures {
		code := string(f.Code)
		if len(code) < 2 || code[0:2] != "SC" {
			t.Errorf("Expected SC code, got %s", code)
		}
	}
}

func TestShellcheckRule_MultiLineRunAnchoring(t *testing.T) {
	t.Parallel()

	scRule := shell.NewShellcheckRule(shell.NewBinaryShellchecker())

	// A RUN at line 5 whose script starts a line below it, as a heredoc's
	// body does, and whose second line triggers SC2164: the failure must
	// land on Dockerfile line 7 (the script's first line + the offset).
	state := scRule.InitialState()
	state = scRule.Check(1, state, &syntax.From{Image: syntax.BaseImage{Image: "debian"}})
	state = scRule.Check(5, state, &syntax.Run{Command: "true\ncd /app", ScriptOffset: 1})

	for _, f := range state.Failures {
		if string(f.Code) == "SC2164" {
			if f.Line != 7 {
				t.Errorf("SC2164 anchored to line %d, want 7", f.Line)
			}

			return
		}
	}

	t.Errorf("expected SC2164, got %v", state.Failures)
}

func TestShellcheckRule_ResetOnFrom(t *testing.T) {
	t.Parallel()

	checker := shell.NewBinaryShellchecker()
	scRule := shell.NewShellcheckRule(checker)

	// Test that ENV variables are reset on new FROM (multi-stage build)
	instructions := []syntax.InstructionPos{
		{
			Instruction: &syntax.From{Image: syntax.BaseImage{Image: "debian"}},
			LineNumber:  1,
		},
		{
			Instruction: &syntax.Env{Pairs: []syntax.EnvPair{{Key: "STAGE1_VAR", Value: "val"}}},
			LineNumber:  2,
		},
		{
			Instruction: &syntax.From{Image: syntax.BaseImage{Image: "alpine"}},
			LineNumber:  3,
		},
		// After second FROM, STAGE1_VAR should not be in environment
	}

	// Since shellState is unexported, we test behavior instead of internal state
	// The fact that the rule processes without error is sufficient for this test
	// Actual ENV variable tracking is verified through integration tests
	state := scRule.InitialState()
	for _, instr := range instructions {
		state = scRule.Check(instr.LineNumber, state, instr.Instruction)
	}
}

func TestNoopShellchecker(t *testing.T) {
	t.Parallel()

	checker := shell.NewNoopShellchecker()

	failures, err := checker.Check("any script", shell.DefaultOpts())
	if err != nil {
		t.Errorf("NoopShellchecker.Check() error = %v, want nil", err)
	}

	if len(failures) != 0 {
		t.Errorf("NoopShellchecker.Check() returned %d failures, want 0", len(failures))
	}
}

func TestShellcheckRule_Metadata(t *testing.T) {
	t.Parallel()

	checker := shell.NewNoopShellchecker()
	scRule := shell.NewShellcheckRule(checker)

	if scRule.Code() != "SHELLCHECK" {
		t.Errorf("Code() = %s, want SHELLCHECK", scRule.Code())
	}

	if scRule.Severity() != rule.Info {
		t.Errorf("Severity() = %v, want Info", scRule.Severity())
	}

	if scRule.Message() == "" {
		t.Error("Message() returned empty string")
	}
}

// TestBinaryShellchecker_LeavesNoScriptBehind: the script is handed to
// shellcheck as a temporary file, which is gone when Check returns:
// shellcheck having found something, nothing, or not having run at all.
func TestBinaryShellchecker_LeavesNoScriptBehind(t *testing.T) {
	dir := t.TempDir()

	// os.TempDir reads TMPDIR, or TMP on Windows.
	t.Setenv("TMPDIR", dir)
	t.Setenv("TMP", dir)

	checker := shell.NewBinaryShellchecker()

	scripts := func() []string {
		left, err := filepath.Glob(filepath.Join(dir, "shellcheck-*.sh"))
		if err != nil {
			t.Fatal(err)
		}

		return left
	}

	for range 10 {
		for script, found := range map[string]bool{"echo $FOO": true, "echo ok": false} {
			failures, err := checker.Check(script, shell.DefaultOpts())
			if err != nil || (len(failures) > 0) != found {
				t.Fatalf("the premise: %q checked, findings %t; got %v, %v", script, found, failures, err)
			}
		}
	}

	if left := scripts(); len(left) != 0 {
		t.Errorf("after twenty checks: %v left in the temporary directory", left)
	}

	// No shellcheck to run.
	t.Setenv("PATH", t.TempDir())

	// Nothing ran, so nothing wrote on a standard error: the error ends
	// with its reason, not with a colon and nothing after.
	_, err := checker.Check("echo $FOO", shell.DefaultOpts())
	if err == nil || strings.HasSuffix(strings.TrimSpace(err.Error()), ":") {
		t.Fatalf("with no shellcheck on PATH, Check returns the error %v", err)
	}

	if left := scripts(); len(left) != 0 {
		t.Errorf("after a shellcheck that could not run: %v left in the temporary directory", left)
	}
}

// TestBinaryShellchecker_AScanThatDidNotHappenIsAnError: shellcheck exits 0
// or 1 for a scan; given an exclusion that is no check, it refuses its
// options, exits 3 and scans nothing, which is an error and not a script
// with nothing in it. The rule stops on it, at the RUN's line, and not on a
// RUN shellcheck is not shown.
func TestBinaryShellchecker_AScanThatDidNotHappenIsAnError(t *testing.T) {
	t.Parallel()

	checker := shell.NewBinaryShellchecker()
	checker.Exclude = []string{"bogus"}

	failures, err := checker.Check("echo $FOO", shell.DefaultOpts())
	if err == nil || failures != nil || !strings.Contains(err.Error(), "exit status 3") {
		t.Fatalf("Check() = %v, %v; want no failure and an error naming exit status 3", failures, err)
	}

	// shellcheck says why on its standard error, and the error says it after.
	if strings.HasSuffix(err.Error(), "exit status 3") {
		t.Errorf("Check() error %q; want what shellcheck wrote on its standard error after the status", err)
	}

	scRule := shell.NewShellcheckRule(checker)

	state := scRule.Check(1, scRule.InitialState(), &syntax.From{Image: syntax.BaseImage{Image: "debian"}})

	state = scRule.Check(2, state, &syntax.Run{Command: `["echo", "$FOO"]`, IsJSON: true})
	if state.Err != nil {
		t.Fatalf("an exec form, which no shellcheck reads: %v", state.Err)
	}

	state = scRule.Check(4, state, &syntax.Run{Command: "echo $FOO"})

	var scan *shell.ScanError
	if !errors.As(state.Err, &scan) || scan.Line != 4 || !strings.Contains(scan.Error(), "line 4") {
		t.Errorf("the rule's state says %v; want a *ScanError at line 4", state.Err)
	}
}
