import sqlite3 from 'better-sqlite3';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");

const db = new sqlite3(dbPath);
console.log(`Connected to tormentnexus.db at: ${dbPath}`);

const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all();

for (const t of tables) {
    const tableName = t.name;
    if (tableName.startsWith('sqlite_')) continue;
    const cols = db.prepare(`PRAGMA table_info(${tableName})`).all();
    
    for (const c of cols) {
        if (c.type !== 'TEXT') continue;
        try {
            const rows = db.prepare(`
                SELECT rowid, [${c.name}] as val 
                FROM [${tableName}] 
                WHERE [${c.name}] LIKE '%sk_test_%' 
                   OR [${c.name}] LIKE '%r8_%' 
                   OR [${c.name}] LIKE '%sbp_%'
            `).all();
            for (const r of rows) {
                console.log(`MATCH in [${tableName}].[${c.name}], rowid [${r.rowid}]:`, r.val.substring(0, 150));
            }
        } catch(e) {
            console.error(e.message);
        }
    }
}
db.close();
