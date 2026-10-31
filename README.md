![godolint logo](logo.png)

# godolint

> Same shit, different language

## godolint? hadolint?

This project is a friendly port of [hadolint](https://github.com/hadolint/hadolint) (a venerable and wildly used dockerfile linter
written in haskell).

While there is nothing wrong with hadolint in itself, it does not fit well inside the
go devtools ecosystem (*):
- it is a foreign dependency, that has to be handled outside the familiar go toolchain
- it comes as a binary, that you have to shell out to, and parse the output of
- modifying it requires haskell knowledge

godolint is a pure go port of hadolint, filling an age-old gap in the container tooling landscape:
the ability to lint dockerfiles in go.

It comes with an example binary that does mimic a small subset of the hadolint binary behavior,
and can also be used (of course, and primarily) as a library in Go projects.

(*) see [Why?](#why) section below for more details

## Installation

### As a CLI tool

```bash
go install github.com/forkcloser/godolint/cmd/godolint@latest
```

### As a library

```bash
go get github.com/forkcloser/godolint
```

## Quick Start

### CLI Usage

```bash
# Lint a Dockerfile
godolint Dockerfile

# Output: JSON array of violations
# [
#   {
#     "Code": "DL3007",
#     "Severity": "warning",
#     "Message": "Using latest is prone to errors if the image will ever update",
#     "Line": 1
#   }
# ]

# Ignore specific rules
godolint --ignore DL3006 --ignore SC2050 Dockerfile

# Disable the shellcheck integration for RUN instructions
godolint --without-shellcheck Dockerfile

# Point the shellcheck integration at a configuration file — without this,
# shellcheck never finds a repository's .shellcheckrc, since the checked
# scripts run from a temp dir (requires shellcheck >= 0.10.0)
godolint --shellcheck-rcfile .shellcheckrc Dockerfile

# Logs go to stderr (stdout is the JSON only): --log-level or LOG_LEVEL,
# one of debug, info (default), warn, error
godolint --log-level debug Dockerfile
```

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

`ignore` covers the next line, `stage ignore`, written above a `FROM`, the
stage that `FROM` starts, `global ignore` the whole file. Spaces around the `=` and the
commas, and a `# comment` after the list, are fine. A list with a name that is
not a rule code (`ignore=crazy,DL3002`) is not a pragma at all, as in hadolint.
`--disable-ignore-pragma` turns the three off. DL1001 reports every `ignore`
pragma (and only that kind) at the `ignore` severity, which the CLI and the SDK
never print.

### SDK Usage

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/forkcloser/godolint/sdk"
)

func main() {
    // Read Dockerfile
    content, _ := os.ReadFile("Dockerfile")

    // Create linter with default configuration (all rules)
    linter := sdk.New()

    // Lint the Dockerfile
    result, err := linter.Lint(context.Background(), content)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Report violations
    for _, v := range result.Violations {
        fmt.Printf("[%s] %s (line %d): %s\n",
            v.Severity, v.Code, v.Line, v.Message)
    }

    // Exit with error code if violations found
    if !result.Passed {
        os.Exit(1)
    }
}
```

### Advanced SDK Usage

```go
// Use a specific rule set
linter := sdk.New(sdk.WithRuleSet(sdk.RuleSetRecommended))

// Disable specific rules
linter := sdk.New(sdk.WithDisabledRules("DL3007", "DL3008"))

// Enable shellcheck integration
linter := sdk.New(sdk.WithShellcheck())

// Enable shellcheck integration with a specific configuration file
linter := sdk.New(sdk.WithShellcheck(sdk.WithShellcheckRCFile(".shellcheckrc")))

// Check for specific severity levels
if result.HasErrors() {
    fmt.Println("Critical issues found!")
}

counts := result.CountBySeverity()
fmt.Printf("Errors: %d, Warnings: %d\n",
    counts[sdk.SeverityError],
    counts[sdk.SeverityWarning])
```

## Why?

> What is wrong with depending on (hadolint) binaries?

### TL;DR

> copy-paste into a powerpoint for your CISO (don't forget to add ✅ and ❌)

| Aspect | External Binary | Pure Go |
|--------|----------------|---------|
| **Supply Chain Risk** | HIGH - Unverified provenance | LOW - Checksummed source |
| **Compromise Scenarios** | Account takeover, build pipeline, MITM | Requires both GitHub + sum.golang.org |
| **Cryptographic Verification** | Manual, often missing | Automatic via go.sum |
| **Reproducibility** | Poor - binary availability not guaranteed | Excellent - source-based |
| **Multi-arch Support** | Complex per-platform logic | `GOARCH=arm64 go build` |
| **Installation Complexity** | Bash scripts, sudo, platform detection | `go install` |
| **CI/CD Integration** | Manual download + caching | Native Go tooling |
| **Container Image Size** | +50-100MB (deps + binary) | +5-10MB (static binary) |
| **SBOM Generation** | Manual, incomplete | `go list -m all` |
| **Vulnerability Scanning** | Opaque binary, hard to scan | `govulncheck ./...` |
| **License Compliance** | Manual transitive analysis | `go-licenses report` |
| **TCB Size** | Large (GitHub, CDN, toolchain) | Small (Go compiler, go.sum) |
| **Offline Builds** | Impossible | `go mod vendor` |
| **Dependency Updates** | Manual version bumps | `go get -u` + Dependabot |

### With more words

Supply chain concerns:
* binary provenance is unverifiable
* no cryptographic verification
* version pinning is fragile
* SBOM is incomplete or missing
* license compliance is opaque / hardcoded
* trusted computing base is WIDE
* scanning binaries is difficult

Operational complexity:
* multi-architecture is hell
* installation script is hell with fries
* CI/CD compounds hell further into full-blown hell
* availability is not guaranteed
* os-level dependencies if not compiled statically (which you have to figure out)

You COULD (should!) instead build from source (that is, if you like Gentoo).
At least you would get proper git commit pinning, scannable / auditable source code, SBOM.
But then, you have to maintain a build environment for it, in a language and with tools
that does not match the rest of your stack, and that may also raise additional supply
chain concerns, time and dedicated expertise.

So, yes, there is nothing wrong with hadolint-the-project-and-tool, per-se.
But if what you do is more than noodling around, if you are not a haskell shop,
or if you take your security seriously, depending on external binaries is definitely not
something you should do.

This is why we wrote godolint, for people doing go.

## Relationship with hadolint

godolint is a faithful port of hadolint to Go. Rule declarations and tests are generated from hadolint's Haskell source, at the version the `hadolint` entry in [pins.yaml](./pins.yaml) pins by digest: `just generate` fetches that source and regenerates `internal/rules`, and `just lint` fails when the committed rules and tests differ from it.

**Current status:**
- All unit tests are automatically converted to Go tests
- Rule metadata (code, severity, message) matches hadolint exactly
- Every DL rule is implemented

**Project philosophy:**
- hadolint is the upstream for rule definitions and baseline tests
- We have no intention of fragmenting the community
- New **core** rule requests should go through hadolint
- We will regularly sync with hadolint's evolution

## SDK API

### Core Types

```go
// Linter performs Dockerfile linting
type Linter struct { ... }

// Result contains linting results
type Result struct {
    Violations []Violation
    Passed     bool
}

// Violation represents a single rule violation
type Violation struct {
    Code     string   // Rule code (e.g., "DL3000")
    Severity Severity // error, warning, info, style
    Message  string   // Human-readable description
    Line     int      // Line number (1-indexed)
}
```

### Configuration Options

```go
// WithRuleSet - Use a predefined rule set
sdk.New(sdk.WithRuleSet(sdk.RuleSetRecommended))

// WithDisabledRules - Disable specific rules
sdk.New(sdk.WithDisabledRules("DL3007", "DL3008"))

// WithShellcheck - Enable shellcheck integration
sdk.New(sdk.WithShellcheck())
```

### Rule Sets

- `RuleSetAll` - All implemented rules (default)
- `RuleSetRecommended` - Only Error and Warning severity rules
- `RuleSetStrict` - Same as All (for compatibility)

### Result Methods

```go
result.HasErrors()           // Check for error-severity violations
result.HasWarnings()         // Check for warning-severity violations
result.CountBySeverity()     // Get violation counts by severity
```

### Typed Errors

```go
*sdk.ParseError    // Dockerfile parsing failed
*sdk.RuleError     // Rule execution failed
```

### Output Format

godolint outputs JSON arrays of violations:

```json
[
  {
    "Code": "DL3007",
    "Severity": "warning",
    "Message": "Using latest is prone to errors if the image will ever update",
    "Line": 1
  },
  {
    "Code": "DL3000",
    "Severity": "error",
    "Message": "Use absolute WORKDIR",
    "Line": 5
  }
]
```

Exit codes:
- `0`: No violations
- `1`: Violations found — or a usage or runtime error, reported on stderr with nothing on stdout

## Architecture

### Parser

godolint uses [moby/buildkit](https://github.com/moby/buildkit) for parsing Dockerfiles, ensuring compatibility with Docker's official parser.

The parser interface is pluggable:

```go
type Parser interface {
    Parse(dockerfile []byte) ([]syntax.InstructionPos, error)
}
```

### Shell Script Validation

godolint includes full shellcheck integration for validating shell commands in RUN instructions:

- **Stateful tracking** - Tracks ENV, ARG, and SHELL instructions across the Dockerfile
- **Multi-stage support** - A `FROM` starts from the defaults, or from the stage it names
  (`FROM base` after `FROM debian AS base` keeps base's `SHELL`, `ENV` and `ARG`)
- **Shell pragma** - `# hadolint shell=powershell` sets the shell for the stage, or for
  every stage when written before the first instruction
- **Smart skipping** - Automatically skips non-POSIX shells (PowerShell, cmd) and the exec
  form (`RUN ["cmd", "arg"]`), which no shell runs
- **Complete context** - Constructs scripts with proper shebang and environment exports

For external scripts validation, godolint does shell out to shellcheck.

This is an optional feature, that you have to require explicitly (as it does present the same
challenges as hadolint).

We also believe this should stay disabled, and that external shell scripts linting should not
be done at all by godolint (or hadolint). Similarly, neither hadolint nor godolint do lint external
yaml, or typescript files, so, linting shell feels inconsistent, and can easily be done
separately entirely.

ShellCheck violations, if enabled, will be reported with SC#### codes alongside DL#### rules.

### Rule Engine

Rules implement a stateful interface allowing cross-instruction validation:

```go
type Rule interface {
    Code() RuleCode              // "DL3007"
    Severity() Severity          // Error, Warning, Info, Style
    Message() string             // Human-readable description
    InitialState() State         // Initialize rule state
    Check(line int, state State, instruction Instruction) State
    Finalize(state State) State  // Process final state
}
```

Rules can maintain state across instructions for complex validations (multi-stage builds, shell context tracking, etc.).

### Instruction Types

The AST covers all Dockerfile instructions:
- `FROM` (image, tag, digest, platform, alias)
- `RUN`, `CMD`, `ENTRYPOINT` (commands)
- `COPY`, `ADD` (sources, destination, flags)
- `ENV`, `ARG`, `LABEL` (key-value pairs)
- `WORKDIR`, `USER`, `EXPOSE`, `VOLUME`
- `HEALTHCHECK`, `STOPSIGNAL`, `SHELL`
- `MAINTAINER`, `ONBUILD`

## Rules & Tests

Rules and tests are (partly) generated from the hadolint codebase.

As such, godolint passes hadolint's internal test suite, (hopefully) guaranteeing
a seamless transition.

Further, as we do not plan on fragmenting the ecosystem for no good reason, we use exactly the same rules,
and plan on keeping up with hadolint's updated/new rules.

### Code Generation

Rule stubs and tests are auto-generated from hadolint's source:

```bash
# Generate rule stubs and tests (runs both generators)
go generate ./internal/rules
```

This single command:
- Extracts rule metadata (code, severity, message) from Haskell source
- Generates Go function stubs for unimplemented rules
- Converts Haskell test cases to Go tests for implemented rules
- Reports implementation status

## Development

### Build

```bash
# Build CLI binary
just do build go

# Or directly
go build -o godolint ./cmd/godolint
```

### Test

```bash
# Run all tests
just test

# Run specific rule tests
go test ./internal/rules -run DL3007
```

### Lint

```bash
# Run linters (tools are pinned by aqua: `aqua install --only-link` once)
just lint
```

### Adding New Rules

1. Check if rule stub exists in `internal/rules/dlXXXX.go`
2. If not, run `go generate ./internal/rules` to generate stub
3. Implement the `checkDLXXXX` function
4. Add rule to `cmd/godolint/main.go` registration
5. Run tests: `go test ./internal/rules -run DLXXXX`

Example:

```go
// internal/rules/dl3007.go
func checkDL3007(instruction syntax.Instruction) bool {
    from, ok := instruction.(*syntax.From)
    if !ok {
        return true // Not a FROM instruction, passes
    }

    // Fail if using :latest tag
    if from.Image.Tag != nil && *from.Image.Tag == "latest" {
        return false
    }

    return true
}
```

## Releasing

A release is a signed, annotated tag `vX.Y.Z` on `main`; no binaries are
published. Consumers pin the tag as a `go_install` package (through aqua) or
with `go install github.com/forkcloser/godolint/cmd/godolint@vX.Y.Z`.

## Limitations

- The command writes JSON only, and neither it nor the SDK takes a
  configuration: the configurable rules (DL3026's allowed registries,
  DL3049 to DL3056's label schema) run on their defaults. Output formats:
  https://github.com/forkcloser/godolint/issues/127; configuration:
  https://github.com/forkcloser/godolint/issues/125.
- Quote handling differs slightly from hadolint (buildkit includes quotes in
  values).

## Contributing

Contributions welcome; the open issues are the backlog:
https://github.com/forkcloser/godolint/issues.

## License

See LICENSE file.

