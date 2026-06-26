import * as fs from 'node:fs/promises';
import * as os from 'node:os';
import * as path from 'node:path';

import type { DatabaseMcpServer } from '../types/mcp-admin/index.js';

export const SUPPORTED_MCP_CLIENTS = ['claude-desktop', 'cursor', 'vscode'] as const;

export type SupportedMcpClient = (typeof SUPPORTED_MCP_CLIENTS)[number];

type JsonValue = null | boolean | number | string | JsonValue[] | { [key: string]: JsonValue };
type JsonObject = { [key: string]: JsonValue };

export type ServerRecord = Pick<
    DatabaseMcpServer,
    'name' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'
>;

interface FileSystemAdapter {
    readFile(filePath: string, encoding: BufferEncoding): Promise<string>;
    writeFile(filePath: string, content: string, encoding: BufferEncoding): Promise<void>;
    mkdir(dirPath: string, options: { recursive: true }): Promise<void>;
    access(filePath: string): Promise<void>;
}

interface ClientConfigSyncOptions {
    env?: NodeJS.ProcessEnv;
    platform?: NodeJS.Platform;
    cwd?: string;
    homedir?: () => string;
    fileSystem?: Partial<FileSystemAdapter>;
    loadServers?: () => Promise<ServerRecord[]>;
}

export interface ResolvedClientConfigTarget {
    client: SupportedMcpClient;
    path: string;
    candidates: string[];
    exists: boolean;
}

export interface ClientConfigPreview {
    client: SupportedMcpClient;
    targetPath: string;
    existed: boolean;
    serverCount: number;
    document: JsonObject;
    json: string;
}

export interface ClientConfigSyncResult extends ClientConfigPreview {
    written: boolean;
}

function toJsonObject(value: unknown): JsonObject {
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
        return {};
    }

    return value as JsonObject;
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return !!value && typeof value === 'object' && !Array.isArray(value);
}

export class ClientConfigSyncService {
    private readonly env: NodeJS.ProcessEnv;
    private readonly platform: NodeJS.Platform;
    private readonly cwd: string;
    private readonly homedir: () => string;
    private readonly fileSystem: FileSystemAdapter;
    private readonly loadServers: () => Promise<ServerRecord[]>;

    constructor(options: ClientConfigSyncOptions = {}) {
        this.env = options.env ?? process.env;
        this.platform = options.platform ?? process.platform;
        this.cwd = options.cwd ?? process.cwd();
        this.homedir = options.homedir ?? os.homedir;
        this.fileSystem = {
            readFile: options.fileSystem?.readFile ?? ((filePath, encoding) => fs.readFile(filePath, encoding)),
            writeFile: options.fileSystem?.writeFile ?? ((filePath, content, encoding) => fs.writeFile(filePath, content, encoding)),
            mkdir: options.fileSystem?.mkdir ?? (async (dirPath, mkdirOptions) => {
                await fs.mkdir(dirPath, mkdirOptions);
            }),
            access: options.fileSystem?.access ?? ((filePath) => fs.access(filePath)),
        };
        this.loadServers = options.loadServers ?? (async () => {
            const { mcpServersRepository } = await import('../db/repositories/mcp-servers.repo.js');
            return mcpServersRepository.findAll();
        });
    }

    async listTargets(): Promise<ResolvedClientConfigTarget[]> {
        const results: ResolvedClientConfigTarget[] = [];
        for (const client of SUPPORTED_MCP_CLIENTS) {
            results.push(await this.resolveTarget(client));
        }
        return results;
    }

    async previewSync(client: SupportedMcpClient, overridePath?: string): Promise<ClientConfigPreview> {
        const target = await this.resolveTarget(client, overridePath);
        const existingDocument = await this.readJsonObjectIfExists(target.path);
        const servers = await this.loadServers();
        const document = this.mergeClientConfig(existingDocument, this.buildMcpServersConfig(servers));
        const json = `${JSON.stringify(document, null, 2)}\n`;

        return {
            client,
            targetPath: target.path,
            existed: target.exists,
            serverCount: Object.keys(document.mcpServers as JsonObject).length,
            document,
            json,
        };
    }

    async syncClientConfig(client: SupportedMcpClient, overridePath?: string): Promise<ClientConfigSyncResult> {
        const preview = await this.previewSync(client, overridePath);
        await this.fileSystem.mkdir(path.dirname(preview.targetPath), { recursive: true });
        await this.fileSystem.writeFile(preview.targetPath, preview.json, 'utf-8');
        return {
            ...preview,
            written: true,
        };
    }

    async resolveTarget(client: SupportedMcpClient, overridePath?: string): Promise<ResolvedClientConfigTarget> {
        const candidates = overridePath ? [overridePath] : this.getClientCandidates(client);
        const existingPath = await this.findExistingPath(candidates);
        const resolvedPath = existingPath ?? candidates[0];

        return {
            client,
            path: resolvedPath,
            candidates,
            exists: existingPath !== null,
        };
    }

    private async findExistingPath(candidates: string[]): Promise<string | null> {
        for (const candidate of candidates) {
            if (await this.pathExists(candidate)) {
                return candidate;
            }
        }
        return null;
    }

    private async pathExists(filePath: string): Promise<boolean> {
        try {
            await this.fileSystem.access(filePath);
            return true;
        } catch {
            return false;
        }
    }

    private async readJsonObjectIfExists(filePath: string): Promise<JsonObject> {
        if (!(await this.pathExists(filePath))) {
            return {};
        }

        const content = await this.fileSystem.readFile(filePath, 'utf-8');
        if (!content.trim()) {
            return {};
        }

        const parsed = JSON.parse(content) as unknown;
        return toJsonObject(parsed);
    }

    private getClientCandidates(client: SupportedMcpClient): string[] {
        const homeDir = this.homedir();
        const appData = this.env.APPDATA ?? path.join(homeDir, 'AppData', 'Roaming');

        switch (client) {
            case 'claude-desktop':
                return this.byPlatform({
                    win32: [path.join(appData, 'Claude', 'claude_desktop_config.json')],
                    darwin: [path.join(homeDir, 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json')],
                    linux: [path.join(homeDir, '.config', 'Claude', 'claude_desktop_config.json')],
                });
            case 'cursor':
                return this.byPlatform({
                    win32: [
                        path.join(appData, 'Cursor', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(appData, 'Cursor', 'User', 'mcp.json'),
                    ],
                    darwin: [
                        path.join(homeDir, 'Library', 'Application Support', 'Cursor', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(homeDir, 'Library', 'Application Support', 'Cursor', 'User', 'mcp.json'),
                    ],
                    linux: [
                        path.join(homeDir, '.config', 'Cursor', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(homeDir, '.config', 'Cursor', 'User', 'mcp.json'),
                    ],
                });
            case 'vscode':
                return this.byPlatform({
                    win32: [
                        path.join(appData, 'Code', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(appData, 'Code', 'User', 'settings.json'),
                        path.join(this.cwd, '.vscode', 'mcp.json'),
                    ],
                    darwin: [
                        path.join(homeDir, 'Library', 'Application Support', 'Code', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(homeDir, 'Library', 'Application Support', 'Code', 'User', 'settings.json'),
                        path.join(this.cwd, '.vscode', 'mcp.json'),
                    ],
                    linux: [
                        path.join(homeDir, '.config', 'Code', 'User', 'globalStorage', 'mcp-servers.json'),
                        path.join(homeDir, '.config', 'Code', 'User', 'settings.json'),
                        path.join(this.cwd, '.vscode', 'mcp.json'),
                    ],
                });
        }
    }

    private byPlatform(pathsByPlatform: Record<'win32' | 'darwin' | 'linux', string[]>): string[] {
        if (this.platform === 'win32') {
            return pathsByPlatform.win32;
        }
        if (this.platform === 'darwin') {
            return pathsByPlatform.darwin;
        }
        return pathsByPlatform.linux;
    }

    private mergeClientConfig(existingDocument: JsonObject, mcpServers: JsonObject): JsonObject {
        const baseDocument = isRecord(existingDocument) ? { ...existingDocument } : {};
        return {
            ...baseDocument,
            mcpServers,
        };
    }

    private buildMcpServersConfig(servers: ServerRecord[]): JsonObject {
        return Object.fromEntries(
            servers
                .map((server) => {
                    const definition = this.toClientDefinition(server);
                    if (!definition) {
                        return null;
                    }
                    return [server.name, definition] as const;
                })
                .filter((entry): entry is readonly [string, JsonObject] => entry !== null),
        );
    }

    private toClientDefinition(server: ServerRecord): JsonObject | null {
        if (server.type === 'STDIO') {
            if (!server.command) {
                return null;
            }

            const definition: JsonObject = {
                command: server.command,
            };

            if (server.args.length > 0) {
                definition.args = [...server.args];
            }

            if (Object.keys(server.env).length > 0) {
                definition.env = { ...server.env };
            }

            return definition;
        }

        if (!server.url) {
            return null;
        }

        const headers: Record<string, string> = { ...server.headers };
        if (server.bearerToken) {
            headers.Authorization = `Bearer ${server.bearerToken}`;
        }

        const definition: JsonObject = {
            url: server.url,
        };

        if (Object.keys(headers).length > 0) {
            definition.headers = headers;
        }

        return definition;
    }
}

export const clientConfigSyncService = new ClientConfigSyncService();
