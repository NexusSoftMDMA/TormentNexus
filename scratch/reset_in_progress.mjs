/**
 * reset_in_progress.mjs
 * Resets any servers stuck in "in_progress" back to "discovered"
 * Run this if the parallel validator crashes mid-way.
 */
import Database from 'better-sqlite3';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const catalogPath = path.resolve(__dirname, '..', 'catalog.db');

const db = new Database(catalogPath);
const stuck = db.prepare("SELECT COUNT(*) as cnt FROM published_mcp_servers WHERE status = 'in_progress'").get();
console.log(`Found ${stuck.cnt} servers stuck in "in_progress" state.`);

if (stuck.cnt > 0) {
    const result = db.prepare("UPDATE published_mcp_servers SET status = 'discovered', updated_at = strftime('%s','now') WHERE status = 'in_progress'").run();
    console.log(`Reset ${result.changes} servers back to "discovered".`);
} else {
    console.log(`Nothing to reset.`);
}

const stats = db.prepare("SELECT status, COUNT(*) as cnt FROM published_mcp_servers GROUP BY status ORDER BY cnt DESC").all();
console.log(`\n=== Catalog Status Breakdown ===`);
stats.forEach(s => console.log(`  ${s.status}: ${s.cnt}`));
db.close();
