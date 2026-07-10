# Coach Connect Go API

API-only Revel service for Coach Connect MVP v2.

## Commands

```bash
go test ./...
go tool revel run -a=. dev
```

The local API listens on `http://127.0.0.1:9000`. The initial liveness route is `GET /api/v1/health`.

Application code follows the boundary:

```text
Revel controller -> application service -> repository port -> infrastructure adapter
```

Revel is pinned through the Go `tool` directive, together with a Go 1.26-compatible `golang.org/x/tools` version. The local server binds `127.0.0.1` deliberately to avoid the Windows `localhost` IPv4/IPv6 proxy mismatch.

Ent 0.14.6 is also pinned as a Go tool. After the first approved entity schema is added, generate the client with:

```bash
go tool ent generate ./ent/schema
```

Local Supabase is initialized under `supabase/`. Ent schemas and migrations are added feature-by-feature; runtime auto-migration is not allowed.
