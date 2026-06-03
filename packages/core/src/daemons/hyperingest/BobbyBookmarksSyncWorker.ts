import Database from 'better-sqlite3';
import path from 'path';
import { randomUUID } from 'crypto';
import { linksBacklogRepository, type UpsertLinkBacklogInput } from '../../db/repositories/links-backlog.repo.js';
import { normalizeBookmarkUrl } from '../../services/bobby-bookmarks-adapter.js';

export class BobbyBookmarksSyncWorker {
    private isRunning = false;
    private interval: NodeJS.Timeout | null = null;
    private readonly dbPath: string;

    constructor(workspaceRoot: string = process.cwd()) {
        this.dbPath = path.join(workspaceRoot, 'data', 'bobbybookmarks', 'resources.db');
    }

    public start(intervalMs: number = 60 * 60 * 1000): void {
        if (this.isRunning) return;
        this.isRunning = true;
        console.log('[tormentnexusingest] Starting BobbyBookmarks local DB sync worker...');
        
        // Run immediately
        void this.sync();

        this.interval = setInterval(() => {
            void this.sync();
        }, intervalMs);
    }

    public stop(): void {
        this.isRunning = false;
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
        console.log('[tormentnexusingest] Stopped BobbyBookmarks local DB sync worker.');
    }

    private async sync(): Promise<void> {
        try {
            // Check if the database file exists before attempting to open it
      const fs = await import('fs/promises');
      try { await fs.access(this.dbPath); } catch { return; }
      const db = new Database(this.dbPath, { readonly: true });
            
            // We just need the basic details from the local DB
            const rows = db.prepare(`SELECT * FROM bookmarks`).all() as any[];
            db.close();

            console.log(`[tormentnexusingest] Found ${rows.length} raw URLs in local BobbyBookmarks DB. Syncing to tormentnexus...`);

            let synced = 0;
            let errors = 0;
            for (const row of rows) {
                try {
                    const normalizedUrl = normalizeBookmarkUrl(row.url);
                    if (!normalizedUrl) continue;

                    const payload: UpsertLinkBacklogInput = {
                        url: row.url,
                        normalized_url: normalizedUrl,
                        title: row.short_description || row.url,
                        description: row.long_description,
                        tags: row.tags ? String(row.tags).split(',').map((t: string) => t.trim()).filter(Boolean) : [],
                        source: 'bobbybookmarks',
                        is_duplicate: false,
                        research_status: row.research_level === 'done' ? 'done' : 'pending',
                        bobbybookmarks_bookmark_id: row.id,
                        raw_payload: row as Record<string, unknown>,
                    };

                    await linksBacklogRepository.upsertLink(payload);
                    synced++;
                } catch (e: any) {
                    errors++;
                }
            }

            console.log(`[tormentnexusingest] Successfully synced ${synced} bookmarks to the Links Backlog (${errors} errors).`);
        } catch (error: any) {
            if (error.code === 'SQLITE_CANTOPEN') {
                console.warn(`[tormentnexusingest] BobbyBookmarks DB not found at ${this.dbPath}. Skipping sync.`);
            } else {
                console.error('[tormentnexusingest] Error during sync:', error);
            }
        }
    }
}

