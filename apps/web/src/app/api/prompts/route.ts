
import { NextResponse } from 'next/server';
import fs from 'fs/promises';
import path from 'path';

// Resolve the monorepo root safely without overly broad path traversals
function getMonorepoRoot(): string {
    return process.env.TORMENTNEXUS_ROOT || path.resolve(process.cwd(), '..', '..');
}

const PROMPTS_DIR = path.join(getMonorepoRoot(), '.tormentnexus', 'prompts');

function normalizePromptId(id: string): string {
    return id.replace(/[^a-zA-Z0-9._-]/g, '_');
}

function extractTemplateVariables(template: string): string[] {
    const names = new Set<string>();

    // Match {{variable}} style placeholders
    const doubleBracePattern = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)\s*\}\}/g;
    for (const match of template.matchAll(doubleBracePattern)) {
        const name = match[1]?.trim();
        if (name) names.add(name);
    }

    // Match ${variable} style placeholders
    const dollarBracePattern = /\$\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)\s*\}/g;
    for (const match of template.matchAll(dollarBracePattern)) {
        const name = match[1]?.trim();
        if (name) names.add(name);
    }

    return Array.from(names).sort((a, b) => a.localeCompare(b));
}

export async function GET() {
    try {
        await fs.mkdir(PROMPTS_DIR, { recursive: true });
        const files = await fs.readdir(PROMPTS_DIR);
        const prompts = [];

        for (const file of files) {
            if (file.endsWith('.json')) {
                const content = await fs.readFile(path.join(PROMPTS_DIR, file), 'utf-8');
                try {
                    prompts.push(JSON.parse(content));
                } catch { }
            }
        }

        return NextResponse.json({ prompts });
    } catch (e: any) {
        return NextResponse.json({ error: e.message }, { status: 500 });
    }
}

export async function POST(req: Request) {
    try {
        await fs.mkdir(PROMPTS_DIR, { recursive: true });

        const body = await req.json();
        const { id, template, description } = body;

        if (!id || !template) {
            return NextResponse.json({ error: "Missing required fields" }, { status: 400 });
        }

        if (typeof id !== 'string' || typeof template !== 'string') {
            return NextResponse.json({ error: "Invalid payload types" }, { status: 400 });
        }

        const safeId = normalizePromptId(id);
        const filePath = path.join(PROMPTS_DIR, `${safeId}.json`);

        // Read existing to bump version
        let version = 1;
        try {
            const existing = JSON.parse(await fs.readFile(filePath, 'utf-8'));
            version = (existing.version || 0) + 1;
        } catch { }

        const promptData = {
            id: safeId,
            version,
            description: description || "",
            template,
            variables: extractTemplateVariables(template),
            updatedAt: new Date().toISOString()
        };

        await fs.writeFile(filePath, JSON.stringify(promptData, null, 2));

        return NextResponse.json({ success: true, prompt: promptData });

    } catch (e: any) {
        return NextResponse.json({ error: e.message }, { status: 500 });
    }
}
