import sqlite3 from 'better-sqlite3';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");

const db = new sqlite3(dbPath);
console.log(`Connected to tormentnexus.db at: ${dbPath}`);

// We will find all text in template column of published_mcp_config_recipes
const recipes = db.prepare("SELECT uuid, template FROM published_mcp_config_recipes").all();

let updatedRecipes = 0;
const updateRecipeStmt = db.prepare("UPDATE published_mcp_config_recipes SET template = ? WHERE uuid = ?");

// Patterns to mask
const patterns = [
    // Stripe test key: sk_test_...
    /sk_test_[a-zA-Z0-9]{20,80}/g,
    // Replicate token: r8_...
    /r8_[a-zA-Z0-9]{30,80}/g,
    // Supabase token: sbp_...
    /sbp_[a-zA-Z0-9]{30,80}/g,
    // Generic high-entropy hex tokens (like Sentry personal tokens which are 64 characters of hex)
    /\b[a-fA-F0-9]{64}\b/g
];

db.transaction(() => {
    for (const r of recipes) {
        if (!r.template) continue;
        let masked = r.template;
        let matched = false;
        
        for (const pattern of patterns) {
            pattern.lastIndex = 0;
            const replaced = masked.replace(pattern, "MASKED_KEY");
            if (replaced !== masked) {
                masked = replaced;
                matched = true;
            }
        }
        
        if (matched) {
            updateRecipeStmt.run(masked, r.uuid);
            updatedRecipes++;
        }
    }
})();

console.log(`Masked secrets in ${updatedRecipes} recipes.`);

// Also let's check published_mcp_server_sources
const sources = db.prepare("SELECT uuid, source_url FROM published_mcp_server_sources").all();
let updatedSources = 0;
const updateSourceStmt = db.prepare("UPDATE published_mcp_server_sources SET source_url = ? WHERE uuid = ?");

db.transaction(() => {
    for (const s of sources) {
        if (!s.source_url) continue;
        let masked = s.source_url;
        let matched = false;
        
        for (const pattern of patterns) {
            pattern.lastIndex = 0;
            const replaced = masked.replace(pattern, "MASKED_KEY");
            if (replaced !== masked) {
                masked = replaced;
                matched = true;
            }
        }
        
        if (matched) {
            updateSourceStmt.run(masked, s.uuid);
            updatedSources++;
        }
    }
})();

console.log(`Masked secrets in ${updatedSources} sources.`);

db.close();
