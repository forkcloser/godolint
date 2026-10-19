package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3041 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3041Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3041(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3041(),
	}

	t.Run(
		"not ok with options that take a separate value and an unpinned package",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y --releasever 9 tomcat`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok with options that take a separate value and an unpinned package (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y --enablerepo crb tomcat`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok with version as variable - different stages, new stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM fedora:fc42 AS build
ENV version=2.51.0-2.fc42
FROM fedora:fc42
RUN dnf -y install git-core-${version}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y tomcat && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y tomcat && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning - modules",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf module install -y tomcat && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning - modules (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf module install -y tomcat && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning - package name with `-`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y rpm-sign && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"not ok without dnf version pinning - package name with `-` (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y rpm-sign && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with a wildcard version",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y rspamd-3.* tini-0.*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with a wildcard version (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y rspamd-*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf group install",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y group install "Development Tools"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf group install (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y --setopt=group_package_types="mandatory" group install "Development Tools"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf group install (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf group install -y "Development Tools" && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf group install (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf group install -y "Development Tools" && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y tomcat-9.0.1 && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y tomcat-9.0.1 && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - modules",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf module install -y tomcat:9 && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - modules (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf module install -y tomcat:9 && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - modules (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN notdnf module install tomcat`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package name with `-`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y rpm-sign-4.16.1.3 && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package name with `-` (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y rpm-sign-4.16.1.3 && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package name with `-` and `+`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y gcc-c++-1.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package name with `-` and `+` (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y gcc-c++-1.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package version with epoch",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y openssl-1:1.1.1k`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with dnf version pinning - package version with epoch (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y openssl-1:1.1.1k`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with options that take a separate value",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y --installroot /mnt/rootfs --releasever 9 --setopt install_weak_deps=false tomcat-9.0.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with options that take a separate value (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y --enablerepo crb --disablerepo epel -x foo tomcat-9.0.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with options that take a separate value (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y --config /etc/dnf.conf tomcat-9.0.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with version as arg - different stages",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM fedora:fc42 AS build
ARG version=2.51.0-2.fc42
FROM fedora:fc42
RUN dnf -y install git-core-${version}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with version as arg - unbraced",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG version=2.51.0-2.fc42
RUN dnf -y install git-core-$version`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with version as env - different stages, reused stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM fedora:fc42 AS build
ENV version=2.51.0-2.fc42
FROM build
RUN dnf -y install git-core-${version}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with version as variable - braced",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ENV version=2.51.0-2.fc42
RUN dnf -y install git-core-${version}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok with version pinning if command is not `dnf` or `microdnf`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN notdnf install openssl-1:1.1.1k`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)

	t.Run(
		"ok without version pinning if command is not `dnf` or `microdnf`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN notdnf install tomcat`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3041")
		},
	)
}
