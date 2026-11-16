package shell_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/shell"
)

// TestParseShellNamesTheCommands: what is a command of a script, by what
// name, with which arguments. A declaration is a command; a wrapper is the
// command it runs, unless an option follows it.
func TestParseShellNamesTheCommands(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		script   string
		commands []string
	}{
		{"an export", `export A=1 B`, []string{"export A=1 B"}},
		{"a declare with an option", `declare -x CC=clang`, []string{"declare -x CC=clang"}},
		{"the other declarations", `readonly A=1; local B; typeset C=2`, []string{"readonly A=1", "local B", "typeset C=2"}},
		{"a declaration chained to a command", `export A=1 && make all`, []string{"export A=1", "make all"}},
		{"an assignment alone", `A=1`, nil},
		{"busybox", `busybox wget -q https://example.com/a`, []string{"wget -q https://example.com/a"}},
		{"command", `command pip install requests`, []string{"pip install requests"}},
		{"builtin", `builtin cd /tmp`, []string{"cd /tmp"}},
		{"exec", `exec gosu app server`, []string{"gosu app server"}},
		{"run", `run make all`, []string{"make all"}},
		{"a wrapper before an option", `command -v pip; busybox --list`, []string{"command -v pip", "busybox --list"}},
		{"a wrapper alone", `busybox`, []string{"busybox"}},
		{"a wrapper of a wrapper", `command busybox wget`, []string{"busybox wget"}},
		{"a command timed", `time make all`, []string{"make all"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := shell.ParseShell(tc.script)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			var commands []string
			for _, cmd := range parsed.PresentCommands {
				commands = append(commands, strings.Join(append([]string{cmd.Name}, shell.GetArgs(cmd)...), " "))
			}

			if !slices.Equal(commands, tc.commands) {
				t.Errorf("commands %q, want %q", commands, tc.commands)
			}
		})
	}
}
