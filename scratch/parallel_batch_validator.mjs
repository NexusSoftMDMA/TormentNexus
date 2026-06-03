/**
 * parallel_batch_validator.mjs
 * Runs multiple bulk_validate_mcp_servers.mjs instances in parallel.
 * Each worker operates on a different slice of the "discovered" servers
 * by claiming them first (atomically marking them as "in_progress").
 */

import Database from 'better-sqlite3';
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import path from "path";
import fs from "fs";
import { fileURLToPath } from "url";
import { randomUUID } from "crypto";
import os from "os";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dbPath = path.resolve(__dirname, "..", "tormentnexus.db");
const catalogPath = path.resolve(__dirname, "..", "catalog.db");

if (!fs.existsSync(dbPath)) { console.error(`DB not found: ${dbPath}`); process.exit(1); }
if (!fs.existsSync(catalogPath)) { console.error(`Catalog not found: ${catalogPath}`); process.exit(1); }

const WORKER_COUNT = parseInt(process.env.WORKERS || '4', 10);
const BATCH_SIZE = parseInt(process.env.BATCH_SIZE || '50', 10);

console.log(`[ParallelValidator] Starting with ${WORKER_COUNT} workers, ${BATCH_SIZE} servers per worker...`);
console.log(`[ParallelValidator] Total target: ${WORKER_COUNT * BATCH_SIZE} servers in this run`);

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
        // AI providers
        "OPENAI_API_KEY", "GEMINI_API_KEY", "ANTHROPIC_API_KEY", "OPENROUTER_API_KEY",
        "TAVILY_API_KEY", "FIRECRAWL_API_KEY", "OCTAGON_API_KEY", "MEM0_API_KEY",
        // Code / Git
        "GITHUB_PERSONAL_ACCESS_TOKEN", "GITHUB_TOKEN", "GITLAB_TOKEN", "GITLAB_API_TOKEN",
        // Communication
        "SLACK_BOT_TOKEN", "SLACK_CLIENT_TOKEN", "SLACK_TEAM_ID",
        "DISCORD_TOKEN", "DISCORD_BOT_TOKEN",
        "TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN", "TWILIO_PHONE_NUMBER",
        // Google
        "GMAIL_CLIENT_ID", "GMAIL_CLIENT_SECRET", "GOOGLE_API_KEY",
        "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_ACCESS_TOKEN",
        "GOOGLE_REFRESH_TOKEN",
        // Cloud / media
        "CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET",
        "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION",
        "AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET",
        // Payment / commerce
        "STRIPE_API_KEY", "STRIPE_SECRET_KEY", "STRIPE_PUBLISHABLE_KEY",
        "SHOPIFY_API_KEY", "SHOPIFY_API_SECRET", "SHOPIFY_STORE_URL",
        // Data
        "SUPABASE_URL", "SUPABASE_KEY", "SUPABASE_SERVICE_ROLE_KEY",
        "DATABASE_URL", "MONGODB_URI", "POSTGRES_URL", "REDIS_URL",
        // Search / monitoring
        "ALGOLIA_APP_ID", "ALGOLIA_API_KEY", "SENTRY_DSN",
        "DATADOG_API_KEY", "DATADOG_APP_KEY",
        // Misc
        "NOTION_API_KEY", "NOTION_TOKEN",
        "LINEAR_API_KEY", "JIRA_API_TOKEN", "JIRA_BASE_URL", "JIRA_EMAIL",
        "AIRTABLE_API_KEY", "AIRTABLE_BASE_ID",
        "HUBSPOT_API_KEY", "SALESFORCE_CLIENT_ID", "SALESFORCE_CLIENT_SECRET",
        "SENDGRID_API_KEY", "MAILCHIMP_API_KEY",
        "PINECONE_API_KEY", "PINECONE_ENVIRONMENT",
        "BIGQUERY_PROJECT", "BIGQUERY_KEY_FILE",
        "PERPLEXITY_API_KEY", "COHERE_API_KEY", "MISTRAL_API_KEY",
        // Mesh, Aha, search providers seen blocking in practice
        "MESH_API_KEY", "AHA_API_TOKEN",
        "BRAVE_API_KEY", "KAGI_API_KEY", "LINKUP_API_KEY", "EXA_API_KEY",
        "FLOWFORGE_API_KEY", "ALPHA_VANTAGE_API_KEY",
        "CONTENTFUL_ACCESS_TOKEN", "CONTENTFUL_SPACE_ID",
        "LINEAR_API_KEY", "ASANA_ACCESS_TOKEN", "TRELLO_API_KEY",
        "ZENDESK_API_TOKEN", "INTERCOM_ACCESS_TOKEN",
        "CLOUDFLARE_API_TOKEN", "CLOUDFLARE_ACCOUNT_ID",
        "VERCEL_TOKEN", "NETLIFY_AUTH_TOKEN",
        "DOCKER_USERNAME", "DOCKER_PASSWORD",
        "NPM_TOKEN", "PYPI_TOKEN",
        "SENTRY_AUTH_TOKEN", "GRAFANA_API_KEY",
        "PAGERDUTY_API_KEY", "OPSGENIE_API_KEY",
        "MIXPANEL_TOKEN", "AMPLITUDE_API_KEY",
        "BOX_CLIENT_ID", "BOX_CLIENT_SECRET",
        "DROPBOX_ACCESS_TOKEN", "ONEDRIVE_ACCESS_TOKEN",
        "FIGMA_ACCESS_TOKEN", "MIRO_ACCESS_TOKEN",
        "ZOOM_API_KEY", "ZOOM_API_SECRET",
        "LOTTIEFILES_API_KEY", "ALPHA_VANTAGE_KEY",
        // Generic auth tokens that servers may prompt for
        "TOKEN", "API_TOKEN", "ACCESS_TOKEN", "AUTH_TOKEN", "SECRET_TOKEN",
        "APIDOG_PROJECT_ID", "APIDOG_API_KEY",
        "PLANFORM_API_KEY", "LEMONSQUEEZY_API_KEY",
        "HIDEMIUM_API_KEY", "HIREY_API_KEY",
        "FAL_KEY", "FAL_API_KEY",
        "SCORECARD_API_KEY", "CHENXI_API_KEY",
        "MONGODB_URL", "MONGODB_URI",
        // Additional token keys from latest batch
        "APP_STORE_CONNECT_KEY_ID", "APP_STORE_CONNECT_ISSUER_ID",
        "APP_STORE_CONNECT_P8_PATH", "APP_STORE_CONNECT_P8_CONTENT",
        "PYLON_API_TOKEN",
        "EVM_PRIVATE_KEY", "FINBUD_DATA_API_KEY",
        "TOMBA_API_KEY", "TOMBA_SECRET_KEY",
        "DD_API_KEY", "DD_APP_KEY",
        "HOSTAWAY_API_TOKEN",
        "VESSELAPI_API_KEY", "KEEPSAKE_API_KEY",
        "CONFLUENCE_BASE_URL", "CANVAS_API_TOKEN",
        "AUDIENSE_API_KEY", "FIRE_CRAWL_API_KEY",
        "TODOIST_API_TOKEN",
    ];
    for (const key of defaultSecrets) {
        if (!envObj[key]) {
            envObj[key] = process.env[key] || "YOUR_KEY_HERE";
        }
    }
    return envObj;
}

async function validateOne(db, pubServer, workerId) {
    const prefix = `[Worker-${workerId}]`;
    console.log(`\n${prefix} Testing: "${pubServer.display_name}" (${pubServer.canonical_id})`);

    let recipe;
    try {
        recipe = JSON.parse(pubServer.recipe_template);
    } catch (e) {
        console.error(`${prefix} Bad recipe: ${e.message}`);
        db.prepare("UPDATE catalog.published_mcp_servers SET status = 'failed', updated_at = strftime('%s','now') WHERE uuid = ?").run(pubServer.uuid);
        return;
    }

    if (!recipe || recipe.type !== 'stdio' || !recipe.command) {
        console.log(`${prefix} Skip: non-stdio or empty recipe.`);
        db.prepare("UPDATE catalog.published_mcp_servers SET status = 'failed', updated_at = strftime('%s','now') WHERE uuid = ?").run(pubServer.uuid);
        return;
    }

    const originalCommand = reconcileNaming(recipe.command);
    const originalArgs = (recipe.args || []).map(a => {
        let r = reconcileNaming(a);
        if (typeof r === 'string') r = r.replace(/YOUR_[A-Z0-9_]+_HERE/g, "YOUR_KEY_HERE");
        return r;
    });
    const env = recipe.env || {};

    let runCommand = originalCommand;
    let runArgs = [...originalArgs];

    // Smithery smart rewrite
    if (pubServer.source_url && pubServer.source_url.includes("smithery.ai")) {
        let slug = "";
        if (pubServer.source_url.includes("/server/")) slug = pubServer.source_url.split('/server/')[1];
        else if (pubServer.source_url.includes("/servers/")) slug = pubServer.source_url.split('/servers/')[1];
        if (slug) {
            console.log(`${prefix} Smithery rewrite: ${slug}`);
            runCommand = "npx";
            runArgs = ["-y", "@smithery/cli@latest", "run", slug];
        }
    }

    let childEnv = { ...process.env, ...env };
    childEnv = populateEnvSecrets(childEnv);

    const transport = new StdioClientTransport({ command: runCommand, args: runArgs, env: childEnv });
    const client = new Client({ name: "tormentnexus-parallel-validator", version: "1.0.0" }, { capabilities: {} });

    let outcome = 'failure';
    let toolCount = 0;
    let failureClass = 'unknown';
    let findingsSummary = '';
    const startTime = Date.now();

    try {
        await Promise.race([
            client.connect(transport),
            new Promise((_, reject) => setTimeout(() => reject(new Error("Connection timeout (60s)")), 60000))
        ]);

        const toolsResult = await client.listTools();
        const toolsList = toolsResult.tools || [];
        toolCount = toolsList.length;
        console.log(`${prefix} SUCCESS! ${toolCount} tools found for "${pubServer.display_name}"`);
        outcome = 'success';
        findingsSummary = `Connected successfully. Found ${toolCount} tools.`;

        db.transaction(() => {
            let existing = db.prepare("SELECT uuid FROM mcp_servers WHERE source_published_server_uuid = ?").get(pubServer.uuid);
            let mcpUuid = existing ? existing.uuid : randomUUID();

            if (!existing) {
                db.prepare(`
                    INSERT INTO mcp_servers (uuid, name, description, type, command, args, env, error_status, always_on, created_at, user_id, source_published_server_uuid)
                    VALUES (?, ?, ?, 'stdio', ?, ?, ?, '', 0, strftime('%s','now'), 'system', ?)
                `).run(mcpUuid, pubServer.display_name, pubServer.description || '', originalCommand, JSON.stringify(originalArgs), JSON.stringify(env), pubServer.uuid);
            } else {
                db.prepare(`UPDATE mcp_servers SET command = ?, args = ?, env = ?, error_status = '' WHERE uuid = ?`).run(originalCommand, JSON.stringify(originalArgs), JSON.stringify(env), mcpUuid);
            }

            db.prepare("DELETE FROM tools WHERE mcp_server_uuid = ?").run(mcpUuid);
            const insertTool = db.prepare(`INSERT INTO tools (uuid, name, description, tool_schema, is_deferred, always_on, created_at, updated_at, mcp_server_uuid) VALUES (?, ?, ?, ?, 0, 0, datetime('now'), datetime('now'), ?)`);
            for (const tool of toolsList) {
                insertTool.run(randomUUID(), tool.name, tool.description || '', JSON.stringify(tool.inputSchema || {}), mcpUuid);
            }

            db.prepare(`UPDATE catalog.published_mcp_servers SET status = 'verified', confidence = 1.0, last_verified_at = strftime('%s','now'), updated_at = strftime('%s','now') WHERE uuid = ?`).run(pubServer.uuid);
        })();

    } catch (err) {
        console.error(`${prefix} Failed "${pubServer.display_name}":`, err.message.substring(0, 100));
        failureClass = err.message.includes("timeout") ? "timeout" : "runtime_crash";
        findingsSummary = err.message.substring(0, 500);
        db.prepare(`UPDATE catalog.published_mcp_servers SET status = 'failed', updated_at = strftime('%s','now') WHERE uuid = ?`).run(pubServer.uuid);
    } finally {
        try { await transport.close(); } catch(e) {}
        const endTime = Date.now();
        db.prepare(`
            INSERT INTO catalog.published_mcp_validation_runs (uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at)
            VALUES (?, ?, 'parallel_bulk', ?, ?, ?, ?, ?, ?, 'ParallelValidatorNode', strftime('%s','now'))
        `).run(randomUUID(), pubServer.uuid, Math.floor(startTime/1000), Math.floor(endTime/1000), outcome, failureClass, toolCount, findingsSummary);
    }
}

async function runWorker(workerId, servers) {
    // Open a separate DB connection per worker for thread safety
    const db = new Database(dbPath, { timeout: 30000 });
    db.pragma('busy_timeout = 30000');
    db.pragma('journal_mode = WAL');
    db.prepare(`ATTACH DATABASE '${catalogPath}' AS catalog`).run();

    console.log(`[Worker-${workerId}] Starting with ${servers.length} servers to validate...`);

    for (const server of servers) {
        await validateOne(db, server, workerId);
    }

    console.log(`[Worker-${workerId}] Done!`);
    db.close();
}

async function main() {
    // Claim all servers atomically in one DB connection
    const claimDb = new Database(dbPath, { timeout: 30000 });
    claimDb.pragma('busy_timeout = 30000');
    claimDb.pragma('journal_mode = WAL');
    claimDb.prepare(`ATTACH DATABASE '${catalogPath}' AS catalog`).run();

    const totalNeeded = WORKER_COUNT * BATCH_SIZE;

    // Claim servers by marking them in_progress atomically
    const candidates = claimDb.prepare(`
        SELECT s.uuid, s.canonical_id, s.display_name, s.description, r.template as recipe_template, src.source_url
        FROM catalog.published_mcp_servers s
        JOIN catalog.published_mcp_config_recipes r ON s.uuid = r.server_uuid
        LEFT JOIN catalog.published_mcp_server_sources src ON s.uuid = src.server_uuid AND src.source_name = 'smithery.ai'
        WHERE s.status = 'discovered'
        LIMIT ?
    `).all(totalNeeded);

    if (candidates.length === 0) {
        console.log(`[ParallelValidator] No more 'discovered' servers to validate!`);
        claimDb.close();
        return;
    }

    console.log(`[ParallelValidator] Claimed ${candidates.length} servers for this run.`);

    // Mark them all as in_progress to avoid being picked up by another run
    claimDb.transaction(() => {
        for (const s of candidates) {
            claimDb.prepare("UPDATE catalog.published_mcp_servers SET status = 'in_progress', updated_at = strftime('%s','now') WHERE uuid = ?").run(s.uuid);
        }
    })();
    claimDb.close();

    // Split candidates among workers
    const workerBatches = [];
    for (let i = 0; i < WORKER_COUNT; i++) {
        workerBatches.push(candidates.slice(i * BATCH_SIZE, (i + 1) * BATCH_SIZE));
    }

    // Run all workers in parallel
    const workerPromises = workerBatches
        .map((batch, i) => batch.length > 0 ? runWorker(i + 1, batch) : Promise.resolve());
    
    await Promise.all(workerPromises);

    console.log(`\n[ParallelValidator] All workers finished!`);

    // Final stats
    const statsDb = new Database(catalogPath, { readonly: true });
    const statuses = statsDb.prepare("SELECT status, COUNT(*) as cnt FROM published_mcp_servers GROUP BY status ORDER BY cnt DESC").all();
    console.log(`\n=== Final Catalog Stats ===`);
    statuses.forEach(s => console.log(`  ${s.status}: ${s.cnt}`));
    statsDb.close();
}

main().catch(console.error);
