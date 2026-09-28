# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## What this is

`golin` switches the Go SDK a machine uses by re-pointing a symbolic link (default name `current`) that `GOROOT` is expected to be set to. Layout it manages:

```
{root}/1.16.5
{root}/1.22.0
{root}/current -> 1.22.0     # GOROOT points here
```

On linux/mac the link is a symlink, so it needs `sudo golin ...` (with `Defaults env_keep += "GOROOT"` in sudoers so `GOROOT` survives). On Windows it is an NTFS junction, which needs no elevation — see "Link creation" below.

## Commands

```powershell
go build ./...
go vet ./...
go test ./...                          # TestDownload fetches golang.org/dl/go1.12 over the network
go test -short ./...                   # skips TestDownload
go test -run TestVersion ./...         # single test
go run ./_cmd/golin list               # run the CLI from source
```

Release build (all 6 GOOS/GOARCH targets, zips into `_cmd/build/`) is `_cmd/release.sh`, run from inside `_cmd/`. It stamps `main.version/revision/date/build` via `-ldflags -X`; the version string is hardcoded near the top of that script and must be bumped by hand.

### Tests

`go build`, `go vet`, and `go test -short` are clean. Two gaps to be aware of:

- `TestCreate` in `golin_test.go` is a `t.Skip` placeholder. `getRoot` reads `os.Stdin` directly for its confirmation prompt, and the old stdin-injection API (`DefaultOption`/`SetOption`) is gone, so `Create` has no end-to-end test. Its pieces are covered instead by `golin_internal_test.go` (package `golin`): `copyDir` permissions, `readyPath` with an existing `tip`, `readyLink` + `restore`, `createLink`, `checkAuthorization`, and `parseGoVersion`.
- `TestReadyLinkExistsSymlink` needs a real `os.Symlink` and skips itself when that isn't permitted (an unelevated Windows prompt without Developer Mode).

## Architecture

Root package `golin` is the library; the `main` package lives in `_cmd/golin/main.go`. The underscore prefix is deliberate (see `_cmd/README.md`): it keeps the go tool from tripping over the v2 module path when installing the command with `go install github.com/shizuokago/golin/v2/_cmd/golin@latest`. Don't "fix" it by renaming.

Subcommand dispatch is a single switch in `_cmd/golin/main.go`; anything that isn't a known subcommand is treated as a version string and routed to `golin.Create(v)`.

- `create.go` — `Create(v)`: the main path. Resolve root → check write/symlink permission → print current `go version` → ensure the version dir exists (downloading if not) → delete + recreate the symlink (calling `restore` to put the old link back if that fails) → print the new `go version`.
- `install.go` — `Install(path, ver)`: bootstrap for machines with no Go at all. Downloads the archive straight from `https://golang.org/dl` and unpacks it, since it can't shell out to `go`.
- `golin.go` — the shared plumbing: `getRoot` (derives the parent dir from `GOROOT` and prompts on stdin if the last path element isn't the configured link name), `checkAuthorization` (probes by creating and removing a temp dir + symlink), `readyPath`, `readyLink`, `Download`, the `go env` / `go version` shell-outs (plus `parseGoVersion`, which also handles `devel` builds), and `CompressReleaseZip` for the `compress` subcommand.
- `development.go` — `CompileLatestSDK()` (the `dev` subcommand) and the `CompileSDK` / `TipVersion` constants. See "`dev` / `tip`" below.
- `version.go` — `Version` parsing and ordering (`v.r.m` plus a `mean` rank of beta < rc < release), plus `createVersionList()` which pulls `https://go.dev/dl/?mode=json&include=all`.
- `list.go` — grouping/printing of the version list by minor version; also cross-references versions already present on disk.
- `compress.go` — zip/tar.gz decompression for `install` (`DecompressURL`), and the zip writer (`Compress`) behind `CompressReleaseZip`.
- `config/` — process-global singleton (`config.Get()`) holding just `LinkName`, set once from the `-d` flag via functional options. The download URLs (`GolangDownloadPage`, `GoDevelopDownloadPage`, `GoGetLink`) are package constants here too.
- `golin_windows.go` / `golin_other.go` — home dir env var (`USERPROFILE` vs `HOME`) and download extension (`zip` vs `tar.gz`).
- `link_windows.go` / `link_other.go` — `createLink(target, link)`, the other OS-specific pair. See below.

### Link creation

Nothing outside `link_*.go` calls `os.Symlink`. `create.go`, `install.go`, `checkAuthorization`, and the `restore` closure in `readyLink` all go through `createLink`.

On Windows `createLink` builds an NTFS junction — `os.Mkdir` an empty dir, then `DeviceIoControl(FSCTL_SET_REPARSE_POINT)` with an `IO_REPARSE_TAG_MOUNT_POINT` `REPARSE_DATA_BUFFER` — using only stdlib `syscall`, no `x/sys` dependency. A junction needs no `SeCreateSymbolicLinkPrivilege`, so version switching works from a normal prompt. It can only target a directory on a local volume, so if the reparse point can't be set (UNC, network drive) it falls back to `os.Symlink`, which does need elevation.

Two consequences to keep in mind:

- Go reports a junction from `os.Lstat` as `ModeIrregular`, **not** `ModeSymlink`. `os.Readlink` and `os.Remove` work on it normally (remove deletes only the junction), which is why `readyLink` needs no OS-specific branch — but any new `mode&os.ModeSymlink` check would silently miss junctions.
- `restore` must use `createLink`, never `os.Symlink`. On an existing machine `current` is a real symlink created back when golin required Administrator. Deleting it needs no privilege, so an unelevated run gets as far as removing it; restoring with `os.Symlink` would then fail and leave `GOROOT` pointing at nothing. Replacing a symlink with a junction is a silent one-way migration, which is fine — the target is unchanged.

### Two different download mechanisms

Version switching (`Create`) does *not* download archives itself. It runs `go get`/`go install golang.org/dl/go{version}` — `go get` only when the currently-active version is known to be below 1.16; if `go version` couldn't be run or parsed, it assumes a modern toolchain and uses `go install ...@latest` — then executes the resulting `$GOPATH/bin/go{version}` binary with `download`. That lands the SDK in `{home}/sdk/go{version}`, which `readyPath` then copies to `{root}/{version}`. `getSDKPath` hardcodes that layout and breaks if `golang.org/dl` changes it.

The copy uses `copyDir`/`copyFile` rather than `os.Rename` because the SDK cache and the GOROOT root are often on different Windows drives, where rename fails.

### `dev` / `tip`

`golin dev` calls `Create(CompileSDK)`, where `CompileSDK == "compile_sdk"` is a sentinel, not a directory name. `readyPath` translates it to `TipVersion` (`"tip"`) and sets a `rebuild` flag: the existing `{root}/tip` is deleted so the SDK is always built from source again.

`golin tip` is not a subcommand — it falls through to `Create("tip")`, finds the existing `{root}/tip`, and just re-points the symlink. Both therefore share `{root}/tip`, which is what the Usage text describes.

## Conventions

- Comments and user-facing messages are largely Japanese; match the surrounding language when editing a file.
- In the library, wrapping an error uses `golang.org/x/xerrors` (`xerrors.Errorf("...: %w", err)`), not `fmt.Errorf`. Fresh errors with nothing to wrap are created with plain `fmt.Errorf` (e.g. in `getRoot`, `version.go`). `_cmd/golin/main.go` uses `fmt.Errorf` for everything.
- `go.mod` declares `go 1.12`, but that directive is already inaccurate: the code uses `os.ReadDir` (1.16), `errors.Is` (1.13), and `%w` wrapping (1.13), so it will not actually build with a 1.12 toolchain. Don't treat the directive as a real constraint — but don't casually raise it either, since it affects the language version applied to the whole module.
