package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3016 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3016Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3016(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3016(),
	}

	t.Run(
		"commit not pinned for git",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git://github.com/npm/npm.git`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit not pinned for git (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git://github.com/npm/npm.git`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"commit not pinned for git+http",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+http://isaacs@github.com/npm/npm`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit not pinned for git+http (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+http://isaacs@github.com/npm/npm`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"commit not pinned for git+https",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+https://isaacs@github.com/npm/npm.git`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit not pinned for git+https (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+https://isaacs@github.com/npm/npm.git`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"commit not pinned for git+ssh",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+ssh://git@github.com:npm/npm.git`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit not pinned for git+ssh (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+ssh://git@github.com:npm/npm.git`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"commit pinned for git",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git://github.com/npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git://github.com/npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+http",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+http://isaacs@github.com/npm/npm#semver:^5.0`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+http (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+http://isaacs@github.com/npm/npm#semver:^5.0`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+https",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+https://isaacs@github.com/npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+https (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+https://isaacs@github.com/npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+ssh",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+ssh://git@github.com:npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"commit pinned for git+ssh (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install git+ssh://git@github.com:npm/npm.git#v1.0.27`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"don't fire on loglevel flag",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --loglevel verbose sax@0.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"don't fire on loglevel flag (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --loglevel verbose sax@0.1.1`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"don't fire on registry flag",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry https://example.com sax@0.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"don't fire on registry flag (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry https://example.com sax@0.1.1`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"npm run install is fine",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm run --crazy install`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"npm run install is fine (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm run --crazy install`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - absolute path",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install /folder`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - absolute path (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install /folder`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path from current folder",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ./folder`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path from current folder (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ./folder`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path from home",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ~/folder`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path from home (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ~/folder`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path to parent folder",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ../folder`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for folder - relative path to parent folder (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install ../folder`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tar",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tar (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tar.gz",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tar.gz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tar.gz (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tar.gz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tgz",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tgz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version does not have to be pinned for tarball suffix .tgz (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install package-v1.2.3.tgz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version not pinned multiple packages",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express sax@0.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned multiple packages (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express sax@0.1.1`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version not pinned with --global",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --global express`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned with --global (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --global express`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version not pinned with --registry flag",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry https://example.com express`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned with --registry flag (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry https://example.com express`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version not pinned with --registry=value form",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry=https://example.com express`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned with --registry=value form (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --registry=https://example.com express`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version not pinned with scope",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version not pinned with scope (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3016", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"version pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express@4.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express@4.1.1`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned in package.json",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned in package.json (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned in package.json with arguments",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --progress=false`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned in package.json with arguments (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --progress=false`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned multiple packages",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express@"4.1.1" sax@0.1.1`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned multiple packages (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install express@"4.1.1" sax@0.1.1`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with --global",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --global express@"4.1.1"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with --global (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install --global express@"4.1.1"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with -g",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install -g express@"4.1.1"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with -g (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install -g express@"4.1.1"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with scope",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage@">=0.1.0"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with scope (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage@">=0.1.0"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with scope (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage@">=0.1.0 <0.2.0"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)

	t.Run(
		"version pinned with scope [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN npm install @myorg/privatepackage@">=0.1.0 <0.2.0"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3016")
		},
	)
}
