# Coach Connect MVP v2

Specification-driven starter scaffold for a football social network and marketplace.

## Structure

- `frontend/` — pnpm/Turborepo, three Next.js Multi-Zone apps, shared UI/auth/tRPC/API packages
- `backend/` — Go/Revel API, OpenAPI contract, Ent schema boundary, local Supabase project
- `context/` — product, architecture, UI, code, workflow, and progress specifications

## Start locally

Terminal 1:

```bash
cd backend
go test ./...
go tool revel run -a=. dev
```

Terminal 2:

```bash
cd frontend
pnpm install
pnpm dev
```

Then open `http://127.0.0.1:3000`.

The starter liveness path is:

```text
React tRPC hook -> Next.js /api/trpc -> Go API client -> Revel /api/v1/health -> application service
```

Local Supabase is initialized under `backend/supabase`; start it only when a database-backed slice is being developed.
