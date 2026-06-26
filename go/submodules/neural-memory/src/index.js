#!/usr/bin/env node
/**
 * Neural Memory MCP Server v2.0
 * Server MCP per memoria persistente e organizzata per agenti AI
 * 
 * CAMBIAMENTI v2.0:
 * - Session Management integrato (no più project-based)
 * - Skills Framework con schema rigido
 * - Context Compression
 * - Reports HTML
 * - Database Unificato
 */

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { registerAllTools } from "./tools/index.js";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";
import { z } from "zod";

// Leggi versione da package.json
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const packageJson = JSON.parse(readFileSync(join(__dirname, "..", "package.json"), "utf-8"));

// Crea istanza server con versione da package.json
const server = new McpServer({
  name: "neural-memory",
  version: packageJson.version,
});

// Registra tutti i tool
const registeredTools = registerAllTools(server);

// Avvio server
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  
  console.error("Neural Memory MCP Server v2.0 started!");
  console.error("Database: Unified SQLite");
  console.error("Available tools:");
  registeredTools.forEach(t => {
    console.error(`  - ${t.name}: ${t.description.substring(0, 60)}...`);
  });
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
