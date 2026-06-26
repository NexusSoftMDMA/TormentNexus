
import { describe, it, expect } from 'vitest';

/**
 * This test is INTENTIONALLY BROKEN to verify Auto-Dev "Fix Loop".
 * The Agent should modify the expected value to match the actual value.
 */
describe('AutoDev Verification', () => {
    it('should be fixed by the agent', () => {
        const actual = "fixed";
        const expected = "fixed";
        expect(actual).toBe(expected);
    });
});
