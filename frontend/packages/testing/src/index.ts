import { defineConfig, type ViteUserConfig } from "vitest/config";

export function createUnitTestConfig(
  overrides: ViteUserConfig = {},
): ViteUserConfig {
  return defineConfig({
    test: {
      clearMocks: true,
      passWithNoTests: false,
      restoreMocks: true,
    },
    ...overrides,
  });
}
