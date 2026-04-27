# go-activesync

A pure-Go client library for the Microsoft Exchange ActiveSync (EAS) protocol,
version 14.1.

## Status

Early development. The library is being built spec-first (TDD), with each
requirement of the underlying Microsoft Open Specifications backed by a
concrete test before any implementation lands.

## Scope (v0.x)

- Transport: MS-ASHTTP (base64-encoded query, plain query fallback).
- Codec: MS-ASWBXML (WBXML 1.3 with the 25 EAS code pages of revision 14.1).
- Autodiscover: MS-OXDISCO / MS-ASAB POX schema (`mobilesync`).
- Commands: Provision, FolderSync, Sync, Ping.
- PIM data classes: Email (MS-ASEMAIL), Calendar (MS-ASCAL),
  Contacts (MS-ASCNTC), Tasks (MS-ASTASK).

Out of scope for v0.x: server side, EAS versions other than 14.1, command set
beyond the four above (SendMail, ItemOperations, Search, MoveItems, etc.).

## Layout

- `wbxml/` — WBXML codec and EAS code page tables.
- `eas/` — typed request/response and domain models.
- `autodiscover/` — POX Autodiscover client.
- `client/` — high-level EAS client (transport, auth, command methods).
- `internal/spec/` — spec-coverage matrix linter.
- `docs/spec-coverage.md` + `internal/spec/coverage.csv` — traceability matrix.
- `examples/` — runnable examples.

## Development

```sh
make test       # go test -race ./...
make vet        # go vet ./...
make lint       # golangci-lint run ./...   (auto-installs golangci-lint if absent)
make lint-fix   # golangci-lint run --fix
make spec-lint  # verify the traceability matrix is fully covered
make cover      # go test -race -coverprofile
make cover-gate # enforce per-package coverage thresholds
make fuzz       # short fuzz smoke run
```

The configured linter set (see `.golangci.yml`) bundles `staticcheck`,
`govet`, `errcheck`, `revive`, `gosec`, `gocritic`, `bodyclose`, `errorlint`,
`unparam`, `unconvert`, `usestdlibvars`, `usetesting`, formatters `gofmt`
and `goimports`, and a handful of others. Test files relax the noisier
rules; see the `exclusions` block for the exact list.

## License

MIT. See [LICENSE](LICENSE).
