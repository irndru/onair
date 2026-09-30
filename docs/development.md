# Development

## Toolchain

[devbox](https://www.jetify.com/devbox) pins Go 1.27 and sets `CGO_ENABLED=0`.
With [direnv](https://direnv.net), `direnv allow` loads the environment when
you enter the directory. Otherwise use `devbox shell` or the scripts:

| Script | Does |
| --- | --- |
| `devbox run build` | Build `./videobg` |
| `devbox run install` | Install `videobg` into `$GOBIN` |
| `devbox run test` | `go test ./...` |
| `devbox run lint` | `gofmt` check, then `go vet` for macOS and Linux |

Plain Go 1.27 works too. Set `CGO_ENABLED=0`.

## Layout

```
cmd/videobg/          the command line: argument handling and output
internal/videofx/
  videofx.go          Current, SetImage, SetEnabled
  apps.go             app discovery and name resolution
  builtin.go          built-in gradients and image resolution
  objc_darwin.go      the bridge to AVFoundation and Foundation
  objc_stub.go        stubs so the module builds and vets off macOS
docs/
```

The only dependency is [purego](https://github.com/ebitengine/purego). There
is no cgo.

## Testing

`go test ./...` never changes your camera settings. No test calls the set
functions.

- `cmd/videobg` tests argument handling, app selection and output.
- `internal/videofx` tests name and image resolution on every platform.
- `objc_darwin_test.go` runs on macOS only. `TestLoad` binds every private
  symbol, so it fails when Apple renames one. The rest test string conversion
  and app discovery against fixture `.app` bundles in a temp directory.

To check a real change, use an app you are not in a call with:

```sh
devbox run build
./videobg status "photo booth"
./videobg set purple "photo booth"
```

Open Photo Booth, then Control Center > Video Effects, and confirm the image.
Then restore what `status` showed first.

Before a commit:

```sh
devbox run lint
devbox run test
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

## CI

`.github/workflows/ci.yml` runs on macOS and Ubuntu: `gofmt` check, `go vet`,
`go test`, `go build`. The Go version comes from `go.mod`. The Ubuntu job
keeps the non-macOS stubs compiling.
