// The command's contract from the outside: the binary built once, run on
// Dockerfiles written for each case, judged by its exit status, the JSON
// on stdout and nothing else there.
package main_test

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
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

	// The build has a deadline as every process these tests start has.
	building, done := context.WithTimeout(context.Background(), buildDeadline)
	defer done()

	build := exec.CommandContext(building, "go", "build", "-o", contractBinary, ".")
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

// run runs the binary with args, in a directory and with a home that hold
// no configuration file: exit status, the failures decoded from stdout,
// stderr.
func run(t *testing.T, args ...string) (int, []failure, string) {
	t.Helper()

	empty := t.TempDir()

	return runIn(t, empty, homeEnv(empty), args...)
}

// homeEnv is the environment that makes home the user's home, and the only
// place beside the working directory a configuration file is looked for in.
func homeEnv(home string) []string {
	return []string{
		"HOME=" + home,
		"XDG_CONFIG_HOME=",
		"APPDATA=" + filepath.Join(home, "AppData"),
		"USERPROFILE=" + home,
	}
}

// How long the build of the binary may take, and one run of it.
const (
	buildDeadline = 5 * time.Minute
	childDeadline = 20 * time.Second
)

// command is a run of the binary with args, killed at its own deadline. A
// test binary stopped at its timeout does not take its children with it,
// and a command that never ends would run on after it, holding what it
// holds: so no child's life hangs on the test's context alone, and the
// deadline is far under any timeout a test binary is given.
func command(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), childDeadline)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, contractBinary, args...)
	// A shellcheck the command started may still hold its pipes.
	cmd.WaitDelay = time.Second

	return cmd
}

// runIn runs the binary with args in dir, env laid over the environment.
func runIn(t *testing.T, dir string, env []string, args ...string) (int, []failure, string) {
	t.Helper()

	cmd := command(t, args...)

	cmd.Dir = dir

	cmd.Env = append(append(os.Environ(), "LOG_LEVEL=info"), env...)

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
// --without-shellcheck leaves it out; --shellcheck-rcfile names the
// configuration file shellcheck reads, and one that cannot be read is an
// error and no lint.
func TestContractShellcheckIntegration(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("shellcheck"); err != nil {
		t.Skip("no shellcheck on PATH")
	}

	file := write(t, "Dockerfile", "FROM debian:bookworm\nRUN echo $undefined_var_quoting\n")

	// The working directory and the home stay the test's own here: a
	// shellcheck that is a shim finds the real one through them. A
	// configuration is named so that none is looked for there.
	_, failures, _ := runIn(t, "", nil, "--config", write(t, "hadolint.yaml", ""), file)

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

	reports := func(failures []failure, code string) bool {
		return slices.ContainsFunc(failures, func(f failure) bool { return f.Code == code })
	}

	if !reports(sc, "SC2086") {
		t.Fatalf("the premise: SC2086 for an unquoted expansion, got %v", sc)
	}

	config := write(t, "hadolint.yaml", "")

	_, failures, _ = runIn(
		t,
		"",
		nil,
		"--config",
		config,
		"--shellcheck-rcfile",
		write(t, "shellcheckrc", "disable=SC2086\n"),
		file,
	)
	if reports(failures, "SC2086") {
		t.Errorf("SC2086 disabled in the file --shellcheck-rcfile names: %v", failures)
	}

	// The name alone is looked for: the error is logged as JSON, where a
	// Windows path is not written as it is given.
	const absent = "no-such-shellcheckrc"

	status, failures, stderr := runIn(
		t,
		"",
		nil,
		"--config",
		config,
		"--shellcheck-rcfile",
		filepath.Join(t.TempDir(), absent),
		file,
	)
	if status == 0 || len(failures) != 0 || !strings.Contains(stderr, absent) {
		t.Errorf(
			"a configuration file that is not there: exit %d, %d failures, stderr %q; want an error naming it and no lint",
			status,
			len(failures),
			stderr,
		)
	}

	const closed = "closed-shellcheckrc"

	unreadable := write(t, closed, "disable=SC2086\n")
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}

	// Windows, and root anywhere, open a file whose mode forbids it.
	if opened, err := os.Open(unreadable); err == nil {
		_ = opened.Close()

		return
	}

	status, failures, stderr = runIn(t, "", nil, "--config", config, "--shellcheck-rcfile", unreadable, file)
	if status == 0 || len(failures) != 0 || !strings.Contains(stderr, closed) {
		t.Errorf(
			"a configuration file that cannot be opened: exit %d, %d failures, stderr %q; want an error naming it and no lint",
			status,
			len(failures),
			stderr,
		)
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

	cmd := command(t, "--without-shellcheck", "--log-level", "debug", good)

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
	usage := command(t, "--without-shellcheck")

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

// TestContractFailuresAreOrdered: the failures of a file come ordered by
// line, and by code on one line, the same on every run; with several files,
// file by file in the order they are given.
func TestContractFailuresAreOrdered(t *testing.T) {
	t.Parallel()

	// Two stages with no HEALTHCHECK, which DL3057 keeps in no order of
	// its own, and a line three rules report.
	first := write(
		t,
		"Dockerfile",
		"FROM debian:latest AS one\nFROM debian:latest\nWORKDIR app\nRUN apt-get install curl && sudo true\n",
	)
	second := write(t, "Dockerfile.second", "FROM ubuntu\nMAINTAINER x\n")

	order := func(failures []failure) []string {
		var out []string
		for _, f := range failures {
			out = append(out, fmt.Sprintf("%s:%d %s", filepath.Base(f.File), f.Line, f.Code))
		}

		return out
	}

	_, failures, _ := run(t, "--without-shellcheck", "--info=DL3057", second, first)
	want := order(failures)

	// The files in the order given, the second and then the first; within
	// one, by line and then by code.
	given := map[string]int{second: 0, first: 1}

	sorted := slices.Clone(failures)
	slices.SortStableFunc(sorted, func(left, right failure) int {
		return cmp.Or(
			cmp.Compare(given[left.File], given[right.File]),
			cmp.Compare(left.Line, right.Line),
			cmp.Compare(left.Code, right.Code),
		)
	})

	if !slices.Equal(want, order(sorted)) {
		t.Errorf("the failures come in the order\n%q\nwant\n%q", want, order(sorted))
	}

	for _, code := range []string{"Dockerfile:1 DL3057", "Dockerfile:2 DL3057", "Dockerfile:4 DL3004", "Dockerfile:4 DL3008"} {
		if !slices.Contains(want, code) {
			t.Fatalf("the run lacks %s, which the order is to be seen on: %q", code, want)
		}
	}

	for range 20 {
		if _, again, _ := run(
			t,
			"--without-shellcheck",
			"--info=DL3057",
			second,
			first,
		); !slices.Equal(
			order(again),
			want,
		) {
			t.Fatalf("another run gives another order:\n%q\nthe first gave\n%q", order(again), want)
		}
	}
}

// shellcheckExiting is a directory holding a shellcheck that scans nothing,
// prints the word on its standard output when there is one, and exits with
// the status.
func shellcheckExiting(t *testing.T, status int, word string) string {
	t.Helper()

	name, head, say, exit := "shellcheck", "#!/bin/sh\n", "echo '"+word+"'\n", "exit %d\n"
	if runtime.GOOS == "windows" {
		name, head, say, exit = "shellcheck.bat", "", "@echo."+word+"\r\n", "@exit %d\r\n"
	}

	// Nothing at all, not even an empty line.
	if word == "" {
		say = ""
	}

	script := head + say + fmt.Sprintf(exit, status)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	return dir
}

// TestContractShellcheckThatDidNotScan: a shellcheck that exits with a
// status that is no scan's (3: an option it does not take), or with a scan's
// status and no report (nothing, or words that are none: what a wrapper
// failing in its place leaves), is an error and no lint: exit 1, the RUN's
// line on stderr, nothing on stdout. With no shellcheck on PATH the run
// lints without it and says so on stderr, at every log level;
// --without-shellcheck lints without it and says nothing.
func TestContractShellcheckThatDidNotScan(t *testing.T) {
	t.Parallel()

	// DL3006: an image with no tag.
	file := write(t, "Dockerfile", "FROM debian\nRUN echo $HOME\n")
	empty := t.TempDir()

	reported := func(failures []failure) bool {
		return slices.ContainsFunc(failures, func(f failure) bool { return f.Code == "DL3006" })
	}

	for name, stand := range map[string]struct {
		status int
		word   string
	}{
		"exiting 3":                 {3, ""},
		"exiting 1 with no report":  {1, ""},
		"exiting 0 with no report":  {0, ""},
		"exiting 1 with words":      {1, "broken"},
		"exiting 0 with words":      {0, "broken"},
		"exiting 3 with a report":   {3, "[]"},
		"exiting 1 with half of it": {1, "["},
	} {
		env := append(homeEnv(empty), "PATH="+shellcheckExiting(t, stand.status, stand.word))

		code, failures, stderr := runIn(t, empty, env, file)
		if code != 1 || failures != nil || !strings.Contains(stderr, "shellcheck did not scan the RUN at line 2") {
			t.Errorf("a shellcheck %s: exit %d, stdout %v, stderr %q; want 1, nothing, the RUN's line",
				name, code, failures, stderr)
		}
	}

	// The premise of the rows above: a stand-in that prints an empty report
	// is a shellcheck that scanned, and the other rules' failures are printed.
	scanning := append(homeEnv(empty), "PATH="+shellcheckExiting(t, 0, "[]"))
	if code, failures, stderr := runIn(t, empty, scanning, file); code != 1 || !reported(failures) {
		t.Fatalf("a shellcheck printing []: exit %d, stdout %v, stderr %q; want 1 and DL3006", code, failures, stderr)
	}

	refusing := append(homeEnv(empty), "PATH="+shellcheckExiting(t, 3, ""))

	// --no-fail is about violations: a scan that did not happen still fails.
	code, failures, _ := runIn(t, empty, refusing, "--no-fail", file)
	if code != 1 || failures != nil {
		t.Errorf("a shellcheck exiting 3, --no-fail: exit %d, stdout %v; want 1 and nothing", code, failures)
	}

	none := append(homeEnv(empty), "PATH="+t.TempDir())

	code, failures, stderr := runIn(t, empty, none, "--log-level", "error", file)
	if code != 1 || !reported(failures) || !strings.Contains(stderr, "no shellcheck in PATH") {
		t.Errorf("no shellcheck, --log-level error: exit %d, stdout %v, stderr %q; want 1, DL3006, the warning",
			code, failures, stderr)
	}

	code, failures, stderr = runIn(t, empty, none, "--without-shellcheck", file)
	if code != 1 || !reported(failures) || strings.Contains(stderr, "shellcheck") {
		t.Errorf("no shellcheck, --without-shellcheck: exit %d, stdout %v, stderr %q; want 1, DL3006, no warning",
			code, failures, stderr)
	}
}
