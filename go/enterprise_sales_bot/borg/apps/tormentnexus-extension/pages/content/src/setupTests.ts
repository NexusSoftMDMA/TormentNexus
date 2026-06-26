import '@testing-library/jest-dom';

// Polyfill for matchMedia
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: () => { }, // Deprecated
        removeListener: () => { }, // Deprecated
        addEventListener: () => { },
        removeEventListener: () => { },
        dispatchEvent: () => false,
    }),
});

// Polyfill WebExtension API `chrome`
if (typeof chrome === 'undefined') {
    (global as any).chrome = {
        runtime: {
            getManifest: () => ({ version: '0.0.0' }),
            sendMessage: () => Promise.resolve(),
            onMessage: { addListener: () => { }, removeListener: () => { } }
        },
        storage: {
            local: {
                get: () => Promise.resolve({}),
                set: () => Promise.resolve(),
                remove: () => Promise.resolve()
            }
        }
    };
}
