import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['tests/**/*.test.ts'],
    // tests/tmux.test.ts is a hand-rolled stand-alone script that predates
    // the vitest harness — it self-execs and calls process.exit. Excluded
    // here so vitest doesn't trip on the bare-script style.
    exclude: ['tests/tmux.test.ts', 'node_modules/**', 'dist/**', '.next/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
