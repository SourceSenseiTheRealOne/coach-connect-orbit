import assert from "node:assert/strict";
import { access, readFile, readdir } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const frontendRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const projectRoot = path.resolve(frontendRoot, "..");
const gatewayRoot = path.join(frontendRoot, "apps", "gateway");

async function read(relativePath) {
  return readFile(path.join(projectRoot, relativePath), "utf8");
}

async function collectFiles(root, extension, files = []) {
  for (const entry of await readdir(root, { withFileTypes: true })) {
    if (
      ["node_modules", ".next", ".turbo", ".clerk", "coverage"].includes(
        entry.name,
      )
    )
      continue;
    const entryPath = path.join(root, entry.name);
    if (entry.isDirectory()) await collectFiles(entryPath, extension, files);
    else if (entry.name.endsWith(extension)) files.push(entryPath);
  }
  return files;
}

test("dashboard scaffold has Clerk protection and an authenticated shell", async () => {
  const expectedFiles = [
    path.join(gatewayRoot, "src/proxy.ts"),
    path.join(gatewayRoot, "src/app/dashboard/page.tsx"),
    path.join(gatewayRoot, "src/app/dashboard/layout.tsx"),
    path.join(frontendRoot, "packages/ui/src/theme-provider.tsx"),
    path.join(frontendRoot, "packages/ui/src/theme-toggle.tsx"),
  ];

  await Promise.all(expectedFiles.map((file) => access(file)));

  const proxy = await read("frontend/apps/gateway/src/proxy.ts");
  assert.match(proxy, /createRouteMatcher/);
  assert.match(proxy, /dashboard\(\.\*\)/);
  assert.match(proxy, /auth\.protect\(\)/);
  assert.match(proxy, /CLERK_SECRET_KEY/);
  assert.match(proxy, /NextResponse\.redirect/);

  const authProvider = await read("frontend/packages/auth/src/index.tsx");
  assert.match(authProvider, /publishableKey/);
  assert.doesNotMatch(authProvider, /dynamic/);

  const gatewayLayout = await read("frontend/apps/gateway/src/app/layout.tsx");
  assert.match(gatewayLayout, /NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY/);

  const frontendGitignore = await read("frontend/.gitignore");
  assert.match(frontendGitignore, /^\.clerk\/$/m);

  const dashboardPage = await read(
    "frontend/apps/gateway/src/app/dashboard/page.tsx",
  );
  assert.match(dashboardPage, /await auth\(\)/);
  assert.match(dashboardPage, /redirect\(/);
  assert.match(dashboardPage, /dynamic\s*=\s*["']force-dynamic["']/);

  const dashboardLayout = await read(
    "frontend/apps/gateway/src/app/dashboard/layout.tsx",
  );
  assert.match(dashboardLayout, /<aside/);
  assert.match(dashboardLayout, /<header/);
  assert.match(dashboardLayout, /ThemeToggle/);
  assert.match(dashboardLayout, /UserButton/);

  const signInPage = await read(
    "frontend/apps/gateway/src/app/sign-in/[[...sign-in]]/page.tsx",
  );
  assert.match(signInPage, /NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY/);
  assert.match(signInPage, /Clerk configuration required/);
});

test("app-authored styling uses Tailwind utilities only", async () => {
  const sourceFiles = await collectFiles(frontendRoot, ".tsx");
  for (const file of sourceFiles) {
    const source = await readFile(file, "utf8");
    assert.doesNotMatch(source, /\bstyle\s*=/, `${file} uses inline styles`);
    assert.doesNotMatch(source, /var\(--/, `${file} uses raw CSS variables`);
  }

  const cssFiles = await collectFiles(frontendRoot, ".css");
  assert.equal(
    cssFiles.length,
    3,
    `expected only the three Tailwind entry CSS files, found ${cssFiles.length}`,
  );
  for (const file of cssFiles) {
    const lines = (await readFile(file, "utf8"))
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean);
    for (const line of lines) {
      assert.match(
        line,
        /^@(import|source|custom-variant)\b/,
        `${file} contains authored CSS: ${line}`,
      );
    }
  }
});

test("repository records GitHub workflow and code intelligence policy", async () => {
  const contributionPolicy = await read("CONTRIBUTING.md");
  assert.match(contributionPolicy, /coach-connect-orbit/i);
  assert.match(contributionPolicy, /pull request|\bPR\b/i);
  assert.match(contributionPolicy, /CodeGraph/);
  assert.match(
    contributionPolicy,
    /(never[^\n]*GitNexus|GitNexus[^\n]*never)/i,
  );

  const projectGitignore = await read(".gitignore");
  assert.match(projectGitignore, /^\.codegraph\/$/m);
  assert.match(projectGitignore, /^\.gitnexus\/$/m);
  assert.match(projectGitignore, /^context\/$/m);
  assert.match(projectGitignore, /^AGENTS\.MD$/m);
});
