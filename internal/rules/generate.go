package rules

// Generate rule stubs and tests from hadolint Haskell sources.
//
// Run: just generate. It fetches the hadolint source pins.yaml names, by
// digest, into .build/, and exports HADOLINT_SRC, its root, before running
// go generate here; the directives below read the path from that variable,
// so a bare go generate fails on a missing source rather than reading a
// checkout of no known version.
//
// This does two things:
// 1. Extracts metadata (code, severity, message) from hadolint .hs rule files
//    and generates Go stub files for unimplemented rules
// 2. Extracts test cases from hadolint test files and generates Go tests
//    for implemented rules
//
// Rule stubs include placeholder check functions that need manual implementation.
// Tests are automatically ported from hadolint's test suite.

//go:generate go run ../../hack/gen-rules/main.go $HADOLINT_SRC/src/Hadolint/Rule
//go:generate go run ../../hack/gen-tests/main.go $HADOLINT_SRC/test/Hadolint/Rule
