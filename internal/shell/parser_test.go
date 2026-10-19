package shell_test

import (
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
