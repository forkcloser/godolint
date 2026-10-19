package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL4006 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL4006Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL4006(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL4006(),
	}

	t.Run(
		"don't warn on commands with no pipes",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
RUN wget -O - https://some.site && wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn on commands with pipes and the pipefail option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/bash", "-eo", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn on commands with pipes and the pipefail option 2",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/bash", "-e", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn on commands with pipes and the pipefail option 3",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/bash", "-o", "errexit", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn on commands with pipes and the pipefail zsh",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/zsh", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn on powershell",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["pwsh", "-c"]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn when bash is invoked via a non-standard absolute path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/usr/bin/bash", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"don't warn when zsh is invoked via a non-standard absolute path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/usr/local/bin/zsh", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"ignore non posix shells: cmd.exe",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM mcr.microsoft.com/powershell:ubuntu-16.04
SHELL [ "cmd.exe", "/c" ]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"ignore non posix shells: powershell",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM mcr.microsoft.com/powershell:ubuntu-16.04
SHELL [ "powershell.exe" ]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"ignore non posix shells: pwsh",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM mcr.microsoft.com/powershell:ubuntu-16.04
SHELL [ "pwsh", "-c" ]
RUN Get-Variable PSVersionTable | Select-Object -ExpandProperty Value`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"warn on missing pipefail",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN wget -O - https://some.site | wc -l > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"warn on missing pipefail if next SHELL is not using it",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number
SHELL ["/bin/sh", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"warn on missing pipefail in the next image",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number
FROM scratch as build2
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4006")
		},
	)

	t.Run(
		"warns when using plain sh",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch as build
SHELL ["/bin/sh", "-o", "pipefail", "-c"]
RUN wget -O - https://some.site | wc -l file > /number`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4006")
		},
	)
}
