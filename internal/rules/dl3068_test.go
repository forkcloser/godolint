package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3068 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3068Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3068(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3068(),
	}

	t.Run(
		"alpha state not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install xdebug-alpha`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"beta state not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-beta`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"define value does not mask an unpinned package",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install -D 'with-foo=yes' redis`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"define value is not treated as a package",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install -D 'with-foo=yes' redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"local package.xml descriptor",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install package.xml`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"local package.xml descriptor with a path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install /src/ext/package.xml`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"local tgz artifact",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-6.1.0.tgz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"local tgz artifact with a relative path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install ./build/redis-6.1.0.tgz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"local tgz artifact with an absolute path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install /tmp/redis-6.1.0.tgz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"multiple packages pinned with variables",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-${R_VER} xdebug-$X_VER`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"package name containing an underscore, pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl_http-4.2.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"package name with underscore, not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl_http`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"pecl run install is fine",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl run --crazy install`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"remote https download",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install https://example.com/redis.tgz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"single-digit version pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-6`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"stable state not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-stable`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version not pinned multiple packages",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install xdebug redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version not pinned multiple packages and channel",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl.php.net/xdebug redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version not pinned with a channel URI",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install channel://pecl.php.net/redis`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version not pinned with channel",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl.php.net/redis`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a braced shell variable",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-${REDIS_VERSION}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a channel URI",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install channel://pecl.php.net/redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a leading force flag",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install -f redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a shell variable",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-$REDIS_VERSION`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a short channel alias",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install mychannel/redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with a variable and channel",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl.php.net/xdebug-$XDEBUG_VERSION`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with arguments",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install redis-6.1.0 --phpize`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)

	t.Run(
		"version pinned with channel",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN pecl install pecl.php.net/redis-6.1.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3068")
		},
	)
}
