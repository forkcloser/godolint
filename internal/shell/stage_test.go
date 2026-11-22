package shell_test

import (
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestShellcheckRule_Stages is hadolint's ShellcheckSpec: the shell and the
// variables shellcheck runs with, per stage, through the Dockerfile parser.
// Each case asserts whether any SC code is reported.
func TestShellcheckRule_Stages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dockerfile string
		fails      bool
	}{
		{
			name:       "runs shellcheck on RUN instructions",
			dockerfile: `RUN echo $MISSING_QUOTES`,
			fails:      true,
		},
		{
			name:       "not warns on valid scripts",
			dockerfile: `RUN echo foo`,
			fails:      false,
		},
		{
			name:       "no warnings on exec notations",
			dockerfile: `RUN ["foobar", "$f"]`,
			fails:      false,
		},
		{
			name: "Does not complain on default env vars",
			dockerfile: `RUN echo "$HTTP_PROXY"
RUN echo "$http_proxy"
RUN echo "$HTTPS_PROXY"
RUN echo "$https_proxy"
RUN echo "$FTP_PROXY"
RUN echo "$ftp_proxy"
RUN echo "$NO_PROXY"
RUN echo "$no_proxy"`,
			fails: false,
		},
		{
			name:       "Complain on missing env vars",
			dockerfile: `RUN echo "$RTTP_PROXY"`,
			fails:      true,
		},
		{
			name: "Is aware of ARGS and ENV",
			dockerfile: `ARG foo=bar
ARG another_foo
ENV bar=10 baz=20
RUN echo "$foo"
RUN echo "$another_foo"
RUN echo "$bar"
RUN echo "$baz"`,
			fails: false,
		},
		{
			name: "Resets env vars after a FROM",
			dockerfile: `ARG foo=bar
ARG another_foo
ENV bar=10 baz=20
FROM debian
RUN echo "$foo"`,
			fails: true,
		},
		{
			name:       "Defaults the shell to sh",
			dockerfile: `RUN echo $RANDOM`,
			fails:      true,
		},
		{
			name: "Can change the shell check to bash",
			dockerfile: `SHELL ["/bin/bash", "-eo", "pipefail", "-c"]
RUN echo $RANDOM`,
			fails: false,
		},
		{
			name: "Resets the SHELL to sh after a FROM",
			dockerfile: `SHELL ["/bin/bash", "-eo", "pipefail", "-c"]
FROM debian
RUN echo $RANDOM`,
			fails: true,
		},
		{
			name: "Does not complain on ash shell",
			dockerfile: `SHELL ["/bin/ash", "-o", "pipefail", "-c"]
RUN echo hello`,
			fails: false,
		},
		{
			name: "Does not complain on non-posix shells: pwsh",
			dockerfile: `SHELL ["pwsh", "-c"]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`,
			fails: false,
		},
		{
			name: "Does not complain on non-posix shells: cmd.exe",
			dockerfile: `SHELL ["cmd.exe", "/c"]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`,
			fails: false,
		},
		{
			name: "Does not complain on non-posix shells, powershell - absolute path",
			dockerfile: `SHELL ["C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe", "-noprofile", "-noninteractive", "-command"]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`,
			fails: false,
		},
		{
			name: "Respects shell pragma",
			dockerfile: `FROM mcr.microsoft.com/foo/bar/windows:10
# hadolint shell = powershell.exe
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`,
			fails: false,
		},
		{
			name: "Respects global shell pragma",
			dockerfile: `# hadolint shell=powershell.exe
FROM mcr.microsoft.com/foo/bar/windows:10
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value
FROM mcr.microsoft.com/foo/bar/windows:10
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`,
			fails: false,
		},
		{
			name: "A shell pragma after a FROM is that stage's only",
			dockerfile: `FROM mcr.microsoft.com/foo/bar/windows:10
# hadolint shell=powershell.exe
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value
FROM debian
RUN echo $RANDOM`,
			fails: true,
		},
		{
			name: "SHELL instructions are propagated in multi-stage builds 1",
			dockerfile: `FROM ubuntu:24.04 AS base
SHELL ["/bin/bash", "-e", "-o", "pipefail", "-c"]

FROM base
RUN source /etc/os-release`,
			fails: false,
		},
		{
			name: "SHELL instructions are propagated only propagated if referencing earlier stage",
			dockerfile: `FROM ubuntu:24.04 AS base
SHELL ["/bin/bash", "-e", "-o", "pipefail", "-c"]

FROM ubuntu:24.04
RUN source /etc/os-release`,
			fails: true,
		},
		{
			name: "SHELL instructions are propagated in multi-stage builds 2",
			dockerfile: `FROM something AS nothing
FROM ubuntu:24.04 AS base
SHELL ["/bin/bash", "-e", "-o", "pipefail", "-c"]
FROM nothing
FROM base
RUN source /etc/os-release`,
			fails: false,
		},
		{
			name: "SHELL instructions are propagated in multi-stage builds 3",
			dockerfile: `FROM ubuntu:24.04 AS base
SHELL ["/bin/bash", "-e", "-o", "pipefail", "-c"]
FROM base AS next
RUN foobar && barfoo | tee logfile
FROM next
RUN source /etc/os-release`,
			fails: false,
		},
		{
			name: "ENV and ARG follow the stage a FROM names",
			dockerfile: `FROM debian AS base
ENV bar=10
ARG foo
FROM base
RUN echo "$foo$bar"`,
			fails: false,
		},
		{
			name: "The RUN an ONBUILD wraps is checked",
			dockerfile: `FROM debian
ONBUILD RUN echo $unquoted`,
			fails: true,
		},
		{
			name: "A shell named without its path is clean on a clean RUN",
			dockerfile: `SHELL ["bash", "-o", "pipefail", "-c"]
RUN echo ok`,
			fails: false,
		},
		{
			name: "A shell named without its path is that shell",
			dockerfile: `SHELL ["bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell named without its path is still checked",
			dockerfile: `SHELL ["bash", "-c"]
RUN echo $MISSING_QUOTES`,
			fails: true,
		},
		{
			name: "A shell run through env is clean on a clean RUN",
			dockerfile: `SHELL ["/usr/bin/env", "bash", "-c"]
RUN echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env is that shell",
			dockerfile: `SHELL ["env", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after its options and variables",
			dockerfile: `SHELL ["/usr/bin/env", "-i", "PATH=/bin", "/bin/bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env is still checked",
			dockerfile: `SHELL ["/usr/bin/env", "sh", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: true,
		},
		{
			name: "A shell run through env is the word after env's options",
			dockerfile: `SHELL ["/usr/bin/env", "-i", "sh", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: true,
		},
		{
			name: "busybox sh has local",
			dockerfile: `SHELL ["busybox", "sh", "-c"]
RUN f() { local x=1; echo "$x"; }; f`,
			fails: false,
		},
		{
			name: "sh through env is sh, which has no local",
			dockerfile: `SHELL ["env", "sh", "-c"]
RUN f() { local x=1; echo "$x"; }; f`,
			fails: true,
		},
		{
			name: "A shell run through env, after an option and its value",
			dockerfile: `SHELL ["/usr/bin/env", "-u", "BASH_ENV", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after a long option and its value",
			dockerfile: `SHELL ["env", "--chdir", "/tmp", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after the name it is given",
			dockerfile: `SHELL ["env", "-a", "sh", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after the name it is given, in the long form",
			dockerfile: `SHELL ["env", "--argv0", "sh", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after options in one word, the last with its value next",
			dockerfile: `SHELL ["env", "-iu", "BASH_ENV", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env, after options in one word, the last with its value in it",
			dockerfile: `SHELL ["env", "-iuBASH_ENV", "bash", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A shell run through env from a split string",
			dockerfile: `SHELL ["env", "-S", "bash -e", "-c"]
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "busybox sh is clean on a clean RUN",
			dockerfile: `SHELL ["busybox", "sh", "-c"]
RUN echo ok`,
			fails: false,
		},
		{
			name: "busybox by its path is clean on a clean RUN",
			dockerfile: `SHELL ["/bin/busybox", "sh", "-c"]
RUN echo ok`,
			fails: false,
		},
		{
			name: "busybox ash is busybox sh",
			dockerfile: `SHELL ["/bin/busybox", "ash", "-c"]
RUN echo ok`,
			fails: false,
		},
		{
			name: "busybox sh is still checked",
			dockerfile: `SHELL ["busybox", "sh", "-c"]
RUN echo $MISSING_QUOTES`,
			fails: true,
		},
		{
			name:       "A bashism under the default shell",
			dockerfile: `RUN [[ -f /x ]] && echo ok`,
			fails:      true,
		},
		{
			name: "A shell pragma of one word is the shell",
			dockerfile: `# hadolint shell=bash
RUN [[ -f /x ]] && echo ok`,
			fails: false,
		},
		{
			name: "A clean RUN in an ONBUILD is clean",
			dockerfile: `FROM debian
ONBUILD RUN echo hello`,
			fails: false,
		},
		{
			name: "A SHELL survives an ENV in the same stage",
			dockerfile: `SHELL ["/bin/bash", "-c"]
ENV bar=10
RUN echo $RANDOM`,
			fails: false,
		},
	}

	scRule := shell.NewShellcheckRule(shell.NewBinaryShellchecker())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tt.dockerfile, []rule.Rule{scRule})

			var codes []string

			for _, v := range violations {
				if strings.HasPrefix(string(v.Code), "SC") {
					codes = append(codes, string(v.Code))
				}
			}

			if (len(codes) > 0) != tt.fails {
				t.Fatalf("shellcheck reported %v, want failure = %v", codes, tt.fails)
			}
		})
	}
}
