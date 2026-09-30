# Development

## Toolchain

You need Go 1.27 on your `PATH` and `make`. The Makefile sets
`CGO_ENABLED=0`.

| Target | Does |
| --- | --- |
| `make build` | Build `./videobg` |
| `make install` | Install `videobg` into `$GOBIN` |
| `make test` | `go test ./...` |
| `make lint` | `gofmt` check, then `go vet` for macOS and Linux |

## Layout

```
cmd/videobg/          the command line: argument handling and output
internal/videofx/
  videofx.go          Current, SetImage, SetEnabled, SetMic
  effects.go          effect and mic mode names
  apps.go             app discovery and name resolution
  builtin.go          built-in gradients and image resolution
  bridge.go           the bridge interface over everything system-side
  objc_darwin.go      the bridge to AVFoundation and Foundation
  objc_stub.go        a bridge that fails, so the module builds off macOS
  fake_test.go        a fake bridge that records set calls
docs/
```

The only dependency is [purego](https://github.com/ebitengine/purego). There
is no cgo.

## Testing

`go test ./...` never changes your camera settings. No test calls the set
functions.

- `cmd/videobg` tests argument handling, app selection and output.
- `internal/videofx` runs its logic against `fakeBridge` on every platform:
  the all-or-nothing checks, the order of set calls, reading state back, and
  app discovery. It also tests name and image resolution.
- `objc_darwin_test.go` runs on macOS only. `TestLoad` binds every private
  symbol. It fails when Apple renames one the background needs and logs any
  other that is missing. The rest test string conversion and `Info.plist`
  reading against fixture `.app` bundles in a temp directory.

A new test that changes settings must use `useFake`. Never call a setter on
the real bridge.

## Debugging

`videobg -v <command>`, or `VIDEOBG_DEBUG=1`, logs to stderr: the macOS
version, each symbol bound or missing, the apps found and why, any check that
stopped a change, every set call, and the raw values read back. stdout does
not change. Add `slog.Debug` calls for anything new that touches the system.

To check a real change, use an app you are not in a call with:

```sh
make build
./videobg status "photo booth"
./videobg set purple "photo booth"
./videobg portrait on "photo booth"
./videobg mic isolation "photo booth"
```

Open Photo Booth, then Control Center > Video Effects, and confirm each
change. Then restore what `status` showed first.

Before a commit:

```sh
make lint
make test
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Adding a known app

Apps that declare `NSCameraUsageDescription` are found automatically. Add an
app to `knownApps` in `internal/videofx/apps.go` when either:

- it uses the camera without declaring that key (Chrome does this), or
- it should be a default target before the user has toggled its background.

Get the bundle identifier with `osascript -e 'id of app "Name"'`. Update the
list in `docs/internals.md` to match.
