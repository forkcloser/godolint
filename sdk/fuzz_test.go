package sdk_test

import (
	"testing"

	"github.com/forkcloser/godolint/sdk"
)

// FuzzLint: whatever the bytes, Lint returns a result or a ParseError and
// never panics, and the same bytes give the same violations twice.
func FuzzLint(f *testing.F) {
	for _, seed := range []string{
		"FROM debian\n",
		"FROM debian:bookworm AS build\nRUN apt-get update && apt-get install -y curl\nCOPY . /app\nWORKDIR /app\nUSER root\nCMD [\"./app\"]\n",
		"# hadolint ignore=DL3006,DL3007\nFROM ubuntu\n# hadolint stage ignore=DL3002\nFROM alpine\nUSER root\n",
		"# hadolint global ignore=DL3008\nFROM debian:bookworm\nRUN apt-get install curl\n",
		"FROM alpine:3.7\nRUN apk add --no-cache curl && rm -rf /var/cache/apk/*\nEXPOSE 80000\nHEALTHCHECK CMD curl -f http://localhost/ || exit 1\n",
		"FROM scratch\nADD http://example.com/a.tar.gz /\nENV a=b c=d\nLABEL maintainer=\"x\"\nARG V=1\nONBUILD RUN echo\nSHELL [\"/bin/sh\", \"-c\"]\nSTOPSIGNAL SIGTERM\nVOLUME /data\nENTRYPOINT [\"sh\"]\n",
		"FROM debian\nRUN set -e; if [ -f x ]; then echo \"$(cat x)\"; fi; for i in 1 2; do echo $i; done | tee /dev/null\n",
		"RUN before from\n",
		"FROM\n",
		"",
		"\x00\xff",
		"FROM debian\nRUN echo 'unterminated\n",
		"FROM debian\nRUN echo \\\n  continued \\\n  more\n",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, dockerfile []byte) {
		linter := sdk.New()

		first, err := linter.Lint(t.Context(), dockerfile)
		if err != nil {
			var perr *sdk.ParseError
			if !errorsAs(err, &perr) {
				t.Fatalf("an error that is not a ParseError: %v", err)
			}

			return
		}

		if first.Passed != (len(first.Violations) == 0) {
			t.Fatalf("Passed %v with %d violations", first.Passed, len(first.Violations))
		}

		for _, v := range first.Violations {
			if v.Code == "" || v.Line < 0 || v.Message == "" {
				t.Fatalf("a violation with no code, a negative line or no message: %+v", v)
			}
		}

		second, err := linter.Lint(t.Context(), dockerfile)
		if err != nil || len(second.Violations) != len(first.Violations) {
			t.Fatalf(
				"a second Lint of the same bytes: %v, %d violations after %d",
				err,
				len(second.Violations),
				len(first.Violations),
			)
		}
	})
}
