package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3064 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3064Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3064(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3064(),
	}

	t.Run(
		"not ok: ARG with a private key",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG SSH_PRIVATE_KEY
FROM debian:bullseye
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"not ok: ARG with sensitive data",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG AWS_ACCESS_KEY_ID
FROM debian:bullseye
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"not ok: ARG with sensitive data, different casing",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG openai_api_key
FROM debian:bullseye
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"not ok: ENV with a sensitive name among other pairs",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian:bullseye
ENV GOPRIVATE=github.com/example/* GITHUB_TOKEN=abcdef
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"not ok: ENV with sensitive data",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian:bullseye
ENV AWS_ACCESS_KEY_ID=abcdefghijklmnopqrstuvwxyz
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"not ok: ENV with sensitive data, different casing",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian:bullseye
ENV my_password=abcdefghijklmnopqrstuvwxyz
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"ok: ARG and ENV referencing a secret file",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM postgres:17
ARG POSTGRES_PASSWORD_FILE=/run/secrets/db_password
ENV AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/token
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"ok: ARG no sensitive data, no ENV",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG foobar
FROM debian:bullseye
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"ok: ENV with names that only resemble sensitive data",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM golang:1.24
ENV GOPRIVATE=github.com/example/*
ENV TRAVIS_OS_NAME=linux TRAVIS_SUDO=false
ENV FACEBOOK_APP_ID=1234567890
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"ok: no ARG, ENV no sensitive data",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian:bullseye
ENV foobar=barfoo
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3064")
		},
	)

	t.Run(
		"ok: no ARG, no ENV",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian:bullseye
RUN foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3064")
		},
	)
}
