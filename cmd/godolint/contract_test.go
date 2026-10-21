// The command's contract from the outside: the binary built once, run on
// Dockerfiles written for each case, judged by its exit status, the JSON
// on stdout and nothing else there.
package main_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// contractBinary is the command built once for every case.
//
//nolint:gochecknoglobals // built in TestMain, read by every test
var contractBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "godolint-contract")
	if err != nil {
		panic(err)
	}

	defer func() { _ = os.RemoveAll(dir) }()

	contractBinary = filepath.Join(dir, "godolint"+exeSuffix())

	build := exec.CommandContext(context.Background(), "go", "build", "-o", contractBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic(string(out))
	}

	m.Run()
}

type failure struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// run runs the binary with args: exit status, the failures decoded from
// stdout, stderr.
func run(t *testing.T, args ...string) (int, []failure, string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), contractBinary, args...)

	cmd.Env = append(os.Environ(), "LOG_LEVEL=info")

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0

	var exit *exec.ExitError

	switch {
	case err == nil:
	case errors.As(err, &exit):
		code = exit.ExitCode()
	default:
		t.Fatalf("running %s: %v", contractBinary, err)
	}

	var failures []failure
	if stdout.Len() > 0 {
		if err := json.Unmarshal([]byte(stdout.String()), &failures); err != nil {
			t.Fatalf("stdout is not a JSON array of failures: %v\n%s", err, stdout.String())
		}
	}

	return code, failures, stderr.String()
}

func write(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// TestContractJSONOnStdoutExitOneOnViolations: stdout is one JSON array of
// failures, each with the file, the 1-based line, column 1, the level,
// the code and the message; exit 1 when there is any, 0 and an empty
// array when there is none; several files are reported in one array,
// each failure naming its file.
func TestContractJSONOnStdoutExitOneOnViolations(t *testing.T) {
	t.Parallel()

	bad := write(t, "Dockerfile", "FROM ubuntu\nMAINTAINER x\n")
	good := write(t, "Dockerfile.good", "FROM debian:bookworm\nCMD [\"true\"]\n")

	code, failures, _ := run(t, "--without-shellcheck", bad)
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}

	var untagged, maintainer *failure

	for i := range failures {
		if failures[i].File != bad || failures[i].Column != 1 || failures[i].Message == "" {
			t.Errorf("failure %+v: want the file, column 1 and a message", failures[i])
		}

		switch failures[i].Code {
		case "DL3006":
			untagged = &failures[i]
		case "DL4000":
			maintainer = &failures[i]
		}
	}

	if untagged == nil || untagged.Line != 1 || untagged.Level != "warning" {
		t.Errorf("DL3006: %+v, want line 1 warning", untagged)
	}

	if maintainer == nil || maintainer.Line != 2 || maintainer.Level != "error" {
		t.Errorf("DL4000: %+v, want line 2 error", maintainer)
	}

	code, failures, _ = run(t, "--without-shellcheck", good)
	if code != 0 || len(failures) != 0 {
		t.Errorf("a clean file: exit %d, %v", code, failures)
	}

	code, failures, _ = run(t, "--without-shellcheck", good, bad)
	if code != 1 || len(failures) == 0 {
		t.Errorf("two files: exit %d, %d failures", code, len(failures))
	}

	for _, f := range failures {
		if f.File != bad {
			t.Errorf("a failure named for %s, want %s", f.File, bad)
		}
	}
}

// TestContractIgnoreFlagAndPragmas: --ignore drops a code from the
// report; the file's pragmas drop theirs; --disable-ignore-pragma turns
// the pragmas off and the pragma's own line is then reported by nothing
// the CLI prints (DL1001 is ignore-severity).
func TestContractIgnoreFlagAndPragmas(t *testing.T) {
	t.Parallel()

	codesOf := func(fs []failure) []string {
		out := make([]string, 0, len(fs))
		for _, f := range fs {
			out = append(out, f.Code)
		}

		slices.Sort(out)

		return slices.Compact(out)
	}

	file := write(t, "Dockerfile", "FROM ubuntu\nMAINTAINER x\n")

	_, failures, _ := run(t, "--without-shellcheck", "--ignore", "DL3006", file)
	if got := codesOf(failures); slices.Contains(got, "DL3006") || !slices.Contains(got, "DL4000") {
		t.Errorf("--ignore DL3006: %v", got)
	}

	code, failures, _ := run(t, "--without-shellcheck", "--ignore", "DL3006", "--ignore", "DL4000", file)
	if code != 0 || len(failures) != 0 {
		t.Errorf("--ignore of every code: exit %d, %v", code, failures)
	}

	pragma := write(t, "Dockerfile", "# hadolint ignore=DL3006\nFROM ubuntu\n")

	code, failures, _ = run(t, "--without-shellcheck", pragma)
	if code != 0 || len(failures) != 0 {
		t.Errorf("an ignore pragma: exit %d, %v", code, failures)
	}

	code, failures, _ = run(t, "--without-shellcheck", "--disable-ignore-pragma", pragma)
	if code != 1 || !slices.Contains(codesOf(failures), "DL3006") || slices.Contains(codesOf(failures), "DL1001") {
		t.Errorf("--disable-ignore-pragma: exit %d, %v", code, codesOf(failures))
	}
}

// TestContractShellcheckIntegration: with shellcheck on PATH a RUN's
// shell finding is reported under its SC code at the RUN's line;
// --without-shellcheck leaves it out.
func TestContractShellcheckIntegration(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("shellcheck"); err != nil {
		t.Skip("no shellcheck on PATH")
	}

	file := write(t, "Dockerfile", "FROM debian:bookworm\nRUN echo $undefined_var_quoting\n")

	_, failures, _ := run(t, file)

	var sc []failure

	for _, f := range failures {
		if strings.HasPrefix(f.Code, "SC") {
			sc = append(sc, f)
		}
	}

	if len(sc) == 0 {
		t.Fatalf("no SC finding for an unquoted expansion: %v", failures)
	}

	for _, f := range sc {
		if f.Line != 2 || f.File != file {
			t.Errorf("%+v: want line 2 in %s", f, file)
		}
	}

	_, failures, _ = run(t, "--without-shellcheck", file)
	for _, f := range failures {
		if strings.HasPrefix(f.Code, "SC") {
			t.Errorf("--without-shellcheck reported %+v", f)
		}
	}
}

// TestContractErrorsAndLogsOnStderr: a missing file and a file with no
// instruction are errors on stderr with exit 1 and nothing on stdout;
// logs never reach stdout; no file is a usage error.
func TestContractErrorsAndLogsOnStderr(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "nonesuch")

	code, failures, stderr := run(t, "--without-shellcheck", missing)
	if code != 1 || failures != nil || stderr == "" {
		t.Errorf("a missing file: exit %d, stdout %v, stderr %q", code, failures, stderr)
	}

	empty := write(t, "Dockerfile", "# nothing\n")

	code, failures, stderr = run(t, "--without-shellcheck", empty)
	if code != 1 || failures != nil || !strings.Contains(stderr, "no instructions") {
		t.Errorf("a file with no instruction: exit %d, stdout %v, stderr %q", code, failures, stderr)
	}

	good := write(t, "Dockerfile", "FROM debian:bookworm\n")

	cmd := exec.CommandContext(t.Context(), contractBinary, "--without-shellcheck", "--log-level", "debug", good)

	var stdout strings.Builder

	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(stdout.String()) != "[]" {
		t.Errorf("stdout with debug logging: %q, want the empty array alone", stdout.String())
	}

	// No file is a usage error, exit 1; kong prints the usage on stdout,
	// so it is not decoded here.
	usage := exec.CommandContext(t.Context(), contractBinary, "--without-shellcheck")

	var exit *exec.ExitError
	if err := usage.Run(); !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Errorf("no file: %v, want exit 1", err)
	}
}

// exeSuffix is what the OS needs appended to run a binary built to a bare name.
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}

	return ""
}
