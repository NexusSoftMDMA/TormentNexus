import { describe, expect, it, vi } from 'vitest';

import { HealerService } from './HealerService.js';
import { shouldIgnoreExpectedStartupError } from '../reactors/HealerReactor.js';

describe('Healer degraded startup handling', () => {
    it('ignores expected SQLite-unavailable startup errors', () => {
        expect(shouldIgnoreExpectedStartupError(
            'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better_sqlite3.node)',
        )).toBe(true);
    });

    it('parses fenced JSON diagnoses without logging a parse failure', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        const generateText = vi.fn(async () => ({
            content: [
                '```json',
                '{',
                '  "errorType": "RuntimeError",',
                '  "description": "SQLite runtime is unavailable.",',
                '  "file": "packages/core/src/db/index.ts",',
                '  "line": 724,',
                '  "suggestedFix": "Install or rebuild better-sqlite3.",',
                '  "confidence": 0.92',
                '}',
                '```',
            ].join('\n'),
        }));
        const service = new HealerService(
            {
                generateText,
            } as any,
            {} as any,
        );

        const diagnosis = await service.analyzeError('SQLite runtime is unavailable');

        expect(diagnosis).toMatchObject({
            errorType: 'RuntimeError',
            description: 'SQLite runtime is unavailable.',
            file: 'packages/core/src/db/index.ts',
            line: 724,
            suggestedFix: 'Install or rebuild better-sqlite3.',
            confidence: 0.92,
        });
        expect(consoleErrorSpy).not.toHaveBeenCalledWith(
            'Failed to parse Healer diagnosis',
            expect.anything(),
        );
        expect(generateText).toHaveBeenCalledWith(
            'openrouter',
            expect.any(String),
            'You are a JSON-only debugging tool.',
            expect.any(String),
            {},
        );
    });
});
