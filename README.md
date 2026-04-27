# go-activesync

[![CI](https://github.com/remdev/go-activesync/actions/workflows/ci.yml/badge.svg)](https://github.com/remdev/go-activesync/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/remdev/go-activesync.svg)](https://pkg.go.dev/github.com/remdev/go-activesync)
[![Go Report Card](https://goreportcard.com/badge/github.com/remdev/go-activesync)](https://goreportcard.com/report/github.com/remdev/go-activesync)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8?logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A pure-Go client library for the **Microsoft Exchange ActiveSync (EAS) protocol**, version 14.1.

The library is built spec-first (TDD): every requirement of the underlying
Microsoft Open Specifications is backed by a concrete test before any
implementation lands. See [`docs/spec-coverage.md`](docs/spec-coverage.md) for
the traceability matrix.

---

## Quick examples

### Discover the EAS endpoint and provision a device

```go
import (
    "context"
    "net/http"

    "github.com/remdev/go-activesync/autodiscover"
    "github.com/remdev/go-activesync/client"
)

ctx := context.Background()

ad, err := autodiscover.New(http.DefaultClient).Discover(ctx, "user@example.com",
    &autodiscover.Credentials{Username: "user@example.com", Password: "pass"})
if err != nil { /* handle */ }

c, _ := client.New(client.Config{
    BaseURL:    ad.URL,
    Auth:       &client.BasicAuth{Username: "user@example.com", Password: "pass"},
    DeviceID:   "stable-device-id",
    DeviceType: "SmartPhone",
    UserAgent:  "my-app/1.0",
})

if _, err := c.Provision(ctx, "user@example.com"); err != nil { /* handle */ }
```

### Sync new e-mails from the inbox

```go
import "github.com/remdev/go-activesync/eas"

initial, _ := c.Sync(ctx, user, &eas.SyncRequest{
    Collections: eas.SyncCollections{
        Collection: []eas.SyncCollection{{SyncKey: "0", CollectionID: inboxID}},
    },
})
syncKey := initial.Collections.Collection[0].SyncKey

resp, _ := c.Sync(ctx, user, &eas.SyncRequest{
    Collections: eas.SyncCollections{
        Collection: []eas.SyncCollection{{
            SyncKey:      syncKey,
            CollectionID: inboxID,
            GetChanges:   1,
            WindowSize:   25,
        }},
    },
})

for _, col := range resp.Collections.Collection {
    for _, add := range col.Commands.Add {
        // add.ApplicationData → typed eas.Email
    }
}
```

### Long-poll for changes with Ping

```go
resp, _ := c.Ping(ctx, user, &eas.PingRequest{
    HeartbeatInterval: 480,
    Folders: eas.PingFolders{
        Folder: []eas.PingFolder{{ID: inboxID, Class: "Email"}},
    },
})
if eas.PingHasChanges(resp.Status) {
    // pull the changed folders with Sync
}
```

Runnable end-to-end programs live under [`examples/`](examples/):
[`login`](examples/login), [`inbox-sync`](examples/inbox-sync),
[`calendar-sync`](examples/calendar-sync), [`ping`](examples/ping).

---

## Install

```sh
go get github.com/remdev/go-activesync@latest
```

Requires Go **1.26** or newer.

---

## Status (v0.x)

Implemented and covered by the test suite:

| Area            | Detail                                                                    |
| --------------- | ------------------------------------------------------------------------- |
| Transport       | MS-ASHTTP — base64-encoded query, plain query fallback, mandatory headers |
| Codec           | MS-ASWBXML — WBXML 1.3 encoder/decoder, all 25 EAS 14.1 code pages        |
| Reflection      | `wbxml.Marshal` / `wbxml.Unmarshal` driven by `wbxml:"Page.Tag"` tags     |
| Autodiscover    | MS-OXDISCO + MS-ASAB POX (`mobilesync` schema, SRV fallback, redirects)   |
| Auth            | HTTP Basic; pluggable `Authenticator` interface                           |
| Provisioning    | Two-pass MS-ASPROV with auto re-provision on Status 142/143               |
| Commands        | `Provision`, `FolderSync`, `Sync`, `Ping`                                 |
| PIM data models | `MS-ASEMAIL`, `MS-ASCAL`, `MS-ASCNTC`, `MS-ASTASK`                        |
| Stores          | In-memory `PolicyStore` and `SyncStateStore`; pluggable interfaces        |
| Hardening       | Bounded decoder allocations + `FuzzDecode` over the WBXML reader          |

---

## Roadmap

Out of scope for v0.x; tracked for future releases.

- **Commands**: `SendMail`, `SmartReply`, `SmartForward`, `MeetingResponse`,
  `MoveItems`, `ItemOperations` (Fetch/EmptyFolderContents),
  `GetItemEstimate`, `Search`, `ResolveRecipients`, `ValidateCert`,
  `Settings`, `ResolveRecipients`, `Find`.
- **Protocol versions**: negotiation of EAS 12.1, 14.0, 16.0, 16.1 in
  addition to the current hard-coded 14.1.
- **Code pages**: per-version code-page selection (the current set is
  pinned to 14.1).
- **Auth**: OAuth 2.0 bearer (Microsoft 365 / EWS-style),
  client-certificate / mutual-TLS authenticator, NTLM/Negotiate.
- **Body**: `MIME` body type round-tripping, `BodyPartPreference` + Rights
  Management (MS-ASRM).
- **Search & Document Library**: GAL `Search`, `MS-ASDOC` document fetch.
- **Notes class**: typed `MS-ASNOTE` model.
- **Persistence**: SQLite/Bolt-backed `PolicyStore` and `SyncStateStore`
  alongside the in-memory implementations.
- **Server side**: there is no server skeleton; this is purely a client
  library.
- **Observability**: structured logging hooks, OpenTelemetry spans on
  command boundaries.

---

## Repository layout

```
wbxml/             WBXML 1.3 codec + EAS code page tables, fuzz harness
eas/               typed request/response/domain models (one file per spec)
autodiscover/      POX Autodiscover client
client/            high-level EAS client (transport, auth, command methods, stores)
examples/          runnable demos (login, inbox-sync, calendar-sync, ping)
internal/spec/     traceability-matrix linter + coverage gate tool
docs/              spec-coverage.md and other design notes
```

---

## Development

```sh
make test       # go test -race ./...
make vet        # go vet ./...
make lint       # golangci-lint run ./...   (auto-installs golangci-lint if absent)
make lint-fix   # golangci-lint run --fix
make spec-lint  # verify the traceability matrix is fully covered
make cover      # go test -race -coverprofile=cover.out
make cover-gate # enforce per-package coverage thresholds
make fuzz       # short FuzzDecode smoke run
make all        # vet + lint + test + cover-gate
make ci         # run the exact CI pipeline locally (mod verify, vet, lint,
                # race tests, cover-gate, spec-lint, fuzz smoke)
```

Always run `make ci` before pushing or opening a PR — it mirrors
`.github/workflows/ci.yml` step-for-step. See [`AGENTS.md`](AGENTS.md) for
the full contributor checklist.

The configured linter set (see [`.golangci.yml`](.golangci.yml)) bundles
`staticcheck`, `govet`, `errcheck`, `revive`, `gosec`, `gocritic`,
`bodyclose`, `errorlint`, `unparam`, `unconvert`, `usestdlibvars`,
`usetesting`, formatters `gofmt` and `goimports`, and a handful of others.
Test files relax the noisier rules; see the `exclusions` block for the
exact list.

CI enforces per-package coverage thresholds (`covergate`):

| Package         | Threshold |
| --------------- | --------- |
| `wbxml/`        | 90%       |
| `eas/`          | 90%       |
| `client/`       | 80%       |
| `autodiscover/` | 80%       |

---

## License

MIT. See [LICENSE](LICENSE).
