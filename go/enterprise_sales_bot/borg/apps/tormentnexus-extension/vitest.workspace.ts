import { defineWorkspace } from 'vitest/config';

export default defineWorkspace([
    {
        test: {
            name: 'extension-content-script',
            include: ['pages/content/src/**/*.test.{ts,tsx}'],
            environment: 'jsdom',
            setupFiles: ['pages/content/src/setupTests.ts'],
            alias: {
                '@src': new URL('./pages/content/src', import.meta.url).pathname,
            },
        },
    },
]);
