import fs from 'fs/promises';
import path from 'path';

import { SECTIONED_MEMORY_DEFAULT_SECTIONS } from '../services/memory/SectionedMemoryAdapter.js';
import type { MemoryPipelineSummary } from '../services/memory/MemoryManager.js';

export type SectionedMemoryRuntimePipelineStatus = {
    configuredMode: MemoryPipelineSummary['configuredMode'] | 'unknown';
    providerNames: string[];
    providerCount: number;
    sectionedStoreEnabled: boolean;
};

export type SectionedMemorySectionStatus = {
    section: string;
    entryCount: number;
};

export type SectionedMemoryStoreStatus = {
    exists: boolean;
    storePath: string;
    totalEntries: number;
    sectionCount: number;
    defaultSectionCount: number;
    presentDefaultSectionCount: number;
    populatedSectionCount: number;
    missingSections: string[];
    runtimePipeline: SectionedMemoryRuntimePipelineStatus;
    sections: SectionedMemorySectionStatus[];
    lastUpdatedAt: string | null;
};

type RawSectionedMemoryStore = {
    sections?: Array<{
        section?: string;
        entries?: Array<{
            createdAt?: string;
        }>;
    }>;
};

const LEGACY_STORE_FILE = ['claude', '_mem.json'].join('');

export function summarizeSectionedMemoryRuntimePipeline(pipelineSummary?: MemoryPipelineSummary | null): SectionedMemoryRuntimePipelineStatus {
    if (!pipelineSummary) {
        return {
            configuredMode: 'unknown',
            providerNames: [],
            providerCount: 0,
            sectionedStoreEnabled: false,
        };
    }

    return {
        configuredMode: pipelineSummary.configuredMode,
        providerNames: pipelineSummary.providerNames,
        providerCount: pipelineSummary.providerCount,
        sectionedStoreEnabled: pipelineSummary.sectionedStoreEnabled,
    };
}

export function summarizeSectionedMemoryStore(storePath: string, rawStore: RawSectionedMemoryStore | null | undefined, pipelineSummary?: MemoryPipelineSummary | null): SectionedMemoryStoreStatus {
    const sections = Array.isArray(rawStore?.sections) ? rawStore.sections : [];
    const normalizedSections = sections.map((section, index) => ({
        section: typeof section?.section === 'string' && section.section.trim().length > 0
            ? section.section
            : `section_${index + 1}`,
        entryCount: Array.isArray(section?.entries) ? section.entries.length : 0,
    }));
    const presentSectionNames = new Set(normalizedSections.map((section) => section.section));
    const missingSections = SECTIONED_MEMORY_DEFAULT_SECTIONS.filter((section) => !presentSectionNames.has(section));
    const populatedSectionCount = normalizedSections.filter((section) => section.entryCount > 0).length;
    const presentDefaultSectionCount = SECTIONED_MEMORY_DEFAULT_SECTIONS.filter((section) => presentSectionNames.has(section)).length;

    let lastUpdatedAt: string | null = null;
    for (const section of sections) {
        const entries = Array.isArray(section?.entries) ? section.entries : [];
        for (const entry of entries) {
            if (typeof entry?.createdAt !== 'string') continue;
            const currentDate = new Date(entry.createdAt);
            if (Number.isNaN(currentDate.getTime())) continue;

            if (!lastUpdatedAt || currentDate.getTime() > new Date(lastUpdatedAt).getTime()) {
                lastUpdatedAt = currentDate.toISOString();
            }
        }
    }

    const totalEntries = normalizedSections.reduce((sum, section) => sum + section.entryCount, 0);

    return {
        exists: Boolean(rawStore),
        storePath,
        totalEntries,
        sectionCount: normalizedSections.length,
        defaultSectionCount: SECTIONED_MEMORY_DEFAULT_SECTIONS.length,
        presentDefaultSectionCount,
        populatedSectionCount,
        missingSections,
        runtimePipeline: summarizeSectionedMemoryRuntimePipeline(pipelineSummary),
        sections: normalizedSections,
        lastUpdatedAt,
    };
}

export async function readSectionedMemoryStoreStatus(workspaceRoot: string, pipelineSummary?: MemoryPipelineSummary | null): Promise<SectionedMemoryStoreStatus> {
    const storePath = path.join(workspaceRoot, '.tormentnexus', 'sectioned_memory.json');
    const legacyStorePath = path.join(workspaceRoot, '.tormentnexus', LEGACY_STORE_FILE);

    for (const candidatePath of [storePath, legacyStorePath]) {
        try {
            const raw = await fs.readFile(candidatePath, 'utf-8');
            const parsed = JSON.parse(raw) as RawSectionedMemoryStore;
            return summarizeSectionedMemoryStore(storePath, parsed, pipelineSummary);
        } catch (error: unknown) {
            const code = typeof error === 'object' && error !== null && 'code' in error ? String((error as { code?: unknown }).code) : '';
            if (code === 'ENOENT') {
                continue;
            }

            throw error;
        }
    }

    return summarizeSectionedMemoryStore(storePath, null, pipelineSummary);
}