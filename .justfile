# This file is the project's own — add recipes below. Keep the import: it
# mounts every shared limen task under `just do ...`.
import '.limen/just/main.just'

lint: do::lint::default do::lint::go::default do::lint::go::deadcode lint-generated
fix: do::fix::default do::fix::go::default
test: do::test::go::unit do::test::go::race do::test::go::cover
# The security workflow runs `just security`.
security: do::security::default

# Regenerate internal/rules from the hadolint source pins.yaml pins: the
# DLxxxx metadata files and the ported tests are rewritten, stubs are written
# for rules with no implementation. See internal/rules/generate.go.
generate:
    #!/usr/bin/env bash
    set -euo pipefail
    version="$(limen pins get hadolint version)"
    url="$(limen pins get hadolint url)"
    sum="$(limen pins get hadolint sha256)"
    dl=".build/hadolint/hadolint-${version}.tar.gz"
    src=".build/hadolint/hadolint-${version#v}"
    mkdir -p .build/hadolint
    if [ ! -f "${dl}" ]; then
      curl --proto '=https' --tlsv1.2 -fsSL --retry 5 --retry-delay 3 -o "${dl}.part" "${url}"
      mv "${dl}.part" "${dl}"
    fi
    echo "${sum}  ${dl}" | sha256sum -c - >/dev/null
    if [ ! -d "${src}" ]; then
      tar -xzf "${dl}" -C .build/hadolint
    fi
    HADOLINT_SRC="${PWD}/${src}" go generate ./internal/rules

# The committed rules and tests match the pinned hadolint: regenerate, and
# fail on any difference, a changed file or a new one.
lint-generated: generate
    #!/usr/bin/env bash
    set -euo pipefail
    git diff --exit-code --stat -- internal/rules
    untracked="$(git ls-files --others --exclude-standard -- internal/rules)"
    if [ -n "${untracked}" ]; then
      echo "regeneration added files the tree does not have:" >&2
      echo "${untracked}" >&2
      exit 1
    fi
