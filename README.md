# Coach Connect MVP v2

Specification-driven starter scaffold for a football social network and marketplace.

## Structure

- `frontend/` — pnpm/Turborepo, three Next.js Multi-Zone apps, shared UI/auth/tRPC/API packages
- `backend/` — Go/Revel API, OpenAPI contract, Ent schema boundary, local Supabase project
- `context/` — product, architecture, UI, code, workflow, and progress specifications

## Start the complete stack with Docker

Create a local environment file and replace `APP_SECRET` with a high-entropy value:

```bash
cp .env.example .env
docker compose up --build -d
docker compose ps
```

The Go API emits structured JSON request logs to stdout. Inspect them with `docker compose logs --follow api`; each completion record includes a generated `X-Request-ID` correlation value and excludes headers, bodies, and query strings.

Open the gateway at `http://127.0.0.1:3000`. The individual services are also exposed for diagnostics:

- Social zone: `http://127.0.0.1:3001/feed`
- Marketplace zone: `http://127.0.0.1:3002/marketplace`
- Go API health: `http://127.0.0.1:9000/api/v1/health`

Without Clerk development keys, `/sign-in` and `/sign-up` show the intentional configuration screen and `/dashboard` redirects to sign-in. Add `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` and `CLERK_SECRET_KEY` to the ignored root `.env`, then recreate the frontend services to activate Clerk.

```bash
docker compose up -d --force-recreate gateway social marketplace
docker compose down
```

Compose starts the three independently deployable Next.js standalone images and the compiled Go/Revel API with health-aware dependencies. Supabase is intentionally not part of the default graph until an approved database-backed slice requires it.

## Start directly on the host

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
