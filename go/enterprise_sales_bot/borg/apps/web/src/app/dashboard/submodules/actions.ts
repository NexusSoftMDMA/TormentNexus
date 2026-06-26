'use server';

import { getSubmodules, SubmoduleInfo } from '@/lib/git';
import path from 'path';
import fs from 'fs/promises';

// Resolve the monorepo root safely without overly broad path traversals
function getMonorepoRoot(): string {
    return process.env.TORMENTNEXUS_ROOT || path.resolve(process.cwd(), '..', '..');
}

export async function fetchSubmodulesAction(): Promise<SubmoduleInfo[]> {
    const root = getMonorepoRoot();
    console.log("Scanning submodules in:", root);
    return await getSubmodules(root);
}

export interface LinkCategory {
    name: string;
    links: string[];
}

export async function fetchUserLinksAction(): Promise<LinkCategory[]> {
    const root = getMonorepoRoot();
    const linksPath = path.join(root, 'docs', 'USER_LINKS_ARCHIVE.md');

    try {
        const content = await fs.readFile(linksPath, 'utf-8');
        const lines = content.split('\n');
        const categories: LinkCategory[] = [];
        let currentCategory: LinkCategory | null = null;

        for (const line of lines) {
            if (line.startsWith('## ')) {
                if (currentCategory) {
                    categories.push(currentCategory);
                }
                currentCategory = { name: line.substring(3).trim(), links: [] };
            } else if (line.trim().startsWith('- http')) {
                if (currentCategory) {
                    currentCategory.links.push(line.trim().substring(2).trim());
                }
            }
        }
        if (currentCategory) {
            categories.push(currentCategory);
        }
        return categories;
    } catch (e) {
        console.error("Failed to read user links:", e);
        return [];
    }
}

export async function healSubmodulesAction(): Promise<{ success: boolean, message: string }> {
    const root = getMonorepoRoot();
    const { exec } = await import('child_process');
    const { promisify } = await import('util');
    const execAsync = promisify(exec);

    try {
        console.log("Healing submodules in:", root);
        // This command initializes and updates all submodules, fixing "missing" or "empty" states.
        // We use --remote to fetch latest if desired, but standard update is safer for stability.
        await execAsync('git submodule update --init --recursive', { cwd: root });
        return { success: true, message: "Submodule heal command executed successfully." };
    } catch (e: any) {
        console.error("Heal failed:", e);
        return { success: false, message: `Failed to heal submodules: ${e.message}` };
    }
}
