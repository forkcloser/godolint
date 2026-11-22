![godolint logo](logo.png)

# godolint

> Same shit, different language

A Dockerfile linter in Go: the rules of [hadolint](https://github.com/hadolint/hadolint), its configuration file, its ignore pragmas and its JSON output, as a command and as a Go package.

- Every rule hadolint has at the commit [pins.yaml](./pins.yaml) pins, under hadolint's code, severity and message. The rules' metadata and tests are generated from hadolint's source; `just lint` fails when they differ from it.
- The `RUN` instructions are checked with [shellcheck](https://github.com/koalaman/shellcheck), run as a program, when there is one.
- Where hadolint's rule and what Docker does disagree (the build, and the commands a Dockerfile runs), godolint follows Docker, and every place it answers differently from hadolint is a row in [UPSTREAM.md](./UPSTREAM.md).

## Install

```bash
go install github.com/forkcloser/godolint/cmd/godolint@<version>
```

A [release](https://github.com/forkcloser/godolint/releases) is a signed tag `vX.Y.Z` with the command built for Linux, macOS and Windows, and a `checksums.txt` signed with sigstore. As a Go package:

```bash
go get github.com/forkcloser/godolint@<version>
```

## The command

```bash
godolint Dockerfile
godolint --ignore DL3006 --ignore SC2050 Dockerfile build/Dockerfile.release
```

### Output

One JSON array on stdout, on one line, with an object for each violation, and `[]` when there is none. Indented here:

```json
[
  {
    "file": "Dockerfile",
    "line": 1,
    "column": 1,
    "level": "warning",
    "code": "DL3007",
    "message": "Using latest is prone to errors if the image will ever update. Pin the version explicitly to a release tag"
  }
]
```

`file` is the path as it was given, `column` is always 1, and `level` is `error`, `warning`, `info` or `style`. `line` is the first line of the instruction, but for what shellcheck finds in the body of a heredoc, which is on the line it stands on. The violations of a file are ordered by line, and by code on one line; several files come one after the other, in the order they were given.

This is hadolint's `json` format, and the only one: there is no `--format` and no `--output`, and a configuration file that asks for another of hadolint's formats gets the JSON, with a warning on standard error. A caller that wants another shape reads the JSON.

Logs go to stderr: `--log-level` or `LOG_LEVEL`, one of `debug`, `info` (the default), `warn`, `error`.

### Exit status

- `0`: no violation at the failure threshold or graver. The threshold is `info` unless `--failure-threshold` or the configuration file names another, as in hadolint: a run whose violations are all of the `style` severity exits 0, with the violations on stdout. With `--no-fail`, no violation makes the exit status 1.
- `1`: a violation at the failure threshold or graver; or a usage or runtime error (a file that cannot be read or parsed, a configuration that cannot be obeyed, a `RUN` shellcheck did not scan), reported on stderr with nothing on stdout, whatever `--no-fail` says.

### Settings

godolint takes hadolint's settings, under hadolint's flags and from hadolint's own configuration file:

```bash
# Give a rule's violations another severity (--error, --warning, --info, --style)
godolint --error DL3006 --style DL3059 Dockerfile

# Exit 1 only for a violation of this severity or a graver one (default: info)
godolint --failure-threshold warning Dockerfile

# Never exit 1 for a violation
godolint --no-fail Dockerfile

# The registries images may come from (DL3026)
godolint --trusted-registry docker.io --trusted-registry '*.example.com' Dockerfile

# The labels every image must carry and how each is written (DL3049, DL3051 to
# DL3056, DL3058), and no other label (DL3050)
godolint --require-label maintainer:email --require-label version:semver --strict-labels Dockerfile

# A configuration file, and which one is used said on stderr
godolint --config .hadolint.yaml --verbose Dockerfile
```

```yaml
# .hadolint.yaml
ignored:
  - DL3008
override:
  error:
    - DL3006
  style:
    - DL3059
trustedRegistries:
  - docker.io
  - "*.example.com"
label-schema:
  maintainer: email # text, url, spdx, hash, rfc3339, semver or email
  version: semver
strict-labels: true
disable-ignore-pragma: false
failure-threshold: warning # error, warning, info, style, ignore or none
no-fail: false
```

Without `--config`, the file used is the first that exists of `.hadolint.yaml` and `.hadolint.yml` in the working directory; `hadolint.yaml` and `hadolint.yml` in `$XDG_CONFIG_HOME` (or `~/.config`); those two, `config.yaml` and `config.yml` in `~/.hadolint`; `.hadolint.yaml` and `.hadolint.yml` in the home. On Windows the three directories are `%APPDATA%`, `%APPDATA%\hadolint` and `Documents` under the user's profile. The file and the command line add up as in hadolint: a list is the file's and the command line's together, a switch is on when either turns it on, the failure threshold is the command line's when it gives one, and a label both name keeps the file's type. A code given several severities gets the gravest.

A file hadolint would not take as its configuration stops the run, where hadolint prints the error and lints with its defaults: a format, a label type or a threshold that is not one, a label left with no type, YAML that does not parse. What hadolint takes and godolint does not act on is a warning on standard error naming it, and the rest of the file is obeyed: a key hadolint does not have, which hadolint passes over without a word (YAML's merge key `<<` is one: hadolint reads YAML 1.2, where it merges nothing), and a `format` or `formats` of hadolint's other than `json`. `no-color` is accepted: the JSON has no colour. hadolint's environment variables (`HADOLINT_IGNORE`, `HADOLINT_FAILURE_THRESHOLD` and the others, `NO_COLOR`) are not read.

### Ignore pragmas

A Dockerfile silences rules with hadolint's pragmas, in a comment:

```dockerfile
# hadolint global ignore=DL3008
FROM debian:bookworm
# hadolint ignore=DL3006,DL3007
FROM ubuntu
# hadolint stage ignore=DL3002
FROM alpine
USER root
```

`ignore` covers the instruction on the next line, down to its last line (the body of a heredoc included), `stage ignore`, written above a `FROM`, the stage that `FROM` starts, `global ignore` the whole file. Spaces around the `=` and the commas, and a `# comment` after the list, are fine. A list with a name that is not a rule code (`ignore=crazy,DL3002`) is not a pragma at all, as in hadolint. `--disable-ignore-pragma` turns the three off. DL1001 reports every `ignore` pragma (and only that kind) at the `ignore` severity, which is never printed unless the rule is given another.

### Shellcheck

The command checks the `RUN` instructions with shellcheck when there is one in `PATH`, and reports what it finds under shellcheck's codes (`SC2086`) beside the rules'.

- With no shellcheck in `PATH`, it lints without it and says so on stderr, at every log level: the exit status is that of the other rules. `--without-shellcheck` lints without it and without the warning.
- A shellcheck that is there and does not scan a `RUN` (an exit status other than 0 and 1, no answer in thirty seconds, or no report on its output, as a wrapper that fails in its place leaves) is a runtime error.
- `--shellcheck-rcfile FILE` names the configuration file shellcheck reads (shellcheck 0.10.0 and later). Without it shellcheck finds none: the script it is shown is a temporary file, outside the repository. A file that cannot be opened is an error.
- The script shellcheck reads is the `RUN`'s, under the stage's shell (`SHELL`, or `# hadolint shell=bash` above an instruction: for the stage, or for every stage when written before the first instruction) and with the names of its `ARG` and `ENV`. A `FROM` that names an earlier stage starts from that stage's shell and names.
- Not read: the exec form (`RUN ["cmd", "arg"]`), which no shell runs; a `RUN` under PowerShell or `cmd`; a heredoc whose first line names an interpreter that is no shell.
- shellcheck leaves out SC2187, SC1090 and SC1091, as hadolint has it.

## The Go package

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/forkcloser/godolint/sdk"
)

func main() {
	result, err := sdk.New().LintFile(context.Background(), "Dockerfile")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, v := range result.Violations {
		fmt.Printf("%d: %s (%s) %s\n", v.Line, v.Code, v.Severity, v.Message)
	}

	if !result.Passed {
		os.Exit(1)
	}
}
```

`sdk.New()` has every rule and no shellcheck. `Lint` takes the Dockerfile's bytes, `LintFile` a path. A `Result` holds the `Violations`, ordered by line and by code on one line, and `Passed`, true when there is none; `HasErrors`, `HasWarnings` and `CountBySeverity` read them. A `Violation` carries its `Code`, `Severity`, `Message` and 1-based `Line`, and marshals to JSON under those names in lower case.

### Options

```go
// The error and warning rules alone
sdk.New(sdk.WithRuleSet(sdk.RuleSetRecommended))

// Rules that are not run, and whose codes reach no Result
sdk.New(sdk.WithDisabledRules("DL3007", "DL3008"))

// Another severity for the violations of a code
sdk.New(sdk.WithSeverity(sdk.SeverityError, "DL3006", "SC2086"))

// The labels an image must carry, and how each is written (DL3049, DL3051
// to DL3056, DL3058)
sdk.New(sdk.WithLabelSchema(map[string]sdk.LabelType{
	"maintainer":                       sdk.LabelEmail,
	"org.opencontainers.image.version": sdk.LabelSemVer,
}))

// A label the schema does not name is a violation (DL3050)
sdk.New(sdk.WithStrictLabels())

// The registries an image may come from (DL3026)
sdk.New(sdk.WithTrustedRegistries("docker.io", "*.example.com"))

// Read no `# hadolint ignore=` pragma
sdk.New(sdk.WithDisableIgnorePragma())

// shellcheck on the RUN instructions, with the checks it leaves out (in
// place of SC2187, SC1090 and SC1091) and its configuration file
sdk.New(sdk.WithShellcheck(
	sdk.WithShellcheckExclusions("SC2187"),
	sdk.WithShellcheckRCFile(".shellcheckrc"),
))
```

These are hadolint's settings under hadolint's meanings: `ignored`, `override`, `label-schema` (the types `text`, `url`, `spdx`, `hash`, `rfc3339`, `semver`, `email`), `strict-labels`, `trustedRegistries` and `disable-ignore-pragma`. A code given to `WithDisabledRules` or `WithSeverity` may be one shellcheck reports (`SC2086`). A code given several severities gets the gravest, and a disabled code stays disabled whatever severity it is given. The options hold in any order: `WithShellcheck` and `WithDisabledRules` mean the same before `WithRuleSet` and after it. `WithLabelSchema` and `WithTrustedRegistries` given again add to what was given; `WithShellcheck` given again replaces the first, with its options alone.

### Rules

```go
// Every rule, ordered by code
for _, info := range sdk.Rules() {
	fmt.Println(info.Code, info.Severity, info.Message)
}

// One rule by its code
info, ok := sdk.Rule("DL3006")
```

A `RuleInfo` carries the rule's `Code`, the `Severity` every violation of it has unless `WithSeverity` gives the code another, and its `Message` in hadolint's words. A rule at `SeverityIgnore` (DL1001, DL3057) is off: its violations never reach a `Result` unless `WithSeverity` gives it a severity. A violation's message is its rule's, but for a rule about a label, which says `<label>` where the violation names the label (``Label `<label>` is missing.``); any other name in angle brackets (`<package>`, `<version>`) is the violation's text too. The codes shellcheck reports are not rules of this list.

### Errors

```go
*sdk.ParseError      // The Dockerfile cannot be parsed
*sdk.ConfigError     // An option was given a value it does not take
*sdk.ShellcheckError // shellcheck, asked for, did not scan a RUN
```

`Lint` returns one of them and no `Result`, or the context's error when the context is done before the lint starts.

- A `ConfigError` names the option and the value: a severity or a label type that is none of the constants, a shellcheck configuration file that cannot be opened. Of several, the first option's is returned.
- A `ShellcheckError` names the line of the `RUN` shellcheck did not scan: no shellcheck in `PATH`, an exit status that is no scan's (shellcheck exits 3 for an option it does not take, as one older than 0.10.0 does for the configuration file), no answer in thirty seconds, or no report on its output. A `Result` would say of that `RUN` what was not checked. A Dockerfile with no `RUN` for shellcheck to read is linted either way.

## hadolint and godolint

godolint follows hadolint's `master`, at the commit the `hadolint` entry of [pins.yaml](./pins.yaml) pins by digest. `just converge` fetches that source and writes the files [.converge](./.converge) names in `internal/rules`: for each rule its code, severity and message, and hadolint's own tests for it, every `it` of its spec. What a rule does is written in Go, by hand, to pass them. `just lint` regenerates and fails on a difference, so the committed rules and tests are always those of the pin; a block of hadolint's tests the generator cannot read stops it, and a case godolint does not pass is named in the generator, so no case is left out without a word.

New rules belong in hadolint: godolint ports what hadolint has, and adds none of its own.

[UPSTREAM.md](./UPSTREAM.md) records the commit ported from and through, each upstream commit that changes a rule, and every place godolint differs: where it follows what a build does and hadolint does not, where it keeps a verdict hadolint misses, and what the port leaves out (hadolint's other output formats, its environment variables).

## Development

```bash
just lint      # every linter, and the generated files held to the pin
just test      # the tests, with the race detector and coverage
just fix       # formatting
just converge  # regenerate from the sources pins.yaml pins
```

A rule is three files in `internal/rules`: `dlNNNN.go` (generated: the metadata), `dlNNNN_impl.go` (by hand: what the rule does), `dlNNNN_test.go` (generated: hadolint's tests); tests of our own are in `dlNNNN_hand_test.go`. When hadolint gets a rule, `just converge` writes its metadata, its tests and a stub; the rule is then written in the `_impl.go` file and added to `internal/rules/all.go`.

## Releasing

A release is a signed, annotated tag `vX.Y.Z` on `main`; the release workflow builds the command for the platforms of [.release-go.yaml](./.release-go.yaml) and publishes them with their signed checksums. Its notes are the titles of the pull requests merged since the last one.

## License

See [LICENSE](./LICENSE).
