import assert from "node:assert/strict";
import { access, readFile } from "node:fs/promises";
import test from "node:test";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDirectory, "..");
const projectRoot = path.resolve(frontendRoot, "..");

const readProjectFile = (relativePath) =>
  readFile(path.join(projectRoot, relativePath), "utf8");

const frontendApps = ["gateway", "social", "marketplace"];

test("Compose orchestrates every deployable service with health-aware wiring", async () => {
  const compose = await readProjectFile("compose.yaml");

  for (const service of ["gateway", "social", "marketplace", "api"]) {
    assert.match(compose, new RegExp(`^  ${service}:`, "m"));
  }

  assert.match(compose, /"3000:3000"/);
  assert.match(compose, /"3001:3001"/);
  assert.match(compose, /"3002:3002"/);
  assert.match(compose, /"9000:9000"/);
  assert.match(compose, /GO_API_URL:\s*http:\/\/api:9000/);
  assert.match(compose, /GATEWAY_ORIGIN:\s*http:\/\/gateway:3000/);
  assert.match(compose, /GATEWAY_PUBLIC_ORIGIN:/);
  assert.match(compose, /SOCIAL_ORIGIN:\s*http:\/\/social:3001/);
  assert.match(compose, /MARKETPLACE_ORIGIN:\s*http:\/\/marketplace:3002/);
  assert.match(compose, /condition:\s*service_healthy/);
  assert.equal((compose.match(/healthcheck:/g) ?? []).length, 4);
  assert.doesNotMatch(compose, /sk_(?:test|live)_/);
  assert.doesNotMatch(compose, /pk_(?:test|live)_[A-Za-z0-9]{12,}/);
});

test("each Next zone has a production standalone image contract", async () => {
  for (const app of frontendApps) {
    const [dockerfile, nextConfig] = await Promise.all([
      readProjectFile(`frontend/apps/${app}/Dockerfile`),
      readProjectFile(`frontend/apps/${app}/next.config.ts`),
      access(path.join(frontendRoot, ".dockerignore")),
    ]);

    assert.match(dockerfile, /FROM node:.* AS builder/);
    assert.match(dockerfile, /FROM node:.* AS runner/);
    assert.match(dockerfile, /pnpm --filter @coach-connect\//);
    assert.match(dockerfile, /USER nextjs/);
    assert.match(dockerfile, new RegExp(`apps/${app}/server\\.js`));
    assert.match(nextConfig, /output:\s*["']standalone["']/);
    assert.match(nextConfig, /outputFileTracingRoot:/);
  }
});

test("gateway image bakes Compose service DNS into its rewrite manifest", async () => {
  const [compose, dockerfile] = await Promise.all([
    readProjectFile("compose.yaml"),
    readProjectFile("frontend/apps/gateway/Dockerfile"),
  ]);

  assert.match(compose, /args:[\s\S]*SOCIAL_ORIGIN:\s*http:\/\/social:3001/);
  assert.match(
    compose,
    /args:[\s\S]*MARKETPLACE_ORIGIN:\s*http:\/\/marketplace:3002/,
  );
  assert.match(dockerfile, /ARG SOCIAL_ORIGIN/);
  assert.match(dockerfile, /ARG MARKETPLACE_ORIGIN/);
  assert.match(dockerfile, /\bSOCIAL_ORIGIN=\$SOCIAL_ORIGIN/);
  assert.match(dockerfile, /\bMARKETPLACE_ORIGIN=\$MARKETPLACE_ORIGIN/);
});

test("Go API image is production-mode, non-root, and externally reachable", async () => {
  const [dockerfile, appConfig] = await Promise.all([
    readProjectFile("backend/Dockerfile"),
    readProjectFile("backend/conf/app.conf"),
    access(path.join(projectRoot, "backend/.dockerignore")),
  ]);

  assert.match(dockerfile, /FROM golang:.* AS builder/);
  assert.match(dockerfile, /WORKDIR \/workspace\/backend/);
  assert.match(dockerfile, /USER app/);
  assert.match(dockerfile, /revel.*prod/);
  assert.match(dockerfile, /-t=\.\/dist/);
  assert.match(dockerfile, /\/workspace\/backend\/dist\//);
  assert.match(appConfig, /\[prod\][\s\S]*http\.addr\s*=\s*0\.0\.0\.0/);
});

test("Compose configuration accepts runtime secrets without committing values", async () => {
  const [compose, envExample, gitignore] = await Promise.all([
    readProjectFile("compose.yaml"),
    readProjectFile(".env.example"),
    readProjectFile(".gitignore"),
  ]);

  assert.match(
    compose,
    /NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY:\s*\$\{NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY:-\}/,
  );
  assert.match(compose, /CLERK_SECRET_KEY:\s*\$\{CLERK_SECRET_KEY:-\}/);
  assert.match(compose, /NEXT_PUBLIC_CLERK_SIGN_IN_FALLBACK_REDIRECT_URL:\s*\/dashboard/);
  assert.doesNotMatch(compose, /CLERK_SIGN_(?:IN|UP)_FORCE_REDIRECT_URL/);
  assert.equal((compose.match(/^\s+CLERK_SECRET_KEY:/gm) ?? []).length, 3);
  assert.match(envExample, /GATEWAY_PUBLIC_ORIGIN=http:\/\/localhost:3000/);
  assert.match(
    compose,
    /DATABASE_URL:\s*\$\{DATABASE_URL:-postgresql:\/\/postgres:postgres@host\.docker\.internal:54322\/postgres\?sslmode=disable\}/,
  );
  assert.match(compose, /APP_SECRET:\s*\$\{APP_SECRET/);
  assert.match(envExample, /NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=/);
  assert.match(envExample, /CLERK_SECRET_KEY=/);
  assert.match(envExample, /DATABASE_URL=/);
  assert.match(envExample, /APP_SECRET=/);
  assert.match(gitignore, /^\.env$/m);
});
