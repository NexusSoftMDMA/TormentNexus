import { defineConfig } from 'vite';

// Stub build — the browser extension source lives in apps/tormentnexus-extension
// This package only holds shared dependencies (@mozilla/readability, turndown).
export default defineConfig({
  build: {
    // Skip the default index.html entry; we have no frontend bundle here.
    lib: {
      entry: 'stub.ts',
      formats: ['es'],
    },
  },
});
