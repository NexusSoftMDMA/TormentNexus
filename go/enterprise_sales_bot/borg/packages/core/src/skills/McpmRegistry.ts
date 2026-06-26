
import fs from 'fs/promises';
import path from 'path';

import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export interface RegistryItem {
    name: string;
    url: string;
    tags?: string[];
}

interface RegistryData {
    directories: RegistryItem[];
    skills: RegistryItem[];
    [category: string]: unknown;
}

const DEFAULT_REGISTRY_DATA: RegistryData = {
    directories: [],
    skills: [],
};

function isRegistryItem(value: unknown): value is RegistryItem {
    if (!value || typeof value !== 'object') {
        return false;
    }

    const record = value as Record<string, unknown>;
    const hasValidName = typeof record.name === 'string';
    const hasValidUrl = typeof record.url === 'string';
    const hasValidTags =
        record.tags === undefined ||
        (Array.isArray(record.tags) && record.tags.every(tag => typeof tag === 'string'));

    return hasValidName && hasValidUrl && hasValidTags;
}

function normalizeRegistryData(value: unknown): RegistryData {
    if (!value || typeof value !== 'object') {
        return DEFAULT_REGISTRY_DATA;
    }

    const record = value as Record<string, unknown>;
    const directories = Array.isArray(record.directories)
        ? record.directories.filter(isRegistryItem)
        : [];
    const skills = Array.isArray(record.skills)
        ? record.skills.filter(isRegistryItem)
        : [];

    return {
        ...record,
        directories,
        skills,
    };
}

export class McpmRegistry {
    private dataPath: string;
    private cache: RegistryData | null = null;

    constructor() {
        // Point to skills_registry.json in ../data
        this.dataPath = path.join(__dirname, '../../data/skills_registry.json');
    }

    async load(): Promise<RegistryData> {
        if (this.cache) return this.cache;

        try {
            const content = await fs.readFile(this.dataPath, 'utf-8');
            this.cache = normalizeRegistryData(JSON.parse(content));
            return this.cache;
        } catch (error) {
            console.error("Failed to load MCP Registry:", error);
            return DEFAULT_REGISTRY_DATA;
        }
    }

    async search(query: string): Promise<RegistryItem[]> {
        const data = await this.load();
        const allItems = [...(data.directories || []), ...(data.skills || [])];

        return allItems.filter(item =>
            item.name.toLowerCase().includes(query.toLowerCase()) ||
            (item.tags && item.tags.some((t: string) => t.includes(query.toLowerCase())))
        );
    }

    async listCategories() {
        const data = await this.load();
        return Object.keys(data);
    }
}
