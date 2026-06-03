import { describe, expect, it } from 'vitest';

import {
    buildAutomaticToolContextFingerprint,
    buildAutomaticToolContextMemory,
    buildAutomaticToolContextPreview,
    buildAutomaticToolContextStartPayload,
    shouldPersistAutomaticToolContext,
    shouldResolveAutomaticToolContext,
} from './toolContextInjection.js';

describe('toolContextInjection', () => {
    it('skips TormentNexus meta-tools for automatic JIT resolution', () => {
        expect(shouldResolveAutomaticToolContext('get_tool_context')).toBe(false);
        expect(shouldResolveAutomaticToolContext('search_tools')).toBe(false);
        expect(shouldResolveAutomaticToolContext('read_file')).toBe(true);
    });

    it('persists only relevant tool context payloads', () => {
        expect(shouldPersistAutomaticToolContext(null)).toBe(false);
        expect(shouldPersistAutomaticToolContext({
            toolName: 'read_file',
            query: 'read_file page.tsx',
            matchedPaths: [],
            observationCount: 0,
            summaryCount: 0,
            prompt: 'No relevant prior memory.',
        })).toBe(false);

        expect(shouldPersistAutomaticToolContext({
            toolName: 'read_file',
            query: 'read_file page.tsx',
            matchedPaths: ['src/app/page.tsx'],
            observationCount: 1,
            summaryCount: 0,
            prompt: 'JIT tool context for read_file:\n- Prior note',
        })).toBe(true);
    });

    it('builds a compact preview and session-memory payload', () => {
        const payload = {
            toolName: 'read_file',
            query: 'read_file page.tsx',
            matchedPaths: ['src/app/page.tsx'],
            observationCount: 2,
            summaryCount: 1,
            prompt: 'JIT tool context for read_file:\n- Prior note',
        };

        expect(buildAutomaticToolContextPreview(payload)).toContain('2 observations');
        expect(buildAutomaticToolContextPreview(payload)).toContain('1 summary');

        const memory = buildAutomaticToolContextMemory(payload);
        expect(memory.content).toContain('Prior note');
        expect(memory.metadata.source).toBe('pre_tool_context');
        expect(memory.metadata.toolName).toBe('read_file');

        const startPayload = buildAutomaticToolContextStartPayload(payload);
        expect(startPayload.contextPreview).toContain('JIT context');
        expect(startPayload.contextMatchedPaths).toEqual(['src/app/page.tsx']);
    });

    it('creates a stable fingerprint for deduping repeated injections', () => {
        const fingerprint = buildAutomaticToolContextFingerprint({
            toolName: 'apply_patch',
            query: 'apply_patch MCPServer.ts',
            matchedPaths: ['packages/core/src/MCPServer.ts'],
            observationCount: 1,
            summaryCount: 1,
            prompt: 'JIT tool context for apply_patch',
        });

        expect(fingerprint).toContain('apply_patch');
        expect(fingerprint).toContain('MCPServer.ts');
    });
});