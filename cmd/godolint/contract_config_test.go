package main_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Dockerfiles with one violation each, of each severity.
const (
	oneError   = "FROM debian:12\nWORKDIR app\n"                      // DL3000
	oneWarning = "FROM ubuntu\n"                                      // DL3006
	oneInfo    = "FROM debian:12\nRUN true\nRUN true\n"               // DL3059
	oneStyle   = "FROM debian:12\nLABEL invalid..character=\"foo\"\n" // DL3048
)

// TestContractExitStatusBySeverity: the exit status is 1 when a violation is
// at the failure threshold or graver, 0 otherwise, the violations on stdout
// either way; the threshold is info unless -t or the file says another;
// ignore and none are below style; --no-fail exits 0 whatever is found.
func TestContractExitStatusBySeverity(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"error":   write(t, "Dockerfile", oneError),
		"warning": write(t, "Dockerfile", oneWarning),
		"info":    write(t, "Dockerfile", oneInfo),
		"style":   write(t, "Dockerfile", oneStyle),
	}
	severities := []string{"error", "warning", "info", "style"}

	for _, tc := range []struct {
		threshold []string
		fails     []string
	}{
		{nil, []string{"error", "warning", "info"}},
		{[]string{"-t", "error"}, []string{"error"}},
		{[]string{"--failure-threshold", "warning"}, []string{"error", "warning"}},
		{[]string{"-t", "info"}, []string{"error", "warning", "info"}},
		{[]string{"-t", "style"}, severities},
		{[]string{"-t", "ignore"}, severities},
		{[]string{"-t", "none"}, severities},
		{[]string{"--no-fail"}, nil},
		{[]string{"--no-fail", "-t", "style"}, nil},
	} {
		for _, severity := range severities {
			args := append([]string{"--without-shellcheck"}, tc.threshold...)

			code, failures, stderr := run(t, append(args, files[severity])...)
			if len(failures) != 1 || failures[0].Level != severity {
				t.Fatalf("%v, one %s: %v\n%s", tc.threshold, severity, failures, stderr)
			}

			want := 0
			if slices.Contains(tc.fails, severity) {
				want = 1
			}

			if code != want {
				t.Errorf("%v, one %s: exit %d, want %d", tc.threshold, severity, code, want)
			}
		}
	}

	if code, _, stderr := run(t, "-t", "fatal", files["error"]); code != 1 || !strings.Contains(stderr, "fatal") {
		t.Errorf("-t fatal: exit %d, stderr %q; want 1 and the value named", code, stderr)
	}
}

// TestContractSettingsOnTheCommandLine: hadolint's flags do here what they
// do there: --error, --warning, --info and --style give a code a severity,
// the gravest when several do; --trusted-registry names the registries
// DL3026 allows; --require-label a label DL3049 asks for and its type, text
// when none is given; --strict-labels refuses any other label; --no-color
// is taken and changes nothing in the JSON.
func TestContractSettingsOnTheCommandLine(t *testing.T) {
	t.Parallel()

	untagged := write(t, "Dockerfile", oneWarning)

	for _, tc := range []struct {
		args  []string
		level string
	}{
		{[]string{"--error", "DL3006"}, "error"},
		{[]string{"--warning", "DL3006"}, "warning"},
		{[]string{"--info", "DL3006"}, "info"},
		{[]string{"--style", "DL3006"}, "style"},
		{[]string{"--style", "DL3006", "--error", "DL3006", "--info", "DL3006"}, "error"},
		{[]string{"--error", "DL3000", "--no-color"}, "warning"},
	} {
		_, failures, stderr := run(t, append([]string{"--without-shellcheck", untagged}, tc.args...)...)
		if len(failures) != 1 || failures[0].Level != tc.level {
			t.Errorf("%v: %v, want one %s\n%s", tc.args, failures, tc.level, stderr)
		}
	}

	registries := write(t, "Dockerfile", "FROM other.io/tools:1\nCOPY --from=mine.example.com/a:1 /a /b\n")

	for _, tc := range []struct {
		args  []string
		lines []int
	}{
		{nil, nil},
		{[]string{"--trusted-registry", "mine.example.com"}, []int{1}},
		{[]string{"--trusted-registry", "other.io"}, []int{2}},
		{[]string{"--trusted-registry", "other.io", "--trusted-registry", "*.example.com"}, nil},
	} {
		_, failures, _ := run(t, append([]string{"--without-shellcheck", registries}, tc.args...)...)

		var lines []int

		for _, f := range failures {
			if f.Code == "DL3026" {
				lines = append(lines, f.Line)
			}
		}

		if !slices.Equal(lines, tc.lines) {
			t.Errorf("%v: DL3026 at lines %v, want %v", tc.args, lines, tc.lines)
		}
	}

	labelled := write(t, "Dockerfile", "FROM debian:12\nLABEL maintainer=\"someone\" extra=\"x\"\n")

	for _, tc := range []struct {
		args  []string
		codes []string
	}{
		{nil, nil},
		{[]string{"--require-label", "maintainer"}, nil},
		{[]string{"--require-label", "maintainer:text"}, nil},
		{[]string{"--require-label", "maintainer:email"}, []string{"DL3058"}},
		{[]string{"--require-label", "version:semver"}, []string{"DL3049"}},
		{[]string{"--require-label", "maintainer", "--strict-labels"}, []string{"DL3050"}},
		{[]string{"--strict-labels"}, []string{"DL3050"}},
	} {
		_, failures, stderr := run(t, append([]string{"--without-shellcheck", "-t", "style", labelled}, tc.args...)...)
		if got := codesIn(failures); !slices.Equal(got, tc.codes) {
			t.Errorf("%v: %v, want %v\n%s", tc.args, got, tc.codes, stderr)
		}
	}

	code, failures, stderr := run(t, "--require-label", "maintainer:markdown", labelled)
	if code != 1 || failures != nil || !strings.Contains(stderr, "maintainer:markdown") {
		t.Errorf(
			"an unknown label type: exit %d, %v, stderr %q; want 1, no output and the value named",
			code,
			failures,
			stderr,
		)
	}
}

// codesIn is the codes of failures, sorted, each as often as it is reported.
func codesIn(failures []failure) []string {
	var codes []string

	for _, f := range failures {
		codes = append(codes, f.Code)
	}

	slices.Sort(codes)

	return codes
}

// logged reports whether the error log on stderr names path. The log is JSON
// when stderr is not a terminal, so the backslashes of a Windows path are
// escaped in it.
func logged(stderr, path string) bool {
	return strings.Contains(stderr, path) || strings.Contains(stderr, strings.ReplaceAll(path, `\`, `\\`))
}

// config writes a configuration file and returns its path.
func config(t *testing.T, content string) string {
	t.Helper()

	return write(t, "hadolint.yaml", content)
}

// TestContractConfigurationFile: --config reads hadolint's own file, every
// key doing what its flag does; a file of comments says nothing; verbose
// says on stderr which file is used, in hadolint's words.
func TestContractConfigurationFile(t *testing.T) {
	t.Parallel()

	dockerfile := write(t, "Dockerfile",
		"FROM other.io/tools\n# hadolint ignore=DL4000\nMAINTAINER x\nLABEL extra=\"x\"\n")

	plain := []string{"DL3006"}

	for _, tc := range []struct {
		name  string
		file  string
		codes []string
		exit  int
	}{
		{"nothing but comments", "# nothing\n\n", plain, 1},
		{"ignored", "ignored:\n  - DL3006\n", nil, 0},
		{"override", "override:\n  style:\n    - DL3006\n", plain, 0},
		{"override, the gravest", "override:\n  style: [DL3006]\n  error: [DL3006]\n  info: [DL3006]\n", plain, 1},
		{"trustedRegistries", "trustedRegistries:\n  - docker.io\n", []string{"DL3006", "DL3026"}, 1},
		{"label-schema", "label-schema:\n  maintainer: email\n", []string{"DL3006", "DL3049"}, 1},
		{"label-schema with the empty type", "label-schema:\n  extra: \"\"\n", plain, 1},
		{"strict-labels", "strict-labels: true\n", []string{"DL3006", "DL3050"}, 1},
		{"strict-labels off", "strict-labels: false\n", plain, 1},
		{"disable-ignore-pragma", "disable-ignore-pragma: true\n", []string{"DL3006", "DL4000"}, 1},
		{"failure-threshold", "failure-threshold: error\n", plain, 0},
		{"no-fail", "no-fail: true\n", plain, 0},
		{"no-color", "no-color: true\n", plain, 1},
		{"format json", "format: json\n", plain, 1},
		{"formats json", "formats:\n  - json\n", plain, 1},
		{
			"every key at once",
			"no-fail: false\nno-color: true\nverbose: false\nformat: json\nformats: [json]\n" +
				"override:\n  error: [DL3006]\n  warning: []\n  info: []\n  style: []\n" +
				"ignored: [DL3050]\ntrustedRegistries: [other.io]\nlabel-schema:\n  extra: text\n" +
				"strict-labels: true\ndisable-ignore-pragma: false\nfailure-threshold: style\n",
			plain, 1,
		},
	} {
		code, failures, stderr := run(t, "--without-shellcheck", "--config", config(t, tc.file), dockerfile)
		if got := codesIn(failures); code != tc.exit || !slices.Equal(got, tc.codes) {
			t.Errorf("%s: exit %d, %v; want %d, %v\n%s", tc.name, code, got, tc.exit, tc.codes, stderr)
		}
	}

	file := config(t, "verbose: true\n")

	_, _, stderr := run(t, "--without-shellcheck", "-c", file, dockerfile)
	if !strings.Contains(stderr, "Configuration file used: "+file) {
		t.Errorf("verbose in the file: stderr %q; want the file named", stderr)
	}

	_, _, stderr = run(t, "--without-shellcheck", "--verbose", dockerfile)
	if !strings.Contains(stderr, "No configuration was specified. Using default configuration") {
		t.Errorf("--verbose with no file: stderr %q", stderr)
	}

	_, _, stderr = run(t, "--without-shellcheck", "-c", config(t, "ignored: []\n"), dockerfile)
	if strings.Contains(stderr, "onfiguration") {
		t.Errorf("not verbose: stderr %q; want nothing of the configuration", stderr)
	}
}

// TestContractConfigurationFileRefused: a file godolint cannot obey whole is
// obeyed in nothing: exit 1, nothing on stdout, and stderr naming the file
// and what is wrong in it: a key hadolint does not have, at the top or under
// override; a format other than json; a label type or a threshold that is
// not one; a value of the wrong shape; a file that is not there.
func TestContractConfigurationFileRefused(t *testing.T) {
	t.Parallel()

	dockerfile := write(t, "Dockerfile", "FROM debian:12\n")

	for _, tc := range []struct {
		name  string
		file  string
		named string
	}{
		{"an unknown key", "ignored: [DL3006]\nfailureThreshold: error\n", "`failureThreshold`"},
		{"an unknown key under override", "override:\n  fatal: [DL3006]\n", "`override.fatal`"},
		{"ignore is not an override", "override:\n  ignore: [DL3006]\n", "`override.ignore`"},
		{"format tty", "format: tty\n", "`tty`"},
		{"formats with sarif", "formats: [json, sarif]\n", "`sarif`"},
		{"an unknown label type", "label-schema:\n  maintainer: markdown\n", "`markdown`"},
		{"a label with no type", "label-schema:\n  maintainer:\n", "`maintainer` no type"},
		{"an unknown threshold", "failure-threshold: fatal\n", "`fatal`"},
		{"a list where a switch goes", "ignored: []\nno-fail: [yes]\n", "line 2"},
		{"a word where a list goes", "ignored: DL3006\n", "DL3006"},
		{"a list at the top", "- ignored\n", "hadolint.yaml"},
		{"not YAML", "ignored: [DL3006\n", "hadolint.yaml"},
	} {
		file := config(t, tc.file)

		code, failures, stderr := run(t, "--without-shellcheck", "--config", file, dockerfile)
		if code != 1 || failures != nil || !strings.Contains(stderr, tc.named) || !logged(stderr, file) {
			t.Errorf("%s: exit %d, %v, stderr %q; want 1, no output, the file and %s named",
				tc.name, code, failures, stderr, tc.named)
		}
	}

	missing := filepath.Join(t.TempDir(), "absent.yaml")
	if code, failures, stderr := run(t, "--config", missing, dockerfile); code != 1 || failures != nil ||
		!logged(stderr, missing) {
		t.Errorf("--config on a file that is not there: exit %d, %v, stderr %q", code, failures, stderr)
	}
}

// TestContractCommandLineOverFile: a list is the file's and the command
// line's together; a switch is on when either turns it on; the threshold is
// the command line's when it gives one.
func TestContractCommandLineOverFile(t *testing.T) {
	t.Parallel()

	dockerfile := write(t, "Dockerfile", "FROM ubuntu\nMAINTAINER x\nLABEL extra=\"x\"\n")
	file := config(t, "ignored: [DL3006]\noverride:\n  style: [DL4000]\nfailure-threshold: error\n"+
		"label-schema:\n  maintainer: text\ntrustedRegistries: [docker.io]\n")

	for _, tc := range []struct {
		name  string
		args  []string
		codes []string
		exit  int
	}{
		{"the file alone", nil, []string{"DL3049", "DL4000"}, 0},
		{"another code ignored", []string{"--ignore", "DL4000"}, []string{"DL3049"}, 0},
		{"a graver severity for the same code", []string{"--error", "DL4000"}, []string{"DL3049", "DL4000"}, 1},
		{"a lighter severity for the same code", []string{"--info", "DL4000"}, []string{"DL3049", "DL4000"}, 0},
		{"the threshold", []string{"-t", "style"}, []string{"DL3049", "DL4000"}, 1},
		{"a switch", []string{"--strict-labels"}, []string{"DL3049", "DL3050", "DL4000"}, 0},
		{"another label", []string{"--require-label", "extra:semver"}, []string{"DL3049", "DL3056", "DL4000"}, 0},
		{"another registry", []string{"--trusted-registry", "other.io"}, []string{"DL3049", "DL4000"}, 0},
	} {
		args := append([]string{"--without-shellcheck", "--config", file, dockerfile}, tc.args...)

		code, failures, stderr := run(t, args...)
		if got := codesIn(failures); code != tc.exit || !slices.Equal(got, tc.codes) {
			t.Errorf("%s: exit %d, %v; want %d, %v\n%s", tc.name, code, got, tc.exit, tc.codes, stderr)
		}
	}
}

// TestContractLabelInFileAndFlag: a label the file and the command line both
// name keeps the file's type, as in hadolint; one only the command line
// names has the command line's.
func TestContractLabelInFileAndFlag(t *testing.T) {
	t.Parallel()

	// "x" is no URL (DL3052) and no email (DL3058).
	dockerfile := write(t, "Dockerfile", "FROM debian:12\nLABEL extra=\"x\" other=\"x\"\n")
	file := config(t, "label-schema:\n  extra: url\n")

	for _, tc := range []struct {
		name  string
		args  []string
		codes []string
	}{
		{"the file alone", nil, []string{"DL3052"}},
		{"the same label as an email", []string{"--require-label", "extra:email"}, []string{"DL3052"}},
		{"another label as an email", []string{"--require-label", "other:email"}, []string{"DL3052", "DL3058"}},
	} {
		args := append([]string{"--without-shellcheck", "--config", file, dockerfile}, tc.args...)

		_, failures, stderr := run(t, args...)
		if got := codesIn(failures); !slices.Equal(got, tc.codes) {
			t.Errorf("%s: %v, want %v\n%s", tc.name, got, tc.codes, stderr)
		}
	}
}

// TestContractConfigurationFileSearch: with no --config, the file used is
// the first that exists of hadolint's places, in hadolint's order: the
// working directory, the user's configuration directory, hadolint's own
// directory, the home; --config goes before them all; a directory of a
// file's name is not a file; with none, there is no configuration.
func TestContractConfigurationFileSearch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cwd, home := filepath.Join(root, "cwd"), filepath.Join(root, "home")

	configDir, ownDir, documents := filepath.Join(home, ".config"), filepath.Join(home, ".hadolint"), home
	if runtime.GOOS == "windows" {
		configDir = filepath.Join(home, "AppData")
		ownDir = filepath.Join(configDir, "hadolint")
		documents = filepath.Join(home, "Documents")
	}

	places := []string{
		filepath.Join(cwd, ".hadolint.yaml"),
		filepath.Join(cwd, ".hadolint.yml"),
		filepath.Join(configDir, "hadolint.yaml"),
		filepath.Join(configDir, "hadolint.yml"),
		filepath.Join(ownDir, "hadolint.yaml"),
		filepath.Join(ownDir, "hadolint.yml"),
		filepath.Join(ownDir, "config.yaml"),
		filepath.Join(ownDir, "config.yml"),
		filepath.Join(documents, ".hadolint.yaml"),
		filepath.Join(documents, ".hadolint.yml"),
	}

	for _, place := range places {
		if err := os.MkdirAll(filepath.Dir(place), 0o700); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(place, []byte("ignored: []\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	dockerfile := write(t, "Dockerfile", "FROM debian:12\n")

	used := func(args ...string) string {
		_, _, stderr := runIn(
			t,
			cwd,
			homeEnv(home),
			append([]string{"--verbose", "--without-shellcheck", dockerfile}, args...)...,
		)

		for line := range strings.SplitSeq(stderr, "\n") {
			if file, ok := strings.CutPrefix(line, "Configuration file used: "); ok {
				return file
			}
		}

		return strings.TrimSpace(stderr)
	}

	// The working directory is told by the system, which may name it through
	// another path than the one it was made under.
	same := func(a, b string) bool {
		infoA, errA := os.Stat(a)
		infoB, errB := os.Stat(b)

		return errA == nil && errB == nil && os.SameFile(infoA, infoB)
	}

	explicit := config(t, "ignored: []\n")
	if got := used("--config", explicit); got != explicit {
		t.Errorf("--config with a file in every place: %q used, want %q", got, explicit)
	}

	for i, place := range places {
		if got := used(); !same(got, place) {
			t.Fatalf("with the %d places before it gone: %q used, want %q", i, got, place)
		}

		if err := os.Remove(place); err != nil {
			t.Fatal(err)
		}

		// What is left in its place is a directory, which is no file.
		if err := os.Mkdir(place, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	if got := used(); got != "No configuration was specified. Using default configuration" {
		t.Errorf("no file in any place: %q", got)
	}
}

// TestContractConfigurationDirectory: the user's configuration directory is
// the one XDG_CONFIG_HOME names when it names an absolute one, and .config
// in the home otherwise. Windows has %APPDATA% and no such variable.
func TestContractConfigurationDirectory(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("no XDG_CONFIG_HOME on Windows")
	}

	root := t.TempDir()
	cwd, home, xdg := filepath.Join(root, "cwd"), filepath.Join(root, "home"), filepath.Join(root, "xdg")

	inHome := filepath.Join(home, ".config", "hadolint.yaml")
	inXDG := filepath.Join(xdg, "hadolint.yaml")
	inRelative := filepath.Join(cwd, "relative", "hadolint.yaml")

	for _, file := range []string{inHome, inXDG, inRelative} {
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file, []byte("ignored: []\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	dockerfile := write(t, "Dockerfile", "FROM debian:12\n")

	for value, want := range map[string]string{
		"":         inHome,
		xdg:        inXDG,
		"relative": inHome,
	} {
		env := append(homeEnv(home), "XDG_CONFIG_HOME="+value)

		_, _, stderr := runIn(t, cwd, env, "--verbose", "--without-shellcheck", dockerfile)
		if !strings.Contains(stderr, "Configuration file used: "+want) {
			t.Errorf("XDG_CONFIG_HOME=%q: stderr %q; want %q used", value, stderr, want)
		}
	}
}

// TestContractOverrideEachSeverity: each of the four severities is given by
// its flag and by its list in the file, to a rule that has another.
func TestContractOverrideEachSeverity(t *testing.T) {
	t.Parallel()

	// DL3000 is an error, DL3006 a warning.
	dockerfile := write(t, "Dockerfile", "FROM ubuntu\nWORKDIR app\n")

	for _, tc := range []struct {
		severity string
		code     string
	}{
		{"error", "DL3006"},
		{"warning", "DL3000"},
		{"info", "DL3006"},
		{"style", "DL3006"},
	} {
		for name, args := range map[string][]string{
			"the flag": {"--" + tc.severity, tc.code},
			"the file": {"--config", config(t, "override:\n  "+tc.severity+":\n    - "+tc.code+"\n")},
		} {
			_, failures, stderr := run(t, append([]string{"--without-shellcheck", dockerfile}, args...)...)

			var level string

			for _, f := range failures {
				if f.Code == tc.code {
					level = f.Level
				}
			}

			if level != tc.severity || len(failures) != 2 {
				t.Errorf(
					"%s, %s for %s: reported as %q among %v\n%s",
					name,
					tc.severity,
					tc.code,
					level,
					failures,
					stderr,
				)
			}
		}
	}
}
