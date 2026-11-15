// Package main demonstrates integration with shellcheck for RUN instruction validation.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/forkcloser/godolint/sdk"
)

func main() {
	dockerfile := `FROM debian:bookworm@sha256:6ebd97fa83deb272194a2cf015b3d26a4d538e9ad3a7a79d544c8af5b0a01443
ENV MY_VAR=value
ARG BUILD_ARG=default
SHELL ["/bin/bash", "-c"]

# Not reported: shellcheck is told the variables ENV and ARG set
RUN echo $MY_VAR

# SC2086: an unquoted variable shellcheck knows nothing of
RUN echo $UNDEFINED_VAR

# Nothing for shellcheck here
RUN apt-get update && apt-get install -y curl
`

	// WithShellcheck adds shellcheck's findings on the RUN instructions to
	// the rules of the linter. It needs the shellcheck binary in PATH.
	linter := sdk.New(sdk.WithShellcheck())

	result, err := linter.Lint(context.Background(), []byte(dockerfile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to lint: %v\n", err)
		os.Exit(1)
	}

	// Shellcheck's codes are SC and four digits; the others are godolint's own.
	fmt.Println("Shellcheck violations:")

	for _, v := range result.Violations {
		if strings.HasPrefix(v.Code, "SC") {
			fmt.Printf("Line %d: [%s] %s - %s\n", v.Line, v.Severity, v.Code, v.Message)
		}
	}
}
