'use server';

import fs from 'fs/promises';
import path from 'path';

// Resolve the monorepo root safely without overly broad path traversals
function getMonorepoRoot(): string {
    return process.env.TORMENTNEXUS_ROOT || path.resolve(process.cwd(), '..', '..');
}

const CONFIG_PATH = path.join(getMonorepoRoot(), 'packages', 'core', 'config', 'council.json');

export async function getCouncilConfig() {
    try {
        const content = await fs.readFile(CONFIG_PATH, 'utf-8');
        return JSON.parse(content);
    } catch (e) {
        console.error("Failed to read council config:", e);
        // Return default if missing
        return {
            members: [
                { name: "The Architect", provider: "ollama", modelId: "gemma", systemPrompt: "Default architect." }
            ]
        };
    }
}

export async function saveCouncilConfig(config: any) {
    try {
        await fs.writeFile(CONFIG_PATH, JSON.stringify(config, null, 4), 'utf-8');
        return { success: true };
    } catch (e: any) {
        console.error("Failed to save config:", e);
        return { success: false, error: e.message };
    }
}
