package shell_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/shell"
)

func TestParseShell(t *testing.T) {
	t.Parallel()

	t.Run("simple command", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("apt-get update")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		if len(ps.PresentCommands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(ps.PresentCommands))
		}

		cmd := ps.PresentCommands[0]
		if cmd.Name != "apt-get" {
			t.Errorf("expected command name 'apt-get', got '%s'", cmd.Name)
		}

		args := shell.GetArgs(cmd)
		if len(args) != 1 || args[0] != "update" {
			t.Errorf("expected args ['update'], got %v", args)
		}
	})

	t.Run("command with flags", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("apt-get install -y vim")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		cmd := ps.PresentCommands[0]
		if cmd.Name != "apt-get" {
			t.Errorf("expected command name 'apt-get', got '%s'", cmd.Name)
		}

		if !shell.HasFlag("y", cmd) {
			t.Error("expected command to have -y flag")
		}

		argsNoFlags := shell.GetArgsNoFlags(cmd)
		expectedArgs := []string{"install", "vim"}

		if len(argsNoFlags) != len(expectedArgs) {
			t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(argsNoFlags), argsNoFlags)
		}
	})

	t.Run("chained commands with &&", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("apt-get update && apt-get install -y vim")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		if len(ps.PresentCommands) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(ps.PresentCommands))
		}

		names := shell.FindCommandNames(ps)
		if len(names) != 2 || names[0] != "apt-get" || names[1] != "apt-get" {
			t.Errorf("expected ['apt-get', 'apt-get'], got %v", names)
		}
	})

	t.Run("chained commands with semicolon", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("echo foo; echo bar")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		if len(ps.PresentCommands) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(ps.PresentCommands))
		}
	})

	t.Run("long flag with equals", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("useradd --uid=1000 user")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		cmd := ps.PresentCommands[0]
		if !shell.HasFlag("uid", cmd) {
			t.Error("expected command to have --uid flag")
		}
	})

	t.Run("multiple short flags", func(t *testing.T) {
		t.Parallel()

		ps, err := shell.ParseShell("tar -xzf archive.tar.gz")
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		cmd := ps.PresentCommands[0]
		if !shell.HasFlag("x", cmd) || !shell.HasFlag("z", cmd) || !shell.HasFlag("f", cmd) {
			t.Error("expected command to have -x, -z, and -f flags")
		}
	})
}

func TestCmdHasArgs(t *testing.T) {
	t.Parallel()

	ps, err := shell.ParseShell("apt-get install vim")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	cmd := ps.PresentCommands[0]

	if !shell.CmdHasArgs("apt-get", []string{"install"}, cmd) {
		t.Error("expected CmdHasArgs to return true for apt-get with install")
	}

	if shell.CmdHasArgs("apt-get", []string{"update"}, cmd) {
		t.Error("expected CmdHasArgs to return false for apt-get with update")
	}

	if shell.CmdHasArgs("yum", []string{"install"}, cmd) {
		t.Error("expected CmdHasArgs to return false for different command")
	}
}

// TestCmdHasAllArgs: every expected argument must be present, where CmdHasArgs
// settles for any of them.
func TestCmdHasAllArgs(t *testing.T) {
	t.Parallel()

	ps, err := shell.ParseShell("yum clean")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	cmd := ps.PresentCommands[0]

	if !shell.CmdHasArgs("yum", []string{"clean", "all"}, cmd) {
		t.Error("CmdHasArgs: yum clean has one of clean, all")
	}

	if shell.CmdHasAllArgs("yum", []string{"clean", "all"}, cmd) {
		t.Error("CmdHasAllArgs: yum clean lacks all")
	}

	if !shell.CmdHasAllArgs("yum", []string{"clean"}, cmd) {
		t.Error("CmdHasAllArgs: yum clean has clean")
	}

	if shell.CmdHasAllArgs("dnf", []string{"clean"}, cmd) {
		t.Error("CmdHasAllArgs: a different command")
	}
}

// TestIsRecursiveRemoval: an rm of a listed path counts under any spelling of
// the recursive flag, and neither without one nor for another path or command.
func TestIsRecursiveRemoval(t *testing.T) {
	t.Parallel()

	paths := []string{"/var/cache/yum", "/var/cache/yum/*"}

	for script, want := range map[string]bool{
		"rm -rf /var/cache/yum/*":               true,
		"rm -fr /var/cache/yum":                 true,
		"rm -r -f /var/cache/yum":               true,
		"rm -Rf /var/cache/yum":                 true,
		"rm --recursive --force /var/cache/yum": true,
		"rm -rf /tmp/x /var/cache/yum":          true,
		"rm -f /var/cache/yum/*":                false,
		"rm /var/cache/yum":                     false,
		"rm -rf /var/cache/dnf":                 false,
		"rm -rf /var/cache/yum/repos":           false,
		"rmdir -r /var/cache/yum":               false,
	} {
		ps, err := shell.ParseShell(script)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", script, err)
		}

		if got := shell.IsRecursiveRemoval(paths, ps.PresentCommands[0]); got != want {
			t.Errorf("IsRecursiveRemoval(%q) = %t; want %t", script, got, want)
		}
	}
}

// TestBasename: a shell is known by the last path segment of its name.
func TestBasename(t *testing.T) {
	t.Parallel()

	for name, want := range map[string]string{
		"/usr/bin/bash": "bash",
		"/bin/zsh":      "zsh",
		"ash":           "ash",
		"pwsh":          "pwsh",
	} {
		if got := shell.Basename(name); got != want {
			t.Errorf("Basename(%q) = %q; want %q", name, got, want)
		}
	}
}

// TestDropFlagArg: the argument after a listed flag is its value and goes; a
// flag carrying its value inline (--flag=value) keeps the argument after it;
// the value is the argument with the smallest id above the flag's, so an
// unlisted flag between them is not mistaken for it; other flags stay.
func TestDropFlagArg(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		script string
		drop   []string
		want   []string
	}{
		{"dnf install -y --releasever 9 tomcat", []string{"releasever"}, []string{"install", "tomcat"}},
		{"dnf install -y --releasever=9 tomcat", []string{"releasever"}, []string{"install", "tomcat"}},
		{"dnf install -y --setopt a=b tomcat", []string{"setopt"}, []string{"install", "tomcat"}},
		{"pecl install -D with-foo=yes redis", []string{"D", "define"}, []string{"install", "redis"}},
		{"dnf install -y --enablerepo crb -x foo tomcat", []string{"enablerepo", "x"}, []string{"install", "tomcat"}},
		{"dnf install -y --releasever 9 tomcat", []string{"other"}, []string{"install", "9", "tomcat"}},
	} {
		ps, err := shell.ParseShell(tc.script)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", tc.script, err)
		}

		cmd := shell.DropFlagArg(tc.drop, ps.PresentCommands[0])

		got := shell.GetArgsNoFlags(cmd)
		if len(got) != len(tc.want) {
			t.Errorf("DropFlagArg(%v) on %q: %v; want %v", tc.drop, tc.script, got, tc.want)

			continue
		}

		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("DropFlagArg(%v) on %q: %v; want %v", tc.drop, tc.script, got, tc.want)

				break
			}
		}

		if len(cmd.Flags) != len(ps.PresentCommands[0].Flags) {
			t.Errorf("DropFlagArg(%v) on %q dropped flags", tc.drop, tc.script)
		}
	}
}

func TestCountCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		script   string
		expected int
	}{
		{"single command", "echo hello", 1},
		{"two commands with &&", "echo foo && echo bar", 2},
		{"three commands with semicolon", "a; b; c", 3},
		{"pipeline", "cat file | grep foo", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, err := shell.ParseShell(tt.script)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			count := shell.CountCommands(ps)
			if count != tt.expected {
				t.Errorf("expected %d commands, got %d", tt.expected, count)
			}
		})
	}
}

// TestParseShellReadsAWordAsHadolintDoes: the text of a command's name and
// of each argument, which the rules look into: quotes gone, escapes read,
// an expansion as its name and what it holds, a brace expansion nothing.
// Each row is hadolint's simplify (Shell.hs) over ShellCheck's parse, but
// the braces around two dots that make no sequence, which stay as the
// shell leaves them.
func TestParseShellReadsAWordAsHadolintDoes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, script string
		words        []string
	}{
		{"plain", `cmd one two`, []string{"cmd", "one", "two"}},
		{"quotes", `cmd "a b" 'c d' e"f"'g'`, []string{"cmd", "a b", "c d", "efg"}},
		{"a backslash outside quotes", `\cmd a\ b pkg-1.2\* /var/cache/yum/\* \\x`, []string{"cmd", "a b", "pkg-1.2*", "/var/cache/yum/*", `\x`}},
		{"a backslash in double quotes", `cmd "a\"b" "c\$d" "e\\f" "g\nh" "i\'j"`, []string{"cmd", `a"b`, "c$d", `e\f`, `g\nh`, `i\'j`}},
		{"a backslash in single quotes", `cmd 'a\b' 'c\'`, []string{"cmd", `a\b`, `c\`}},
		{"a continued line", "cmd one\\\ntwo", []string{"cmd", "onetwo"}},
		{"a variable", `cmd $A ${B} "$C" x${D}y`, []string{"cmd", "${A}", "${B}", "${C}", "x${D}y"}},
		{
			"a variable with an operator",
			`cmd ${A:-b} "pkg${V:+==$V}" ${A:=c d} ${A:?} ${A-"x y"}`,
			[]string{"cmd", "${A:-b}", "pkg${V:+==${V}}", "${A:=c d}", "${A:?}", "${A-x y}"},
		},
		{"other forms of a variable", `cmd ${#A} ${A[0]} ${A/x/y} ${A:1:2}`, []string{"cmd", "${#A}", "${A[0]}", "${A/x/y}", "${A:1:2}"}},
		{
			"a command substitution",
			`cmd $(cat /tmp/list) "$(id -u)" v$(echo ${V})`,
			[]string{"cmd", "${cat/tmp/list}", "${id-u}", "v${echo${V}}"},
		},
		{"a substitution of more than a command", `cmd $(a | b) $(a && b) $(a; b c)`, []string{"cmd", "${}", "${}", "${abc}"}},
		{"backquotes and arithmetic", "cmd `id -u` $((1 + 2))", []string{"cmd", "${VAR}", "${VAR}"}},
		{"a brace expansion", `cmd pkg-{a,b}-1.0 /lists/{a,b} {1..3} x{a,{b,c}}y`, []string{"cmd", "pkg--1.0", "/lists/", "", "xy"}},
		{"braces that expand nothing", `cmd {a} {} a{b}c ${A}{B}`, []string{"cmd", "{a}", "{}", "a{b}c", "${A}{B}"}},
		{"a glob", `cmd /a/* b? [c]d`, []string{"cmd", "/a/*", "b?", "[c]d"}},
		{"a process substitution", `cmd <(a b) x`, []string{"cmd", "", "x"}},
		{"a string to translate", `cmd $"a b" x`, []string{"cmd", "", "x"}},
		{"the old arithmetic", `cmd $[1+2] x`, []string{"cmd", "", "x"}},
		{"a substitution negated or in the background", `cmd $(! a b) $(a b &) $(a b)`, []string{"cmd", "${}", "${}", "${ab}"}},
		{
			"two dots that make no sequence",
			`cmd {foo..bar} x{a..}y {..} {a.b} {a..3} {$A..$B} {1..$N} {"a..b"} {1..3} {a..c}`,
			[]string{"cmd", "{foo..bar}", "x{a..}y", "{..}", "{a.b}", "{a..3}", "{${A}..${B}}", "{1..${N}}", "{a..b}", "", ""},
		},
		{
			"a brace a backslash escapes",
			`cmd \{1..3\} \{1..3} {1..3\} \{a..b\} \{a..b} \{a,b\} \{a,b}`,
			[]string{"cmd", "{1..3}", "{1..3}", "{1..3}", "{a..b}", "{a..b}", "{a,b}", "{a,b}"},
		},
		{
			"a backslash in the word of a variable",
			"cmd ${A:-a\\b} ${A:-a\\}b} ${A:-a\\$b} \"${A:-a\\\"b}\"",
			[]string{"cmd", `${A:-a\b}`, "${A:-a}b}", "${A:-a$b}", `${A:-a"b}`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := shell.ParseShell(tc.script)
			if err != nil || len(parsed.PresentCommands) == 0 {
				t.Fatalf("ParseShell(%q) = %v, %v", tc.script, parsed, err)
			}

			first := parsed.PresentCommands[0]

			got := []string{first.Name}
			for _, arg := range first.Arguments {
				got = append(got, arg.Arg)
			}

			if !slices.Equal(got, tc.words) {
				t.Errorf("%q reads as\n%q\nwant\n%q", tc.script, got, tc.words)
			}
		})
	}
}
