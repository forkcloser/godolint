# Upstream

Port of <https://github.com/hadolint/hadolint> (`master`), the Dockerfile linter written in
Haskell. Nothing is merged: the rules' metadata and their tests are generated from the
pinned hadolint sources (`pins.yaml`, `just generate`), and each rule's behaviour is
written in Go to pass them.

- Ported from: hadolint [`db9e652`](https://github.com/hadolint/hadolint/commit/db9e652935fd4d77790b8e16eef8300f660b73bf), its `master` of 2025-10-30 (v2.14.0 and after), when this port started
- Ported through: [`8ae25c0`](https://github.com/hadolint/hadolint/commit/8ae25c05945f030b1ad520f1802e4143e7d0055d), hadolint's `master` of 2026-10-14, the pin in `pins.yaml`; reviewed 2026-10-18

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

## Not ported

| Upstream commit | Why |
| --- | --- |
| [`474a638`](https://github.com/hadolint/hadolint/commit/474a638) | Shellcheck: an `ARG` or `ENV` no longer resets the shell a `SHELL` instruction set. This port never reset it (`shellState.addVars`), so the fix is already here. |

## Changes upstream does not have

| Change | Here |
| --- | --- |
| The port's own behaviour: the `# hadolint ignore` pragmas' rules, `--disable-ignore-pragma`, DL1001 reporting every `ignore`, the SDK | README |
