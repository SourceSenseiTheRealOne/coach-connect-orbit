import assert from "node:assert/strict";
import { access, readFile } from "node:fs/promises";
import test from "node:test";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDirectory, "..");

const expectedPaths = [
  "package.json",
  "pnpm-workspace.yaml",
  "turbo.json",
  "apps/gateway/package.json",
  "apps/social/package.json",
  "apps/marketplace/package.json",
  "packages/ui/package.json",
  "packages/design-tokens/package.json",
  "packages/config/package.json",
  "packages/testing/package.json",
  "packages/go-api-client/package.json",
  "packages/trpc-contract/package.json",
  "packages/trpc-client/package.json",
];

test("frontend scaffold exposes the three zones and reusable boundary packages", async () => {
  await Promise.all(
    expectedPaths.map((relativePath) =>
      access(path.join(frontendRoot, relativePath)),
    ),
  );
});

test("workspace package manager is pinned to pnpm", async () => {
  const packageJson = JSON.parse(
    await readFile(path.join(frontendRoot, "package.json"), "utf8"),
  );

  assert.match(packageJson.packageManager, /^pnpm@\d+\.\d+\.\d+$/);
});
