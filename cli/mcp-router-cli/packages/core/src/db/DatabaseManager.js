import { EventEmitter } from 'events';
import Database from 'better-sqlite3';
import fs from 'fs';
import path from 'path';
import crypto from 'crypto';
import { SQL_SCHEMA, generateId, fromDbRow } from './schema.js';
export class DatabaseManager extends EventEmitter {
    db;
    static instance = null;
    constructor(dataDir) {
        super();
        const dir = dataDir || path.join(process.cwd(), 'data');
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }
        const tormentnexusDbPath = path.join(dir, 'tormentnexus.db');
        const legacyDbPath = path.join(dir, 'legacy_tormentnexus.db');
        const dbPath = fs.existsSync(tormentnexusDbPath) ? tormentnexusDbPath : (fs.existsSync(legacyDbPath) ? legacyDbPath : tormentnexusDbPath);
        this.db = new Database(dbPath);
        // Enable foreign keys
        this.db.pragma('foreign_keys = ON');
        // Initialize schema
        this.db.exec(SQL_SCHEMA);
        this.migrateSchema();
        console.log(`[DatabaseManager] Initialized at ${dbPath}`);
    }
    migrateSchema() {
        const hasColumn = (table, column) => {
            const rows = this.db.prepare(`PRAGMA table_info(${table})`).all();
            return rows.some(r => r.name === column);
        };
        const ensureColumn = (table, column, ddl) => {
            if (hasColumn(table, column))
                return;
            this.db.exec(`ALTER TABLE ${table} ADD COLUMN ${ddl}`);
        };
        ensureColumn('namespace_tool_mappings', 'override_description', 'override_description TEXT');
        ensureColumn('namespace_tool_mappings', 'override_annotations', 'override_annotations TEXT');
        this.db.exec('CREATE INDEX IF NOT EXISTS idx_ns_tool_map_alias ON namespace_tool_mappings(alias_name)');
        this.db.exec(`
            CREATE TABLE IF NOT EXISTS tool_annotations (
                id TEXT PRIMARY KEY,
                namespace_id TEXT,
                mcp_server_name TEXT NOT NULL,
                tool_name TEXT NOT NULL,
                data TEXT NOT NULL,
                updated_at INTEGER NOT NULL,
                created_at INTEGER NOT NULL,
                UNIQUE(namespace_id, mcp_server_name, tool_name)
            );
        `);
        this.db.exec('CREATE INDEX IF NOT EXISTS idx_tool_annotations_namespace ON tool_annotations(namespace_id)');
        this.db.exec('CREATE INDEX IF NOT EXISTS idx_tool_annotations_server_tool ON tool_annotations(mcp_server_name, tool_name)');
    }
    static getInstance(dataDir) {
        if (!DatabaseManager.instance) {
            DatabaseManager.instance = new DatabaseManager(dataDir);
        }
        return DatabaseManager.instance;
    }
    // ============================================
    // MCP Servers
    // ============================================
    createMcpServer(server) {
        const now = Date.now();
        const fullServer = {
            id: generateId(),
            ...server,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO mcp_servers (id, name, type, command, args, env, url, bearer_token, headers, description, icon, enabled, created_at, updated_at)
            VALUES (@id, @name, @type, @command, @args, @env, @url, @bearer_token, @headers, @description, @icon, @enabled, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullServer.id,
            name: fullServer.name,
            type: fullServer.type,
            command: fullServer.command || null,
            args: fullServer.args ? JSON.stringify(fullServer.args) : null,
            env: fullServer.env ? JSON.stringify(fullServer.env) : null,
            url: fullServer.url || null,
            bearer_token: fullServer.bearerToken || null,
            headers: fullServer.headers ? JSON.stringify(fullServer.headers) : null,
            description: fullServer.description || null,
            icon: fullServer.icon || null,
            enabled: fullServer.enabled ? 1 : 0,
            created_at: fullServer.createdAt,
            updated_at: fullServer.updatedAt
        });
        this.emit('server:created', fullServer);
        return fullServer;
    }
    getMcpServer(id) {
        const stmt = this.db.prepare('SELECT * FROM mcp_servers WHERE id = ?');
        const row = stmt.get(id);
        if (!row)
            return null;
        return fromDbRow(row, ['args', 'env', 'headers']);
    }
    getMcpServerByName(name) {
        const stmt = this.db.prepare('SELECT * FROM mcp_servers WHERE name = ?');
        const row = stmt.get(name);
        if (!row)
            return null;
        return fromDbRow(row, ['args', 'env', 'headers']);
    }
    getAllMcpServers(filter) {
        let query = 'SELECT * FROM mcp_servers WHERE 1=1';
        const params = [];
        if (filter?.type) {
            query += ' AND type = ?';
            params.push(filter.type);
        }
        if (filter?.enabled !== undefined) {
            query += ' AND enabled = ?';
            params.push(filter.enabled ? 1 : 0);
        }
        query += ' ORDER BY name ASC';
        const stmt = this.db.prepare(query);
        const rows = stmt.all(...params);
        return rows.map(row => fromDbRow(row, ['args', 'env', 'headers']));
    }
    updateMcpServer(id, updates) {
        const existing = this.getMcpServer(id);
        if (!existing)
            return null;
        const updated = { ...existing, ...updates, updatedAt: Date.now() };
        const stmt = this.db.prepare(`
            UPDATE mcp_servers SET
                name = @name, type = @type, command = @command, args = @args, env = @env,
                url = @url, bearer_token = @bearer_token, headers = @headers,
                description = @description, icon = @icon, enabled = @enabled, updated_at = @updated_at
            WHERE id = @id
        `);
        stmt.run({
            id: updated.id,
            name: updated.name,
            type: updated.type,
            command: updated.command || null,
            args: updated.args ? JSON.stringify(updated.args) : null,
            env: updated.env ? JSON.stringify(updated.env) : null,
            url: updated.url || null,
            bearer_token: updated.bearerToken || null,
            headers: updated.headers ? JSON.stringify(updated.headers) : null,
            description: updated.description || null,
            icon: updated.icon || null,
            enabled: updated.enabled ? 1 : 0,
            updated_at: updated.updatedAt
        });
        this.emit('server:updated', updated);
        return updated;
    }
    deleteMcpServer(id) {
        const stmt = this.db.prepare('DELETE FROM mcp_servers WHERE id = ?');
        const result = stmt.run(id);
        if (result.changes > 0) {
            this.emit('server:deleted', id);
            return true;
        }
        return false;
    }
    // ============================================
    // Tools
    // ============================================
    createTool(tool) {
        const now = Date.now();
        const fullTool = {
            id: generateId(),
            ...tool,
            usageCount: 0,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO tools (id, name, description, input_schema, mcp_server_id, embedding, category, tags, usage_count, last_used_at, created_at, updated_at)
            VALUES (@id, @name, @description, @input_schema, @mcp_server_id, @embedding, @category, @tags, @usage_count, @last_used_at, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullTool.id,
            name: fullTool.name,
            description: fullTool.description || null,
            input_schema: fullTool.inputSchema ? JSON.stringify(fullTool.inputSchema) : null,
            mcp_server_id: fullTool.mcpServerId,
            embedding: fullTool.embedding ? JSON.stringify(fullTool.embedding) : null,
            category: fullTool.category || null,
            tags: fullTool.tags ? JSON.stringify(fullTool.tags) : null,
            usage_count: fullTool.usageCount,
            last_used_at: fullTool.lastUsedAt || null,
            created_at: fullTool.createdAt,
            updated_at: fullTool.updatedAt
        });
        this.emit('tool:created', fullTool);
        return fullTool;
    }
    getTool(id) {
        const stmt = this.db.prepare('SELECT * FROM tools WHERE id = ?');
        const row = stmt.get(id);
        if (!row)
            return null;
        return fromDbRow(row, ['inputSchema', 'embedding', 'tags']);
    }
    getToolByName(name, mcpServerId) {
        let query = 'SELECT * FROM tools WHERE name = ?';
        const params = [name];
        if (mcpServerId) {
            query += ' AND mcp_server_id = ?';
            params.push(mcpServerId);
        }
        const stmt = this.db.prepare(query);
        const row = stmt.get(...params);
        if (!row)
            return null;
        return fromDbRow(row, ['inputSchema', 'embedding', 'tags']);
    }
    getToolsByServer(mcpServerId) {
        const stmt = this.db.prepare('SELECT * FROM tools WHERE mcp_server_id = ? ORDER BY name ASC');
        const rows = stmt.all(mcpServerId);
        return rows.map(row => fromDbRow(row, ['inputSchema', 'embedding', 'tags']));
    }
    getAllTools(filter) {
        let query = 'SELECT * FROM tools WHERE 1=1';
        const params = [];
        if (filter?.category) {
            query += ' AND category = ?';
            params.push(filter.category);
        }
        if (filter?.mcpServerId) {
            query += ' AND mcp_server_id = ?';
            params.push(filter.mcpServerId);
        }
        query += ' ORDER BY usage_count DESC, name ASC';
        const stmt = this.db.prepare(query);
        const rows = stmt.all(...params);
        return rows.map(row => fromDbRow(row, ['inputSchema', 'embedding', 'tags']));
    }
    updateTool(id, updates) {
        const existing = this.getTool(id);
        if (!existing)
            return null;
        const updated = { ...existing, ...updates, updatedAt: Date.now() };
        const stmt = this.db.prepare(`
            UPDATE tools SET
                name = @name, description = @description, input_schema = @input_schema,
                mcp_server_id = @mcp_server_id, embedding = @embedding, category = @category,
                tags = @tags, usage_count = @usage_count, last_used_at = @last_used_at, updated_at = @updated_at
            WHERE id = @id
        `);
        stmt.run({
            id: updated.id,
            name: updated.name,
            description: updated.description || null,
            input_schema: updated.inputSchema ? JSON.stringify(updated.inputSchema) : null,
            mcp_server_id: updated.mcpServerId,
            embedding: updated.embedding ? JSON.stringify(updated.embedding) : null,
            category: updated.category || null,
            tags: updated.tags ? JSON.stringify(updated.tags) : null,
            usage_count: updated.usageCount,
            last_used_at: updated.lastUsedAt || null,
            updated_at: updated.updatedAt
        });
        this.emit('tool:updated', updated);
        return updated;
    }
    incrementToolUsage(id) {
        const stmt = this.db.prepare(`
            UPDATE tools SET usage_count = usage_count + 1, last_used_at = ?, updated_at = ? WHERE id = ?
        `);
        const now = Date.now();
        stmt.run(now, now, id);
    }
    deleteTool(id) {
        const stmt = this.db.prepare('DELETE FROM tools WHERE id = ?');
        const result = stmt.run(id);
        if (result.changes > 0) {
            this.emit('tool:deleted', id);
            return true;
        }
        return false;
    }
    // ============================================
    // Namespaces
    // ============================================
    createNamespace(ns) {
        const now = Date.now();
        const fullNs = {
            id: generateId(),
            ...ns,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO namespaces (id, name, description, icon, is_default, created_at, updated_at)
            VALUES (@id, @name, @description, @icon, @is_default, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullNs.id,
            name: fullNs.name,
            description: fullNs.description || null,
            icon: fullNs.icon || null,
            is_default: fullNs.isDefault ? 1 : 0,
            created_at: fullNs.createdAt,
            updated_at: fullNs.updatedAt
        });
        this.emit('namespace:created', fullNs);
        return fullNs;
    }
    getNamespace(id) {
        const stmt = this.db.prepare('SELECT * FROM namespaces WHERE id = ?');
        const row = stmt.get(id);
        if (!row)
            return null;
        return fromDbRow(row, []);
    }
    getDefaultNamespace() {
        const stmt = this.db.prepare('SELECT * FROM namespaces WHERE is_default = 1 LIMIT 1');
        const row = stmt.get();
        if (!row)
            return null;
        return fromDbRow(row, []);
    }
    getAllNamespaces() {
        const stmt = this.db.prepare('SELECT * FROM namespaces ORDER BY is_default DESC, name ASC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, []));
    }
    updateNamespace(id, updates) {
        const existing = this.getNamespace(id);
        if (!existing)
            return null;
        const updated = { ...existing, ...updates, updatedAt: Date.now() };
        const stmt = this.db.prepare(`
            UPDATE namespaces SET name = @name, description = @description, icon = @icon, is_default = @is_default, updated_at = @updated_at
            WHERE id = @id
        `);
        stmt.run({
            id: updated.id,
            name: updated.name,
            description: updated.description || null,
            icon: updated.icon || null,
            is_default: updated.isDefault ? 1 : 0,
            updated_at: updated.updatedAt
        });
        this.emit('namespace:updated', updated);
        return updated;
    }
    deleteNamespace(id) {
        // Don't delete the default namespace
        const ns = this.getNamespace(id);
        if (!ns || ns.isDefault)
            return false;
        const stmt = this.db.prepare('DELETE FROM namespaces WHERE id = ?');
        const result = stmt.run(id);
        if (result.changes > 0) {
            this.emit('namespace:deleted', id);
            return true;
        }
        return false;
    }
    // ============================================
    // Namespace Server Mappings
    // ============================================
    addServerToNamespace(namespaceId, mcpServerId, priority = 0) {
        const mapping = {
            id: generateId(),
            namespaceId,
            mcpServerId,
            priority,
            createdAt: Date.now()
        };
        const stmt = this.db.prepare(`
            INSERT INTO namespace_server_mappings (id, namespace_id, mcp_server_id, priority, created_at)
            VALUES (@id, @namespace_id, @mcp_server_id, @priority, @created_at)
        `);
        stmt.run({
            id: mapping.id,
            namespace_id: mapping.namespaceId,
            mcp_server_id: mapping.mcpServerId,
            priority: mapping.priority,
            created_at: mapping.createdAt
        });
        return mapping;
    }
    removeServerFromNamespace(namespaceId, mcpServerId) {
        const stmt = this.db.prepare('DELETE FROM namespace_server_mappings WHERE namespace_id = ? AND mcp_server_id = ?');
        const result = stmt.run(namespaceId, mcpServerId);
        return result.changes > 0;
    }
    getServersInNamespace(namespaceId) {
        const stmt = this.db.prepare(`
            SELECT s.* FROM mcp_servers s
            JOIN namespace_server_mappings m ON s.id = m.mcp_server_id
            WHERE m.namespace_id = ?
            ORDER BY m.priority ASC, s.name ASC
        `);
        const rows = stmt.all(namespaceId);
        return rows.map(row => fromDbRow(row, ['args', 'env', 'headers']));
    }
    // ============================================
    // Endpoints
    // ============================================
    createEndpoint(endpoint) {
        const now = Date.now();
        const fullEndpoint = {
            id: generateId(),
            ...endpoint,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO endpoints (id, name, path, namespace_id, api_key_id, enabled, created_at, updated_at)
            VALUES (@id, @name, @path, @namespace_id, @api_key_id, @enabled, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullEndpoint.id,
            name: fullEndpoint.name,
            path: fullEndpoint.path,
            namespace_id: fullEndpoint.namespaceId,
            api_key_id: fullEndpoint.apiKeyId || null,
            enabled: fullEndpoint.enabled ? 1 : 0,
            created_at: fullEndpoint.createdAt,
            updated_at: fullEndpoint.updatedAt
        });
        this.emit('endpoint:created', fullEndpoint);
        return fullEndpoint;
    }
    getEndpointByPath(path) {
        const stmt = this.db.prepare('SELECT * FROM endpoints WHERE path = ?');
        const row = stmt.get(path);
        if (!row)
            return null;
        return fromDbRow(row, []);
    }
    getAllEndpoints() {
        const stmt = this.db.prepare('SELECT * FROM endpoints ORDER BY path ASC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, []));
    }
    deleteEndpoint(id) {
        const stmt = this.db.prepare('DELETE FROM endpoints WHERE id = ?');
        const result = stmt.run(id);
        return result.changes > 0;
    }
    // ============================================
    // API Keys
    // ============================================
    createApiKey(name, scopes = ['read']) {
        const plainKey = `tormentnexus_${crypto.randomBytes(32).toString('hex')}`;
        const keyHash = crypto.createHash('sha256').update(plainKey).digest('hex');
        const keyPrefix = plainKey.substring(0, 12);
        const apiKey = {
            id: generateId(),
            name,
            keyHash,
            keyPrefix,
            scopes,
            createdAt: Date.now()
        };
        const stmt = this.db.prepare(`
            INSERT INTO api_keys (id, name, key_hash, key_prefix, scopes, expires_at, last_used_at, created_at)
            VALUES (@id, @name, @key_hash, @key_prefix, @scopes, @expires_at, @last_used_at, @created_at)
        `);
        stmt.run({
            id: apiKey.id,
            name: apiKey.name,
            key_hash: apiKey.keyHash,
            key_prefix: apiKey.keyPrefix,
            scopes: JSON.stringify(apiKey.scopes),
            expires_at: apiKey.expiresAt || null,
            last_used_at: apiKey.lastUsedAt || null,
            created_at: apiKey.createdAt
        });
        return { apiKey, plainKey };
    }
    validateApiKey(plainKey) {
        const keyHash = crypto.createHash('sha256').update(plainKey).digest('hex');
        const stmt = this.db.prepare('SELECT * FROM api_keys WHERE key_hash = ?');
        const row = stmt.get(keyHash);
        if (!row)
            return null;
        const apiKey = fromDbRow(row, ['scopes']);
        // Check expiration
        if (apiKey.expiresAt && apiKey.expiresAt < Date.now()) {
            return null;
        }
        // Update last used
        const updateStmt = this.db.prepare('UPDATE api_keys SET last_used_at = ? WHERE id = ?');
        updateStmt.run(Date.now(), apiKey.id);
        return apiKey;
    }
    getAllApiKeys() {
        const stmt = this.db.prepare('SELECT * FROM api_keys ORDER BY created_at DESC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, ['scopes']));
    }
    deleteApiKey(id) {
        const stmt = this.db.prepare('DELETE FROM api_keys WHERE id = ?');
        const result = stmt.run(id);
        return result.changes > 0;
    }
    // ============================================
    // Policies
    // ============================================
    createPolicy(policy) {
        const now = Date.now();
        const fullPolicy = {
            id: generateId(),
            ...policy,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO policies (id, name, description, priority, rules, enabled, created_at, updated_at)
            VALUES (@id, @name, @description, @priority, @rules, @enabled, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullPolicy.id,
            name: fullPolicy.name,
            description: fullPolicy.description || null,
            priority: fullPolicy.priority,
            rules: JSON.stringify(fullPolicy.rules),
            enabled: fullPolicy.enabled ? 1 : 0,
            created_at: fullPolicy.createdAt,
            updated_at: fullPolicy.updatedAt
        });
        this.emit('policy:created', fullPolicy);
        return fullPolicy;
    }
    getAllPolicies() {
        const stmt = this.db.prepare('SELECT * FROM policies WHERE enabled = 1 ORDER BY priority ASC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, ['rules']));
    }
    evaluatePolicy(toolName, serverName) {
        const policies = this.getAllPolicies();
        for (const policy of policies) {
            for (const rule of policy.rules) {
                const pattern = rule.pattern.replace(/\*/g, '.*');
                const regex = new RegExp(`^${pattern}$`, 'i');
                const target = serverName ? `${serverName}/${toolName}` : toolName;
                if (regex.test(target) || regex.test(toolName)) {
                    return {
                        allowed: rule.action === 'allow',
                        matchedRule: rule
                    };
                }
            }
        }
        // Default: allow if no policy matches
        return { allowed: true };
    }
    deletePolicy(id) {
        const stmt = this.db.prepare('DELETE FROM policies WHERE id = ?');
        const result = stmt.run(id);
        return result.changes > 0;
    }
    // ============================================
    // Tool Call Logs
    // ============================================
    logToolCall(log) {
        const fullLog = {
            id: generateId(),
            ...log
        };
        const stmt = this.db.prepare(`
            INSERT INTO tool_call_logs (id, timestamp, tool_name, mcp_server_id, namespace_id, endpoint_id, api_key_id, request_args, response_data, error_message, duration_ms, success)
            VALUES (@id, @timestamp, @tool_name, @mcp_server_id, @namespace_id, @endpoint_id, @api_key_id, @request_args, @response_data, @error_message, @duration_ms, @success)
        `);
        stmt.run({
            id: fullLog.id,
            timestamp: fullLog.timestamp,
            tool_name: fullLog.toolName,
            mcp_server_id: fullLog.mcpServerId || null,
            namespace_id: fullLog.namespaceId || null,
            endpoint_id: fullLog.endpointId || null,
            api_key_id: fullLog.apiKeyId || null,
            request_args: fullLog.requestArgs ? JSON.stringify(fullLog.requestArgs) : null,
            response_data: fullLog.responseData ? JSON.stringify(fullLog.responseData) : null,
            error_message: fullLog.errorMessage || null,
            duration_ms: fullLog.durationMs,
            success: fullLog.success ? 1 : 0
        });
        this.emit('toolcall:logged', fullLog);
        return fullLog;
    }
    getToolCallLogs(filter) {
        let query = 'SELECT * FROM tool_call_logs WHERE 1=1';
        const params = [];
        if (filter?.toolName) {
            query += ' AND tool_name = ?';
            params.push(filter.toolName);
        }
        if (filter?.mcpServerId) {
            query += ' AND mcp_server_id = ?';
            params.push(filter.mcpServerId);
        }
        if (filter?.success !== undefined) {
            query += ' AND success = ?';
            params.push(filter.success ? 1 : 0);
        }
        if (filter?.since) {
            query += ' AND timestamp >= ?';
            params.push(filter.since);
        }
        query += ' ORDER BY timestamp DESC';
        if (filter?.limit) {
            query += ' LIMIT ?';
            params.push(filter.limit);
        }
        const stmt = this.db.prepare(query);
        const rows = stmt.all(...params);
        return rows.map(row => fromDbRow(row, ['requestArgs', 'responseData']));
    }
    getNamespaceToolOverridesByToolName(namespaceId) {
        const stmt = this.db.prepare(`
            SELECT
                t.name as tool_name,
                m.status as status,
                m.alias_name as alias_name,
                m.override_description as override_description,
                m.override_annotations as override_annotations
            FROM namespace_tool_mappings m
            JOIN tools t ON t.id = m.tool_id
            WHERE m.namespace_id = ?
        `);
        const rows = stmt.all(namespaceId);
        return rows.map(r => ({
            toolName: String(r.tool_name),
            status: r.status,
            aliasName: r.alias_name ? String(r.alias_name) : undefined,
            overrideDescription: r.override_description ? String(r.override_description) : undefined,
            overrideAnnotations: r.override_annotations ? JSON.parse(String(r.override_annotations)) : undefined,
        }));
    }
    upsertNamespaceToolOverrideByToolName(namespaceId, toolName, status, aliasName, overrideDescription, overrideAnnotations) {
        const tool = this.getToolByName(toolName);
        if (!tool) {
            throw new Error(`Tool not found: ${toolName}`);
        }
        const existing = this.db.prepare('SELECT * FROM namespace_tool_mappings WHERE namespace_id = ? AND tool_id = ?').get(namespaceId, tool.id);
        const now = Date.now();
        if (!existing) {
            const stmt = this.db.prepare(`
                INSERT INTO namespace_tool_mappings (id, namespace_id, tool_id, status, alias_name, override_description, override_annotations, created_at, updated_at)
                VALUES (@id, @namespace_id, @tool_id, @status, @alias_name, @override_description, @override_annotations, @created_at, @updated_at)
            `);
            stmt.run({
                id: generateId(),
                namespace_id: namespaceId,
                tool_id: tool.id,
                status,
                alias_name: aliasName || null,
                override_description: overrideDescription || null,
                override_annotations: overrideAnnotations ? JSON.stringify(overrideAnnotations) : null,
                created_at: now,
                updated_at: now
            });
        }
        else {
            const stmt = this.db.prepare(`
                UPDATE namespace_tool_mappings
                SET status = @status, alias_name = @alias_name, override_description = @override_description, override_annotations = @override_annotations, updated_at = @updated_at
                WHERE namespace_id = @namespace_id AND tool_id = @tool_id
            `);
            stmt.run({
                namespace_id: namespaceId,
                tool_id: tool.id,
                status,
                alias_name: aliasName || null,
                override_description: overrideDescription || null,
                override_annotations: overrideAnnotations ? JSON.stringify(overrideAnnotations) : null,
                updated_at: now
            });
        }
        return { toolName, status, aliasName, overrideDescription, overrideAnnotations };
    }
    deleteNamespaceToolOverrideByToolName(namespaceId, toolName) {
        const tool = this.getToolByName(toolName);
        if (!tool)
            return false;
        const stmt = this.db.prepare('DELETE FROM namespace_tool_mappings WHERE namespace_id = ? AND tool_id = ?');
        const result = stmt.run(namespaceId, tool.id);
        return result.changes > 0;
    }
    upsertToolAnnotation(params) {
        const now = Date.now();
        const existing = this.db.prepare('SELECT id FROM tool_annotations WHERE namespace_id IS ? AND mcp_server_name = ? AND tool_name = ?').get(params.namespaceId ?? null, params.mcpServerName, params.toolName);
        if (!existing) {
            const id = generateId();
            this.db.prepare('INSERT INTO tool_annotations (id, namespace_id, mcp_server_name, tool_name, data, updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)').run(id, params.namespaceId ?? null, params.mcpServerName, params.toolName, JSON.stringify(params.data), now, now);
            return { id };
        }
        this.db.prepare('UPDATE tool_annotations SET data = ?, updated_at = ? WHERE id = ?').run(JSON.stringify(params.data), now, existing.id);
        return { id: existing.id };
    }
    getToolAnnotation(params) {
        const row = this.db.prepare('SELECT data FROM tool_annotations WHERE namespace_id IS ? AND mcp_server_name = ? AND tool_name = ?').get(params.namespaceId ?? null, params.mcpServerName, params.toolName);
        if (!row?.data)
            return null;
        try {
            return JSON.parse(row.data);
        }
        catch {
            return null;
        }
    }
    deleteToolAnnotation(params) {
        const res = this.db.prepare('DELETE FROM tool_annotations WHERE namespace_id IS ? AND mcp_server_name = ? AND tool_name = ?').run(params.namespaceId ?? null, params.mcpServerName, params.toolName);
        return res.changes > 0;
    }
    // ============================================
    // Saved Scripts
    // ============================================
    createSavedScript(script) {
        const now = Date.now();
        const fullScript = {
            id: generateId(),
            ...script,
            runCount: 0,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO saved_scripts (id, name, description, language, code, tags, is_favorite, run_count, last_run_at, created_at, updated_at)
            VALUES (@id, @name, @description, @language, @code, @tags, @is_favorite, @run_count, @last_run_at, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullScript.id,
            name: fullScript.name,
            description: fullScript.description || null,
            language: fullScript.language,
            code: fullScript.code,
            tags: fullScript.tags ? JSON.stringify(fullScript.tags) : null,
            is_favorite: fullScript.isFavorite ? 1 : 0,
            run_count: fullScript.runCount,
            last_run_at: fullScript.lastRunAt || null,
            created_at: fullScript.createdAt,
            updated_at: fullScript.updatedAt
        });
        return fullScript;
    }
    getAllSavedScripts() {
        const stmt = this.db.prepare('SELECT * FROM saved_scripts ORDER BY is_favorite DESC, updated_at DESC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, ['tags']));
    }
    incrementScriptRunCount(id) {
        const now = Date.now();
        const stmt = this.db.prepare('UPDATE saved_scripts SET run_count = run_count + 1, last_run_at = ?, updated_at = ? WHERE id = ?');
        stmt.run(now, now, id);
    }
    deleteSavedScript(id) {
        const stmt = this.db.prepare('DELETE FROM saved_scripts WHERE id = ?');
        const result = stmt.run(id);
        return result.changes > 0;
    }
    // ============================================
    // Tool Sets
    // ============================================
    createToolSet(toolSet) {
        const now = Date.now();
        const fullToolSet = {
            id: generateId(),
            ...toolSet,
            createdAt: now,
            updatedAt: now
        };
        const stmt = this.db.prepare(`
            INSERT INTO tool_sets (id, name, description, tool_ids, icon, is_public, created_at, updated_at)
            VALUES (@id, @name, @description, @tool_ids, @icon, @is_public, @created_at, @updated_at)
        `);
        stmt.run({
            id: fullToolSet.id,
            name: fullToolSet.name,
            description: fullToolSet.description || null,
            tool_ids: JSON.stringify(fullToolSet.toolIds),
            icon: fullToolSet.icon || null,
            is_public: fullToolSet.isPublic ? 1 : 0,
            created_at: fullToolSet.createdAt,
            updated_at: fullToolSet.updatedAt
        });
        return fullToolSet;
    }
    getToolSet(id) {
        const stmt = this.db.prepare('SELECT * FROM tool_sets WHERE id = ?');
        const row = stmt.get(id);
        if (!row)
            return null;
        return fromDbRow(row, ['toolIds']);
    }
    getAllToolSets() {
        const stmt = this.db.prepare('SELECT * FROM tool_sets ORDER BY name ASC');
        const rows = stmt.all();
        return rows.map(row => fromDbRow(row, ['toolIds']));
    }
    updateToolSet(id, updates) {
        const existing = this.getToolSet(id);
        if (!existing)
            return null;
        const updated = {
            ...existing,
            ...updates,
            updatedAt: Date.now()
        };
        const stmt = this.db.prepare(`
            UPDATE tool_sets SET
                name = @name,
                description = @description,
                tool_ids = @tool_ids,
                icon = @icon,
                is_public = @is_public,
                updated_at = @updated_at
            WHERE id = @id
        `);
        stmt.run({
            id: updated.id,
            name: updated.name,
            description: updated.description || null,
            tool_ids: JSON.stringify(updated.toolIds),
            icon: updated.icon || null,
            is_public: updated.isPublic ? 1 : 0,
            updated_at: updated.updatedAt
        });
        return updated;
    }
    deleteToolSet(id) {
        const stmt = this.db.prepare('DELETE FROM tool_sets WHERE id = ?');
        const result = stmt.run(id);
        return result.changes > 0;
    }
    // ============================================
    // Utility
    // ============================================
    close() {
        try {
            this.db.close();
            DatabaseManager.instance = null;
            console.log('[DatabaseManager] Closed');
        }
        catch (e) {
            console.error('[DatabaseManager] Failed to close:', e);
        }
    }
    getStats() {
        const counts = this.db.prepare(`
            SELECT
                (SELECT COUNT(*) FROM mcp_servers) as servers,
                (SELECT COUNT(*) FROM tools) as tools,
                (SELECT COUNT(*) FROM namespaces) as namespaces,
                (SELECT COUNT(*) FROM endpoints) as endpoints,
                (SELECT COUNT(*) FROM api_keys) as api_keys,
                (SELECT COUNT(*) FROM policies) as policies,
                (SELECT COUNT(*) FROM tool_call_logs) as tool_call_logs,
                (SELECT COUNT(*) FROM saved_scripts) as saved_scripts,
                (SELECT COUNT(*) FROM tool_sets) as tool_sets
        `).get();
        return {
            servers: counts.servers,
            tools: counts.tools,
            namespaces: counts.namespaces,
            endpoints: counts.endpoints,
            apiKeys: counts.api_keys,
            policies: counts.policies,
            toolCallLogs: counts.tool_call_logs,
            savedScripts: counts.saved_scripts,
            toolSets: counts.tool_sets
        };
    }
}
// Re-export types
export * from './schema.js';
