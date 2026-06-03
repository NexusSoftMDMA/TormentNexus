import sqlite3 from 'better-sqlite3';
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import path from "path";
import fs from "fs";
import { fileURLToPath } from "url";
import { randomUUID } from "crypto";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");

if (!fs.existsSync(dbPath)) {
    console.error(`Database not found at ${dbPath}`);
    process.exit(1);
}

const db = new sqlite3(dbPath);
// Set busy_timeout and journal mode to prevent SQLITE_BUSY
db.pragma('busy_timeout = 20000');
db.pragma('journal_mode = WAL');

const catalogPath = path.resolve(__dirname, "..", "catalog.db");
if (!fs.existsSync(catalogPath)) {
    console.error(`Catalog database not found at ${catalogPath}`);
    process.exit(1);
}
db.prepare(`ATTACH DATABASE '${catalogPath}' AS catalog`).run();

console.log(`[Bulk-Validator] Connected to tormentnexus.db and attached catalog.db successfully.`);

// Brand-reconciliation helper
function reconcileNaming(text) {
    if (typeof text !== 'string') return text;
    let result = text
        .replace(/tormentnexus-supervisor/gi, 'tormentnexus-supervisor')
        .replace(/TormentNexus-supervisor/gi, 'tormentnexus-supervisor');
    
    if (result.toLowerCase() === 'tormentnexus' || result.toLowerCase() === 'nexus' || result.toLowerCase() === 'TormentNexus') {
        return 'tormentnexus';
    }
    return result;
}

// Generate fallback dummy keys to avoid instant crashes on boot
function populateEnvSecrets(envObj) {
    const defaultSecrets = [
        "OPENAI_API_KEY", "GEMINI_API_KEY", "ANTHROPIC_API_KEY", "OPENROUTER_API_KEY", 
        "TAVILY_API_KEY", "FIRECRAWL_API_KEY", "OCTAGON_API_KEY", "MEM0_API_KEY",
        "GITHUB_PERSONAL_ACCESS_TOKEN", "GITHUB_TOKEN", "SLACK_BOT_TOKEN", "SLACK_CLIENT_TOKEN",
        "GMAIL_CLIENT_ID", "GMAIL_CLIENT_SECRET", "GOOGLE_API_KEY"
    ];
    for (const key of defaultSecrets) {
        if (!envObj[key]) {
            if (process.env[key]) {
                envObj[key] = process.env[key];
            } else {
                envObj[key] = "YOUR_KEY_HERE";
            }
        }
    }
    return envObj;
}

async function validatePublishedServer(pubServer) {
    console.log(`\n==================================================`);
    console.log(`[Bulk-Validator] Testing: "${pubServer.display_name}" (${pubServer.canonical_id})`);
    
    let recipe;
    try {
        recipe = JSON.parse(pubServer.recipe_template);
    } catch (e) {
        console.error(`[Bulk-Validator] Failed to parse recipe template for ${pubServer.canonical_id}:`, e.message);
        db.prepare("UPDATE catalog.published_mcp_servers SET status = 'failed' WHERE uuid = ?").run(pubServer.uuid);
        return;
    }

    if (!recipe || recipe.type !== 'stdio' || !recipe.command) {
        console.log(`[Bulk-Validator] Skip: non-stdio or empty recipe template for ${pubServer.canonical_id}.`);
        db.prepare("UPDATE catalog.published_mcp_servers SET status = 'failed' WHERE uuid = ?").run(pubServer.uuid);
        return;
    }

    const originalCommand = reconcileNaming(recipe.command);
    const originalArgs = recipe.args || [];
    const env = recipe.env || {};

    // Resolve original arguments
    const originalArgsParsed = originalArgs.map(arg => {
        let resolved = reconcileNaming(arg);
        if (typeof resolved === 'string') {
            resolved = resolved.replace(/YOUR_[A-Z0-9_]+_HERE/g, "YOUR_KEY_HERE");
        }
        return resolved;
    });

    let runCommand = originalCommand;
    let runArgsParsed = [...originalArgsParsed];

    // SMART REWRITE: If it's a Smithery server, use the Smithery CLI to run it correctly!
    let isSmithery = false;
    let smitherySlug = "";
    if (pubServer.source_url && pubServer.source_url.includes("smithery.ai")) {
        let url = pubServer.source_url;
        if (url.includes("/server/")) {
            smitherySlug = url.split('/server/')[1];
            isSmithery = true;
        } else if (url.includes("/servers/")) {
            smitherySlug = url.split('/servers/')[1];
            isSmithery = true;
        }
    }

    if (isSmithery && smitherySlug) {
        console.log(`[Bulk-Validator] Smart Rewrite: Smithery slug "${smitherySlug}" found! Using @smithery/cli for verification run.`);
        runCommand = "npx";
        runArgsParsed = ["-y", "@smithery/cli@latest", "run", smitherySlug];
    }

    // Standardize & enrich env
    let childEnv = { ...process.env, ...env };
    childEnv = populateEnvSecrets(childEnv);

    console.log(`[Bulk-Validator] Stdio executable (original): "${originalCommand}"`);
    console.log(`[Bulk-Validator] Stdio args (original):`, originalArgsParsed);
    console.log(`[Bulk-Validator] Stdio executable (run): "${runCommand}"`);
    console.log(`[Bulk-Validator] Stdio args (run):`, runArgsParsed);

    const transport = new StdioClientTransport({
        command: runCommand,
        args: runArgsParsed,
        env: childEnv
    });

    const client = new Client(
        { name: "tormentnexus-bulk-validator", version: "1.0.0" },
        { capabilities: {} }
    );

    let runOutcome = 'failure';
    let toolCount = 0;
    let failureClass = 'unknown';
    let findingsSummary = '';

    const startTime = Date.now();

    try {
        // Run connection with a 60 second deadline to support slow packages install (npx/uvx)
        await Promise.race([
            client.connect(transport),
            new Promise((_, reject) => setTimeout(() => reject(new Error("Connection timeout (60s)")), 60000))
        ]);

        console.log(`[Bulk-Validator] Successfully connected! Listing tools...`);
        const toolsResult = await client.listTools();
        const toolsList = toolsResult.tools || [];
        toolCount = toolsList.length;
        console.log(`[Bulk-Validator] Connection successful! Found ${toolCount} tools.`);

        runOutcome = 'success';
        findingsSummary = `Successfully connected and loaded ${toolCount} tools.`;

        // Register the server in the primary mcp_servers table
        db.transaction(() => {
            // Check if server is already registered in mcp_servers
            let mcpServer = db.prepare("SELECT uuid FROM mcp_servers WHERE source_published_server_uuid = ?").get(pubServer.uuid);
            let mcpUuid = mcpServer ? mcpServer.uuid : randomUUID();

            if (!mcpServer) {
                db.prepare(`
                    INSERT INTO mcp_servers (uuid, name, description, type, command, args, env, error_status, always_on, created_at, user_id, source_published_server_uuid)
                    VALUES (?, ?, ?, 'stdio', ?, ?, ?, '', 0, strftime('%s','now'), 'system', ?)
                `).run(
                    mcpUuid,
                    pubServer.display_name,
                    pubServer.description || '',
                    originalCommand,
                    JSON.stringify(originalArgsParsed),
                    JSON.stringify(env),
                    pubServer.uuid
                );
            } else {
                db.prepare(`
                    UPDATE mcp_servers 
                    SET command = ?, args = ?, env = ?, error_status = '', user_id = 'system'
                    WHERE uuid = ?
                `).run(originalCommand, JSON.stringify(originalArgsParsed), JSON.stringify(env), mcpUuid);
            }

            // Remove existing tools associated with this server
            db.prepare("DELETE FROM tools WHERE mcp_server_uuid = ?").run(mcpUuid);

            // Register new tools
            const insertTool = db.prepare(`
                INSERT INTO tools (uuid, name, description, tool_schema, is_deferred, always_on, created_at, updated_at, mcp_server_uuid)
                VALUES (?, ?, ?, ?, 0, 0, datetime('now'), datetime('now'), ?)
            `);

            for (const tool of toolsList) {
                insertTool.run(
                    randomUUID(),
                    tool.name,
                    tool.description || '',
                    JSON.stringify(tool.inputSchema || {}),
                    mcpUuid
                );
            }

            // Mark published catalog server as verified
            db.prepare(`
                UPDATE catalog.published_mcp_servers
                SET status = 'verified', confidence = 1.0, last_verified_at = strftime('%s','now'), updated_at = strftime('%s','now')
                WHERE uuid = ?
            `).run(pubServer.uuid);
        })();

        console.log(`[Bulk-Validator] Registered successfully into tormentnexus.db!`);

    } catch (err) {
        console.error(`[Bulk-Validator] Failed:`, err.message);
        failureClass = err.message.includes("timeout") ? "timeout" : "runtime_crash";
        findingsSummary = err.message;

        db.prepare(`
            UPDATE catalog.published_mcp_servers
            SET status = 'failed', updated_at = strftime('%s','now')
            WHERE uuid = ?
        `).run(pubServer.uuid);
    } finally {
        try {
            await transport.close();
        } catch(e) {}

        const endTime = Date.now();
        // Log to published_mcp_validation_runs
        db.prepare(`
            INSERT INTO catalog.published_mcp_validation_runs (uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at)
            VALUES (?, ?, 'automated_bulk', ?, ?, ?, ?, ?, ?, 'BulkValidatorNode', strftime('%s','now'))
        `).run(
            randomUUID(),
            pubServer.uuid,
            Math.floor(startTime / 1000),
            Math.floor(endTime / 1000),
            runOutcome,
            failureClass,
            toolCount,
            findingsSummary
        );
    }
}

async function main() {
    // Select discovered servers, joining with recipes and sources to get correct Smithery URLs
    const query = `
        SELECT s.uuid, s.canonical_id, s.display_name, s.description, r.template as recipe_template, src.source_url
        FROM catalog.published_mcp_servers s
        JOIN catalog.published_mcp_config_recipes r ON s.uuid = r.server_uuid
        LEFT JOIN catalog.published_mcp_server_sources src ON s.uuid = src.server_uuid AND src.source_name = 'smithery.ai'
        WHERE s.status = 'discovered'
        LIMIT 100
    `;

    const candidateServers = db.prepare(query).all();
    console.log(`[Bulk-Validator] Found ${candidateServers.length} candidate servers to validate in this batch.`);

    for (const s of candidateServers) {
        await validatePublishedServer(s);
    }

    console.log(`\n[Bulk-Validator] Batch validation completed!`);
    db.close();
}

main().catch(console.error);
