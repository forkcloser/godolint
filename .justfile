# This file is the project's own — add recipes below. Keep the import: it
# mounts every shared limen task under `just do ...`.
import '.limen/just/main.just'

lint: do::lint::default do::lint::go::default do::lint::go::deadcode do::lint::converge
fix: do::fix::default do::fix::go::default
test: do::test::go::unit do::test::go::race do::test::go::cover
# The security workflow runs `just security`.
security: do::security::default

# Make the files .converge names from the sources pins.yaml pins. From
# hadolint's: the DLxxxx metadata files and the ported tests are rewritten,
# stubs are written for rules with no implementation. From the SPDX License
# List: the identifiers DL3054 accepts. See internal/rules/generate.go.
# `just do lint converge` holds the tree to what this makes, and on a Renovate
# branch the checksum workflow commits it.
converge:
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
    # A file the generators no longer write must not stay: every declared
    # file goes first, once the sources are there.
    while IFS= read -r line || [ -n "${line}" ]; do
      path="${line%$'\r'}"
      case "${path}" in
        '' | '#'*) continue ;;
      esac
      rm -f -- "${path}"
    done < .converge
    HADOLINT_SRC="${PWD}/${src}" SPDX_LICENSES="${PWD}/.build/spdx-license-list/${spdx}" \
      go generate ./internal/rules
