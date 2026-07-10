# Coach Connect Frontend

A pnpm/Turborepo workspace containing three Next.js 16 Multi-Zone applications and reusable boundary packages.

## Local development

```bash
cp .env.example .env.local
pnpm install
pnpm dev
```

- Gateway: `http://127.0.0.1:3000`
- Social zone: `http://127.0.0.1:3001`
- Marketplace zone: `http://127.0.0.1:3002`

The gateway rewrites `/feed` and related social paths to the social zone, and `/marketplace` paths to the marketplace zone. Browser API traffic uses `/api/trpc`; tRPC procedures call the Go API wrapper and never the database.

## Container runtime

Each zone has its own multi-stage, non-root production image and emits a Next.js standalone artifact. Build and start all zones from the repository root with `docker compose up --build -d`; service-to-service origins are injected by Compose rather than hard-coded into the images.

## Quality gates

```bash
pnpm test
pnpm typecheck
pnpm lint
pnpm build
```

## API type generation

```bash
pnpm generate:api
```

This regenerates transport types from `../backend/contracts/openapi.yaml`. Do not hand-edit generated files.
