import { describe, expect, it } from 'vitest';

import { splitTerminalSensorBuffer } from './TerminalSensor.js';

describe('TerminalSensor', () => {
    it('buffers partial stderr chunks into a single complete line', () => {
        const first = splitTerminalSensorBuffer('[MCPServer] Failed to capture tool observation: Error: Failed to infer ');

        expect(first.lines).toEqual([]);
        expect(first.remaining).toBe('[MCPServer] Failed to capture tool observation: Error: Failed to infer ');

        const second = splitTerminalSensorBuffer(`${first.remaining}data type for field structuredObservation.filesRead at row 0.\n`);

        expect(second.lines).toEqual([
            '[MCPServer] Failed to capture tool observation: Error: Failed to infer data type for field structuredObservation.filesRead at row 0.',
        ]);
        expect(second.remaining).toBe('');
    });
});