# Upstream

Port of <https://github.com/hadolint/hadolint> (`master`), the Dockerfile linter written in
Haskell. Nothing is merged: the rules' metadata and their tests are generated from the
pinned hadolint sources (`pins.yaml`, `just generate`), and each rule's behaviour is
written in Go to pass them.

- Ported from: hadolint [`db9e652`](https://github.com/hadolint/hadolint/commit/db9e652935fd4d77790b8e16eef8300f660b73bf), its `master` of 2025-10-30 (v2.14.0 and after), when this port started
- Ported through: [`1d901b9`](https://github.com/hadolint/hadolint/commit/1d901b9c0b0d0445d19b1a9128d89fa55bb4604e), hadolint's `master` of 2026-10-19, the pin in `pins.yaml`; reviewed 2026-11-14

`just lint` regenerates from the pin and fails on any difference, so the committed rules and
tests always match it. The range from the fork point to v2.15.1, the first pinned release,
was ported as a whole when the pin was introduced
([#75](https://github.com/forkcloser/godolint/pull/75),
[#78](https://github.com/forkcloser/godolint/pull/78), 2026-10-08) and is not rowed here.
From v2.15.1 on, every upstream commit that changes a rule's behaviour is in one of the
tables below; commits to hadolint's own build, CI, docs and Haskell plumbing are not listed.

## Ported

| Upstream commit | Here | What |
| --- | --- | --- |
| [`e2cd8e4`](https://github.com/hadolint/hadolint/commit/e2cd8e4), [`e3437cf`](https://github.com/hadolint/hadolint/commit/e3437cf), [`7ab1217`](https://github.com/hadolint/hadolint/commit/7ab1217) | `internal/rules/dl3066_impl.go` | DL3066: a user id written as a declared `ARG` or `ENV` is resolved at build time; only the last `USER` of the last stage reaches the host |
| [`1509302`](https://github.com/hadolint/hadolint/commit/1509302) | `internal/rules/dl4006_impl.go`, `shell.Basename` | DL4006: the shell is matched by its basename, so `/usr/bin/bash -o pipefail` counts |
| [`e27bfde`](https://github.com/hadolint/hadolint/commit/e27bfde) | `internal/rules/dl3022_impl.go` | DL3022: `COPY --from=$VAR` names no stage seen here and is not a failure |
| [`f142eb1`](https://github.com/hadolint/hadolint/commit/f142eb1) | `internal/rules/dl3041_impl.go`, `shell.DropFlagArg` | DL3041: the value of a dnf option (`--releasever 9`) is not a package; `*` is a version character |
| [`cf64a44`](https://github.com/hadolint/hadolint/commit/cf64a44) | `shell.CmdHasAllArgs`, DL3009, DL3032, DL3040, DL3060 | A cleanup is the command with every one of its arguments, not any of them: `yum clean` is not `yum clean all` |
| [`23779d4`](https://github.com/hadolint/hadolint/commit/23779d4) | `internal/rules/dl3068_impl.go` | DL3068: pecl packages installed at a version |
| [`9d8ccb7`](https://github.com/hadolint/hadolint/commit/9d8ccb7) | `internal/rules/dl4006_impl.go` | DL4006: a stage built `FROM` a named stage inherits the pipefail its `SHELL` set, until a `SHELL` of its own; an image behind a registry is not a stage |
| [`439b7b6`](https://github.com/hadolint/hadolint/commit/439b7b6) | `internal/rules/dl3064_impl.go` | DL3064: a name ending in `_FILE` holds a path, not a secret; `private` alone is no longer suspicious, `private_key` and `privatekey` are; eight names that are not secrets (`FACEBOOK_APP_ID`, `TRAVIS_OS_NAME` and the like) leave the list |
| [`9495461`](https://github.com/hadolint/hadolint/commit/9495461) | `shell.IsRecursiveRemoval`, DL3009, DL3032, DL3040 | A cache is cleaned by any recursive `rm` of it (`-fr`, `-r -f`, `-R`, `--recursive`), of the directory, with or without a trailing slash, or of its contents; a non-recursive `rm` does not count |
| [`a415cb1`](https://github.com/hadolint/hadolint/commit/a415cb1) | `internal/rules/dl3006_impl.go` | DL3006: a variable inside an image name (`base-${TARGETARCH}`) is a wildcard matched against the earlier stages' aliases; a quoted variable (`"${IMAGE}"`) is a variable |

## Not ported

| Upstream commit | Why |
| --- | --- |
| [`474a638`](https://github.com/hadolint/hadolint/commit/474a638) | Shellcheck: an `ARG` or `ENV` no longer resets the shell a `SHELL` instruction set. This port never reset it (`shellState.addVars`), so the fix is already here. |

## Changes upstream does not have

| Change | Here |
| --- | --- |
| The port's own behaviour: the `# hadolint ignore` pragmas' rules, `--disable-ignore-pragma`, DL1001 reporting every `ignore`, the SDK | README |
| DL3006 compares a `FROM` image with the earlier stages by its whole name, registry included, where hadolint drops the registry first: `FROM myreg.io/base-${ARCH}` beside a stage `base-amd64` is reported here and let through there. An image behind a registry is not a stage | `internal/rules/dl3006_impl.go` |
| DL3006: two variables side by side in an image name (`base-${ARCH}${LIBC}-slim`) match whatever stands between the literals around them, where hadolint stops on an error (`Text.breakOn` with an empty needle) | `internal/rules/dl3006_impl.go` |
| `sdk.Linter.LintFile` reads and lints a file, a path that cannot be read an error, where it panicked by design; an instruction the converter refuses (a `FROM` with no image, an instruction the format does not define, a one-argument `COPY`, a nameless `ARG`) is a parse error naming the line, as hadolint and a build refuse it, where it was skipped and the rest linted; a missing file argument puts the usage on stderr with the error and leaves stdout empty, as the README promises the JSON alone | https://github.com/forkcloser/godolint/pull/117 |
