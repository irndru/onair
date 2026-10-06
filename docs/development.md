# Development

Needs Go 1.27 and `make`. No cgo.

```sh
make build    # ./onair
make test
make lint     # gofmt, go vet for macOS and Linux
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

All system calls sit behind the `bridge` interface in
`internal/controlcenter/bridge.go`. Tests use `fakeBridge` (`useFake`) and must
never call a setter on the real one, so `make test` never changes your
settings.

`onair -v <command>` or `ONAIR_DEBUG=1` logs to stderr.

To check a real change, open Photo Booth, run `./onair status "photo booth"`,
change something, confirm it in Control Center > Video Effects, then restore
what `status` showed.
