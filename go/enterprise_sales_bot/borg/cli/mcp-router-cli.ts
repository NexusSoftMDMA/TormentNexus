#!/usr/bin/env node

/**
 * tormentnexus MCP Router CLI
 *
 * A simple CLI to interact with the Ultimate MCP Router services.
 * This provides immediate access to all 4 core services without requiring
 * integration with the complex McpProxyManager.
 */

import { Command } from 'commander';
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { MCPRegistryService, type RegistryServerDefinition } from '../packages/core/src/services/MCPRegistryService.js';
import { ServerRegistryService, type ServerInstallationResult } from '../packages/core/src/services/ServerRegistryService.js';
import { ConfigurationService, type ConfigFile } from '../packages/core/src/services/ConfigurationService.js';
import { McpSessionService, type PerformanceMetrics } from '../packages/core/src/services/McpSessionService.js';

// ============================================
// CLI Implementation
// ============================================

const program = new Command();
const __dirname = dirname(fileURLToPath(import.meta.url));

function readCanonicalVersion(): string {
    const candidates = [
        resolve(__dirname, '..', 'VERSION'),
        resolve(__dirname, '..', 'VERSION.md'),
    ];

    for (const candidate of candidates) {
        try {
            const raw = readFileSync(candidate, 'utf-8').trim();
            const match = raw.match(/\d+\.\d+\.\d+(?:-[\w.-]+)?/);
            if (match) {
                return match[0];
            }
        } catch {
            // Keep trying.
        }
    }

    return '0.0.0';
}

program
    .name('tormentnexus-mcp-router')
    .description('Ultimate MCP Router - Manage MCP servers, configurations, and sessions')
    .version(readCanonicalVersion())
    .option('--data-dir <path>', 'Data directory path', './data')
    .option('--format <type>', 'Output format (json, table)', 'json');

// ============================================
// Services
// ============================================

const dataDir = program.opts().dataDir || './data';

console.log('Initializing tormentnexus MCP Router...');
console.log(`Data directory: ${dataDir}`);

const registry = MCPRegistryService.getInstance(dataDir);
const serverRegistry = ServerRegistryService.getInstance(dataDir);
const configService = ConfigurationService.getInstance(dataDir);
const sessionService = McpSessionService.getInstance(dataDir);

// ============================================
// Commands
// ============================================

// Discover Command
program
    .command('discover')
    .description('Discover servers from all registries')
    .action(async () => {
        console.log('📡 Discovering servers from registries...');
        const servers = await registry.discoverAll();
        console.log(`✅ Discovered ${servers.length} servers from ${registry.getStats().categories.length} categories`);
    });

// Search Command
program
    .command('search [query]')
    .description('Search discovered servers')
    .action(async (options, command) => {
        const query = command.args[0] || '';
        console.log(`🔍 Searching for: ${query}`);
        const results = registry.searchServers(query);
        console.log(`✅ Found ${results.length} servers`);
        if (program.opts().format === 'table') {
            console.table(results.map(r => ({
                Name: r.name,
                Category: r.category,
                Rating: r.rating || '-',
                Installed: r.installed ? '✅' : '❌',
                Source: r.source
            })));
        } else {
            console.log(JSON.stringify(results, null, 2));
        }
    });

// List Categories Command
program
    .command('categories')
    .description('List all available categories')
    .action(async () => {
        const categories = registry.getCategories();
        console.log('📂 Available categories:');
        categories.forEach(cat => console.log(`   - ${cat}`));
    });

// Stats Command
program
    .command('stats')
    .description('Get registry statistics')
    .action(async () => {
        const stats = registry.getStats();
        console.log('📊 Registry Statistics:');
        console.log(`   Total servers: ${stats.totalServers}`);
        console.log(`   Installed: ${stats.installedServers}`);
        console.log(`   Categories: ${stats.categories.length}`);
        console.log(`   Last sync: ${new Date(stats.lastSync).toISOString()}`);
    });

// Install Command
program
    .command('install <name>')
    .description('Install a server from registry')
    .action(async (options, command) => {
        const name = command.args[0];
        console.log(`📦 Installing server: ${name}`);
        const result: ServerInstallationResult = await serverRegistry.installServer(name, {
            type: 'github',
            autoStart: true
        });

        if (result.success) {
            console.log(`✅ Installed successfully: ${result.serverId}`);
        } else {
            console.error(`❌ Installation failed: ${result.error}`);
            process.exit(1);
        }
    });

// Uninstall Command
program
    .command('uninstall <serverId>')
    .description('Uninstall a server')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`🗑️ Uninstalling server: ${serverId}`);
        const success = await serverRegistry.uninstallServer(serverId);

        if (success) {
            console.log('✅ Uninstalled successfully');
        } else {
            console.error('❌ Uninstallation failed');
            process.exit(1);
        }
    });

// Update Check Command
program
    .command('check-updates')
    .description('Check for server updates')
    .action(async () => {
        console.log('🔄 Checking for updates...');
        const updates = await serverRegistry.checkUpdates();
        const withUpdates = updates.filter(u => u.hasUpdates);

        console.log(`✅ Checked ${updates.length} servers`);
        console.log(`   🔄 ${withUpdates.length} servers have updates available`);

        if (program.opts().format === 'table') {
            console.table(withUpdates.map(u => ({
                Server: u.serverName,
                Status: u.hasUpdates ? 'Update Available' : 'Up to Date',
                Latest: u.commitHash || '-'
            })));
        }
    });

// Update Server Command
program
    .command('update <serverId>')
    .description('Update a specific server')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`🔄 Updating server: ${serverId}`);
        const result = await serverRegistry.updateServer(serverId);

        if (result.success) {
            console.log(`✅ Updated successfully`);
        } else {
            console.error(`❌ Update failed: ${result.error}`);
            process.exit(1);
        }
    });

// Health Check Command
program
    .command('health <serverId>')
    .description('Check server health')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`💓 Checking health: ${serverId}`);
        const health = await serverRegistry.checkServerHealth(serverId);

        console.log(`✅ Status: ${health.status}`);
        if (health.latencyMs !== undefined) {
            console.log(`   Latency: ${health.latencyMs}ms`);
        }
        if (health.error) {
            console.log(`   Error: ${health.error}`);
        }
    });

// Detect Configs Command
program
    .command('detect-configs')
    .description('Auto-detect MCP configurations')
    .action(async () => {
        console.log('🔍 Auto-detecting configurations...');
        const configs = await configService.detectConfigs();

        console.log(`✅ Detected ${configs.length} configuration files`);
        if (program.opts().format === 'table') {
            console.table(configs.map(c => ({
                File: c.path,
                Format: c.format,
                Servers: c.servers.length
            })));
        } else {
            console.log(JSON.stringify(configs, null, 2));
        }
    });

// Import Configs Command
program
    .command('import-configs <files...>')
    .description('Import MCP configurations')
    .action(async (options, command) => {
        const files = command.args;
        console.log(`📝 Importing ${files.length} configurations...`);
        const result = await configService.importConfigs(files);

        if (result.success) {
            console.log(`✅ Imported ${result.imported} servers`);
        } else {
            console.error(`❌ Import failed: ${result.errors.length} errors`);
            result.errors.forEach(err => console.log(`   - ${err}`));
            process.exit(1);
        }
    });

// Export Configs Command
program
    .command('export-configs <format>')
    .description('Export configurations')
    .option('--format <type>', 'Export format (tormentnexus, claude, openai, google)', 'tormentnexus')
    .action(async (options, command) => {
        const requestedFormat = command.args[0] || program.opts().format || 'tormentnexus';
        const format = requestedFormat;
        console.log(`📤 Exporting to ${requestedFormat} format...`);

        try {
            const content = await configService.exportConfigs(format);
            console.log('✅ Export successful');
            console.log(content);
        } catch (e: any) {
            console.error(`❌ Export failed: ${e.message}`);
            process.exit(1);
        }
    });

// Init Sessions Command
program
    .command('init-sessions')
    .description('Initialize session service (auto-start all servers)')
    .action(async () => {
        console.log('🚀 Initializing session service...');
        await sessionService.initialize();

        const stats = sessionService.getStats();
        console.log(`✅ Initialized ${stats.totalSessions} sessions`);
        console.log(`   Running: ${stats.running}`);
        console.log(`   Total clients: ${stats.totalClients}`);
    });

// Session Stats Command
program
    .command('session-stats')
    .description('Get session statistics')
    .action(async () => {
        const stats = sessionService.getSessionStats();
        console.log('📊 Session Statistics:');
        console.log(`   Total sessions: ${stats.totalSessions}`);
        console.log(`   Running: ${stats.running}`);
        console.log(`   Stopped: ${stats.stopped}`);
        console.log(`   Error: ${stats.error}`);
        console.log(`   Total clients: ${stats.totalClients}`);
    });

// Session Metrics Command
program
    .command('session-metrics <serverName>')
    .description('Get performance metrics for a server')
    .action(async (options, command) => {
        const serverName = command.args[0];
        const metrics = sessionService.getMetrics(serverName);

        if (metrics) {
            console.log(`📊 Performance Metrics for ${serverName}:`);
            console.log(`   Average latency: ${metrics.avgLatencyMs.toFixed(2)}ms`);
            console.log(`   Min latency: ${metrics.minLatencyMs}ms`);
            console.log(`   Max latency: ${metrics.maxLatencyMs}ms`);
            console.log(`   Request count: ${metrics.requestCount}`);
            console.log(`   Error rate: ${(metrics.errorRate * 100).toFixed(2)}%`);
            console.log(`   Uptime: ${(metrics.uptimeMs / 1000).toFixed(1)}s`);
        } else {
            console.log(`⚠️ No metrics found for ${serverName}`);
        }
    });

// List Sessions Command
program
    .command('list-sessions')
    .description('List all sessions')
    .action(async () => {
        const sessions = sessionService.getAllSessions();

        if (program.opts().format === 'table') {
            console.table(sessions.map(s => ({
                ID: s.sessionId.substring(0, 20),
                Name: s.serverName,
                Status: s.status,
                Clients: s.clients.length,
                AutoRestart: s.autoRestart ? '✅' : '❌'
            })));
        } else {
            console.log(JSON.stringify(sessions, null, 2));
        }
    });

// Start Session Command
program
    .command('start-session <serverId>')
    .description('Start a server session')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`▶️ Starting session: ${serverId}`);
        await sessionService.startSession(serverId);
        console.log('✅ Session started');
    });

// Stop Session Command
program
    .command('stop-session <serverId>')
    .description('Stop a server session')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`⏹️ Stopping session: ${serverId}`);
        const session = sessionService.stopSession(serverId);

        if (session) {
            console.log('✅ Session stopped');
        } else {
            console.error('❌ Session not found');
            process.exit(1);
        }
    });

// Restart Session Command
program
    .command('restart-session <serverId>')
    .description('Restart a server session')
    .action(async (options, command) => {
        const serverId = command.args[0];
        console.log(`🔄 Restarting session: ${serverId}`);
        await sessionService.stopSession(serverId);
        await sessionService.startSession(serverId);
        console.log('✅ Session restarted');
    });

// Shutdown All Sessions Command
program
    .command('shutdown-sessions')
    .description('Shutdown all sessions')
    .action(async () => {
        console.log('🛑 Shutting down all sessions...');
        await sessionService.shutdown();
        console.log('✅ All sessions shut down');
    });

// ============================================
// Parse and Execute
// ============================================

program.parseAsync().then(async () => {
    console.log('✅ Ready to manage MCP servers!');
    console.log('');
    console.log('Available commands:');
    console.log('  Registry Management:');
    console.log('    discover        - Discover servers from registries');
    console.log('    search          - Search discovered servers');
    console.log('    categories       - List all categories');
    console.log('    stats           - Get registry statistics');
    console.log('');
    console.log('  Server Management:');
    console.log('    install          - Install server from registry');
    console.log('    uninstall        - Uninstall server');
    console.log('    check-updates    - Check for updates');
    console.log('    update           - Update a server');
    console.log('    health          - Check server health');
    console.log('');
    console.log('  Configuration Management:');
    console.log('    detect-configs   - Auto-detect configurations');
    console.log('    import-configs   - Import configurations');
    console.log('    export-configs   - Export configurations');
    console.log('');
    console.log('  Session Management:');
    console.log('    init-sessions    - Initialize session service');
    console.log('    session-stats    - Get session statistics');
    console.log('    session-metrics - Get performance metrics');
    console.log('    list-sessions    - List all sessions');
    console.log('    start-session    - Start a server session');
    console.log('    stop-session     - Stop a server session');
    console.log('    restart-session   - Restart a server session');
    console.log('    shutdown-sessions - Shutdown all sessions');
    console.log('');
    console.log('Examples:');
    console.log('  tormentnexus-mcp-router discover          - Discover all servers');
    console.log('  tormentnexus-mcp-router search "file"    - Search for file servers');
    console.log('  tormentnexus-mcp-router install fs-server   - Install filesystem server');
    console.log('  tormentnexus-mcp-router init-sessions       - Auto-start all servers');
    console.log('  tormentnexus-mcp-router session-stats       - Get session statistics');
}).catch(err => {
    console.error('❌ Error:', err.message);
    process.exit(1);
});
