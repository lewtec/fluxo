# GraphQL package

GraphQL API for Fluxo: schema, gqlgen-generated executors, Rain mappers, and resolvers.

## Layout

| File | Role |
|------|------|
| `schema.graphql` | Queries, mutations, subscriptions, and types |
| `schema.resolvers.go` | Resolver implementations (kept across regenerate) |
| `generated.go` | gqlgen executable schema |
| `models_gen.go` | gqlgen model types |
| `mappers.go` | Rain torrent types → GraphQL models |
| `resolver.go` | Root resolver (`session.Manager`) |
| `send_sub_test.go` | Tests for subscription drain helpers |

## Regenerate

Needs network the first time tools are fetched:

```bash
go tool github.com/99designs/gqlgen generate
```

Config: `gqlgen.yml` at the repo root. Prefer regenerating after schema changes; do not hand-edit `generated.go` or `models_gen.go`.

## Errors

Session-level sentinels live in `internal/session/errors.go` (typed `sessionError` table). Resolvers wrap with `%w` so callers can use `errors.Is`.

Current sentinels include:

- `ErrTorrentNotFound`
- `ErrInvalidURI`
- `ErrNoLocalIP`
- `ErrUPNPDiscoveryTimeout`
- `ErrNoUPNPClients`
- `ErrUPNPMappingFailed`

## Subscriptions

`subscribeFilter` / `sendSub` drain the session `EventBus` into buffered channels. When a client is slow, oldest buffered values are dropped so the bus itself does not back up (see `subOutBuffer` and tests in `send_sub_test.go`).

## Rain mapping notes

Mappers target Rain 1.13-style stats:

- Nested `Stats.Bytes.*` and `Stats.Speed.*`
- `t.AddedAt()` method
- `FileStats` / `Tracker` / `Webseed` slices from torrent methods

Some GraphQL fields are intentionally zeroed when Rain does not expose the data (e.g. `downloadTime`, file `priority`).
