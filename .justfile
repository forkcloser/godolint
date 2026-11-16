# This file is the project's own — add recipes below. Keep the import: it
# mounts every shared limen task under `just do ...`.
import '.limen/just/main.just'

lint: do::lint::default do::lint::go::default do::lint::go::deadcode lint-generated
fix: do::fix::default do::fix::go::default
test: do::test::go::unit do::test::go::race do::test::go::cover
# The security workflow runs `just security`.
security: do::security::default

# Regenerate internal/rules from the sources pins.yaml pins. From hadolint's:
# the DLxxxx metadata files and the ported tests are rewritten, stubs are
# written for rules with no implementation. From the SPDX License List: the
# identifiers DL3054 accepts. See internal/rules/generate.go.
generate:
    #!/usr/bin/env bash
    set -euo pipefail
    # fetch <pin> <file>: the pin's artifact as .build/<pin>/<file>, held to
    # the pin's digest whether it was just downloaded or already there.
    fetch() {
      local url sum dl
      url="$(limen pins get "$1" url)"
      sum="$(limen pins get "$1" sha256)"
      dl=".build/$1/$2"
      mkdir -p ".build/$1"
      if [ ! -f "${dl}" ]; then
        curl --proto '=https' --tlsv1.2 -fsSL --retry 5 --retry-delay 3 -o "${dl}.part" "${url}"
        mv "${dl}.part" "${dl}"
      fi
      echo "${sum}  ${dl}" | sha256sum -c - >/dev/null
    }
    commit="$(limen pins get hadolint commit)"
    src=".build/hadolint/hadolint-${commit}"
    fetch hadolint "hadolint-${commit}.tar.gz"
    if [ ! -d "${src}" ]; then
      tar -xzf "${src}.tar.gz" -C .build/hadolint
    fi
    spdx="licenses-$(limen pins get spdx-license-list version).json"
    fetch spdx-license-list "${spdx}"
    HADOLINT_SRC="${PWD}/${src}" SPDX_LICENSES="${PWD}/.build/spdx-license-list/${spdx}" \
      go generate ./internal/rules

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
