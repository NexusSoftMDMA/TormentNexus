
import { NextResponse } from 'next/server';
import fs from 'fs/promises';
import path from 'path';

// Resolve the monorepo root safely without overly broad path traversals
// that trigger Turbopack's file pattern analysis
function getMonorepoRoot(): string {
    return process.env.TORMENTNEXUS_ROOT || path.resolve(process.cwd(), '..', '..');
}

export async function GET() {
    try {
        const rootDir = getMonorepoRoot();

        const logFile = path.join(rootDir, '.tormentnexus', 'data', 'healer_events.jsonl');

        try {
            await fs.access(logFile);
        } catch {
            return NextResponse.json({ events: [] });
        }

        const data = await fs.readFile(logFile, 'utf-8');
        const lines = data.trim().split('\n');

        // Parse and reverse to show newest first
        const events = lines
            .map(line => {
                try { return JSON.parse(line); } catch { return null; }
            })
            .filter(Boolean)
            .reverse()
            .slice(0, 50); // Limit to last 50

        return NextResponse.json({ events });

    } catch (e: any) {
        return NextResponse.json({ error: e.message }, { status: 500 });
    }
}
