import sqlite3 from 'better-sqlite3';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");

const db = new sqlite3(dbPath);
console.log(`Connected to tormentnexus.db at: ${dbPath}`);

const patterns = [
    /sk_test_[a-zA-Z0-9]{20,80}/,
    /r8_[a-zA-Z0-9]{30,80}/,
    /sbp_[a-zA-Z0-9]{30,80}/,
    /\b[a-fA-F0-9]{64}\b/
];

// Get all tables
const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all();

for (const table of tables) {
    const tableName = table.name;
    if (tableName.startsWith('sqlite_')) continue;
    
    // Get all columns of this table
    const columns = db.prepare(`PRAGMA table_info(${tableName})`).all();
    
    for (const column of columns) {
        const colName = column.name;
        if (column.type !== 'TEXT') continue; // only scan text columns
        
        try {
            const rows = db.prepare(`SELECT rowid, [${colName}] as val FROM [${tableName}] WHERE [${colName}] IS NOT NULL`).all();
            for (const r of rows) {
                if (typeof r.val !== 'string') continue;
                
                for (let i = 0; i < patterns.length; i++) {
                    const match = r.val.match(patterns[i]);
                    if (match) {
                        console.log(`FOUND MATCH in table [${tableName}], column [${colName}], rowid [${r.rowid}]: pattern index ${i}, match: ${match[0]}`);
                    }
                }
            }
        } catch (err) {
            console.error(`Error scanning [${tableName}].[${colName}]:`, err.message);
        }
    }
}

db.close();
