import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    exclude: ['node_modules/**', 'dist/**', 'tests/e2e/**'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.ts'],
      exclude: ['src/cli/init.ts', 'src/cli/start.ts', 'src/server/index.ts', 'src/**/*.d.ts', 'src/**/types.ts'],
    },
  },
});
