# Contributing

Open an issue before a large change. Read [ARCHITECTURE.md](ARCHITECTURE.md)
first.

## Development

Needs Go 1.27, `make` and
[golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2. No cgo.

```sh
make build    # ./onair
make test
make lint     # golangci-lint for macOS and Linux
make fmt
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

`onair -v <command>` or `ONAIR_DEBUG=1` logs to stderr.

To check a real change, open Photo Booth, run `./onair status "photo booth"`,
change something, confirm it in Control Center > Video Effects, then restore
what `status` showed.

## Releasing

Versions follow [semantic versioning](https://semver.org). The public API is
the command line: commands, arguments, output columns, exit codes and
`ONAIR_DEBUG`. Nothing under `internal/` is API.

- Patch: fixes.
- Minor: new commands, arguments or columns.
- Major: anything removed, renamed or changed in meaning, output or exit code.
  From v2 the module path needs a `/v2` suffix.

Before v1.0.0 a minor version may break things.

Tag a release on `main` with an annotated tag and push it:

```sh
git tag -a v0.2.0 -m v0.2.0
git push origin v0.2.0
```

Never move or delete a pushed tag: the Go module proxy and checksum database
keep it forever. To withdraw a bad release, add a `retract` directive to
`go.mod` and tag a new patch.
