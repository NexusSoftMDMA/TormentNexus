import sqlite3 from 'better-sqlite3';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");
const catalogPath = path.resolve(__dirname, "..", "catalog.db");

if (fs.existsSync(catalogPath)) {
    console.log(`catalog.db already exists, removing it to start fresh.`);
    fs.unlinkSync(catalogPath);
}

const db = new sqlite3(dbPath);
console.log(`Connected to tormentnexus.db at: ${dbPath}`);

// Attach catalog.db
db.prepare(`ATTACH DATABASE '${catalogPath}' AS catalog`).run();
console.log(`Attached catalog.db at: ${catalogPath}`);

const catalogTables = [
    'published_mcp_servers',
    'published_mcp_server_sources',
    'published_mcp_config_recipes',
    'published_mcp_validation_runs'
];

db.transaction(() => {
    for (const table of catalogTables) {
        console.log(`Copying table: ${table}...`);
        
        // Get the create table statement
        const createSqlRow = db.prepare(`SELECT sql FROM sqlite_master WHERE type='table' AND name=?`).get(table);
        if (!createSqlRow || !createSqlRow.sql) {
            console.warn(`Table ${table} not found in tormentnexus.db, skipping.`);
            continue;
        }
        
        // Create table in catalog.db
        const createSql = createSqlRow.sql.replace(new RegExp(`CREATE TABLE "${table}"`, 'g'), `CREATE TABLE catalog."${table}"`)
                                           .replace(new RegExp(`CREATE TABLE ${table}`, 'g'), `CREATE TABLE catalog.${table}`);
        
        db.prepare(createSql).run();
        
        // Copy data
        db.prepare(`INSERT INTO catalog.${table} SELECT * FROM main.${table}`).run();
        
        const count = db.prepare(`SELECT count(*) as cnt FROM catalog.${table}`).get().cnt;
        console.log(`Successfully copied ${count} rows into catalog.${table}.`);
    }
})();

console.log(`All tables successfully copied to catalog.db.`);

// Drop tables from main
console.log(`Dropping tables from tormentnexus.db...`);
for (const table of catalogTables) {
    db.prepare(`DROP TABLE main.${table}`).run();
    console.log(`Dropped main.${table}.`);
}

// Detach
db.prepare(`DETACH DATABASE catalog`).run();
db.close();

// VACUUM main to reclaim space
console.log(`Vacuuming tormentnexus.db to shrink file size...`);
const dbVacuum = new sqlite3(dbPath);
dbVacuum.prepare(`VACUUM`).run();
dbVacuum.close();

console.log(`Database split complete! tormentnexus.db size: ${fs.statSync(dbPath).size} bytes.`);
