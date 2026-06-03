/**
 * LSP Service - Language Server Protocol Integration for TormentNexus
 * 
 * Provides semantic code understanding via Language Server Protocol integration.
 * Inspired by Serena's architecture but implemented in TypeScript.
 * 
 * Features:
 * - Multi-language LSP server management
 * - Symbol index with cross-references
 * - Find symbol, find references, go to definition
 * - Symbol-based code editing
 */

import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import { EventEmitter } from 'events';

// LSP Types based on Language Server Protocol specification
export interface Position {
    line: number;
    character: number;
}

export interface Range {
    start: Position;
    end: Position;
}

export interface Location {
    uri: string;
    range: Range;
}

export interface SymbolInformation {
    name: string;
    kind: SymbolKind;
    location: Location;
    containerName?: string;
}

export interface DocumentSymbol {
    name: string;
    detail?: string;
    kind: SymbolKind;
    range: Range;
    selectionRange: Range;
    children?: DocumentSymbol[];
}

export enum SymbolKind {
    File = 1,
    Module = 2,
    Namespace = 3,
    Package = 4,
    Class = 5,
    Method = 6,
    Property = 7,
    Field = 8,
    Constructor = 9,
    Enum = 10,
    Interface = 11,
    Function = 12,
    Variable = 13,
    Constant = 14,
    String = 15,
    Number = 16,
    Boolean = 17,
    Array = 18,
    Object = 19,
    Key = 20,
    Null = 21,
    EnumMember = 22,
    Struct = 23,
    Event = 24,
    Operator = 25,
    TypeParameter = 26,
}

export interface LSPMessage {
    jsonrpc: '2.0';
    id?: number;
    method?: string;
    params?: unknown;
    result?: unknown;
    error?: { code: number; message: string };
}

export interface LanguageServerConfig {
    language: string;
    command: string;
    args: string[];
    rootPath: string;
    initializationOptions?: Record<string, unknown>;
}

interface PendingRequest {
    resolve: (result: unknown) => void;
    reject: (error: Error) => void;
}

/**
 * Manages a single Language Server process
 */
export class LanguageServer extends EventEmitter {
    private process: ChildProcess | null = null;
    private messageId = 0;
    private pendingRequests = new Map<number, PendingRequest>();
    private buffer = '';
    private initialized = false;
    private capabilities: Record<string, unknown> = {};

    constructor(
        public readonly config: LanguageServerConfig
    ) {
        super();
    }

    async start(): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                this.process = spawn(this.config.command, this.config.args, {
                    cwd: this.config.rootPath,
                    stdio: ['pipe', 'pipe', 'pipe'],
                });

                this.process.stdout?.on('data', (data) => this.handleData(data));
                this.process.stderr?.on('data', (data) => {
                    console.error(`[LSP ${this.config.language}] ${data.toString()}`);
                });
                this.process.on('error', (err) => {
                    this.emit('error', err);
                    reject(err);
                });
                this.process.on('exit', (code) => {
                    this.emit('exit', code);
                });

                // Send initialize request
                this.initialize()
                    .then(() => {
                        this.initialized = true;
                        resolve();
                    })
                    .catch(reject);
            } catch (error) {
                reject(error);
            }
        });
    }

    async stop(): Promise<void> {
        if (this.process && this.initialized) {
            await this.request('shutdown', {});
            this.notify('exit', {});
            this.process.kill();
            this.process = null;
            this.initialized = false;
        }
    }

    private async initialize(): Promise<void> {
        const result = await this.request('initialize', {
            processId: process.pid,
            rootPath: this.config.rootPath,
            rootUri: `file://${this.config.rootPath}`,
            capabilities: {
                textDocument: {
                    synchronization: { dynamicRegistration: false },
                    completion: { dynamicRegistration: false },
                    hover: { dynamicRegistration: false },
                    definition: { dynamicRegistration: false },
                    references: { dynamicRegistration: false },
                    documentSymbol: { dynamicRegistration: false },
                },
                workspace: {
                    workspaceFolders: true,
                    symbol: { dynamicRegistration: false },
                },
            },
            initializationOptions: this.config.initializationOptions,
        });

        this.capabilities = (result as Record<string, unknown>)?.capabilities as Record<string, unknown> || {};
        this.notify('initialized', {});
    }

    private handleData(data: Buffer): void {
        this.buffer += data.toString();

        while (true) {
            const headerEnd = this.buffer.indexOf('\r\n\r\n');
            if (headerEnd === -1) break;

            const header = this.buffer.slice(0, headerEnd);
            const contentLengthMatch = header.match(/Content-Length: (\d+)/);
            if (!contentLengthMatch) break;

            const contentLength = parseInt(contentLengthMatch[1], 10);
            const messageStart = headerEnd + 4;
            const messageEnd = messageStart + contentLength;

            if (this.buffer.length < messageEnd) break;

            const messageStr = this.buffer.slice(messageStart, messageEnd);
            this.buffer = this.buffer.slice(messageEnd);

            try {
                const message: LSPMessage = JSON.parse(messageStr);
                this.handleMessage(message);
            } catch (error) {
                console.error(`[LSP ${this.config.language}] Failed to parse message:`, error);
            }
        }
    }

    private handleMessage(message: LSPMessage): void {
        if (message.id !== undefined && !message.method) {
            // Response to a request
            const pending = this.pendingRequests.get(message.id);
            if (pending) {
                this.pendingRequests.delete(message.id);
                if (message.error) {
                    pending.reject(new Error(message.error.message));
                } else {
                    pending.resolve(message.result);
                }
            }
        } else if (message.method) {
            // Notification or request from server
            this.emit('notification', message.method, message.params);
        }
    }

    private sendMessage(message: LSPMessage): void {
        if (!this.process?.stdin) {
            throw new Error('Language server not started');
        }
        const content = JSON.stringify(message);
        const header = `Content-Length: ${Buffer.byteLength(content)}\r\n\r\n`;
        this.process.stdin.write(header + content);
    }

    async request<T = unknown>(method: string, params: unknown): Promise<T> {
        const id = ++this.messageId;
        const message: LSPMessage = {
            jsonrpc: '2.0',
            id,
            method,
            params,
        };

        return new Promise((resolve, reject) => {
            this.pendingRequests.set(id, {
                resolve: resolve as (result: unknown) => void,
                reject,
            });
            this.sendMessage(message);

            // Timeout after 30 seconds
            setTimeout(() => {
                if (this.pendingRequests.has(id)) {
                    this.pendingRequests.delete(id);
                    reject(new Error(`Request ${method} timed out`));
                }
            }, 30000);
        });
    }

    notify(method: string, params: unknown): void {
        const message: LSPMessage = {
            jsonrpc: '2.0',
            method,
            params,
        };
        this.sendMessage(message);
    }

    // Document management
    openDocument(uri: string, languageId: string, text: string): void {
        this.notify('textDocument/didOpen', {
            textDocument: { uri, languageId, version: 1, text },
        });
    }

    closeDocument(uri: string): void {
        this.notify('textDocument/didClose', {
            textDocument: { uri },
        });
    }

    // LSP Operations
    async getDocumentSymbols(uri: string): Promise<DocumentSymbol[] | SymbolInformation[]> {
        return this.request('textDocument/documentSymbol', {
            textDocument: { uri },
        });
    }

    async findDefinition(uri: string, position: Position): Promise<Location | Location[] | null> {
        return this.request('textDocument/definition', {
            textDocument: { uri },
            position,
        });
    }

    async findReferences(uri: string, position: Position, includeDeclaration = true): Promise<Location[]> {
        return this.request('textDocument/references', {
            textDocument: { uri },
            position,
            context: { includeDeclaration },
        });
    }

    async getHover(uri: string, position: Position): Promise<unknown> {
        return this.request('textDocument/hover', {
            textDocument: { uri },
            position,
        });
    }

    async rename(uri: string, position: Position, newName: string): Promise<unknown> {
        return this.request('textDocument/rename', {
            textDocument: { uri },
            position,
            newName,
        });
    }
}

/**
 * LSP Service - Manages multiple language servers
 */
export class LSPService {
    private servers = new Map<string, LanguageServer>();
    private symbolIndex = new Map<string, SymbolInformation[]>();

    constructor(private rootPath: string) { }

    public getStatus() {
        return {
            status: this.symbolIndex.size > 0 ? 'indexed' : 'idle',
            filesIndexed: this.symbolIndex.size,
            activeServers: Array.from(this.servers.keys())
        };
    }

    /**
     * Get or create a language server for the given language
     */
    async getServer(language: string): Promise<LanguageServer> {
        if (this.servers.has(language)) {
            return this.servers.get(language)!;
        }

        const config = this.getLanguageConfig(language);
        if (!config) {
            throw new Error(`No configuration available for language: ${language}`);
        }

        const server = new LanguageServer(config);
        await server.start();
        this.servers.set(language, server);
        return server;
    }

    /**
     * Get language server configuration for a language
     */
    private getLanguageConfig(language: string): LanguageServerConfig | null {
        const configs: Record<string, Omit<LanguageServerConfig, 'rootPath'>> = {
            typescript: {
                language: 'typescript',
                command: 'typescript-language-server',
                args: ['--stdio'],
            },
            javascript: {
                language: 'javascript',
                command: 'typescript-language-server',
                args: ['--stdio'],
            },
            python: {
                language: 'python',
                command: 'pyright-langserver',
                args: ['--stdio'],
            },
            rust: {
                language: 'rust',
                command: 'rust-analyzer',
                args: [],
            },
            go: {
                language: 'go',
                command: 'gopls',
                args: [],
            },
        };

        const config = configs[language];
        if (!config) return null;

        return { ...config, rootPath: this.rootPath };
    }

    /**
     * Detect language from file extension
     */
    detectLanguage(filePath: string): string | null {
        const ext = path.extname(filePath).toLowerCase();
        const languageMap: Record<string, string> = {
            '.ts': 'typescript',
            '.tsx': 'typescript',
            '.js': 'javascript',
            '.jsx': 'javascript',
            '.py': 'python',
            '.rs': 'rust',
            '.go': 'go',
        };
        return languageMap[ext] || null;
    }

    /**
     * Find symbol by name in a file
     */
    async findSymbol(filePath: string, symbolName: string): Promise<SymbolInformation | null> {
        const language = this.detectLanguage(filePath);
        if (!language) return null;

        const server = await this.getServer(language);
        const uri = `file://${path.resolve(this.rootPath, filePath)}`;

        // Open the document
        const content = fs.readFileSync(filePath, 'utf-8');
        server.openDocument(uri, language, content);

        // Get symbols
        const symbols = await server.getDocumentSymbols(uri);

        // Find matching symbol
        const findInSymbols = (items: (DocumentSymbol | SymbolInformation)[]): SymbolInformation | null => {
            for (const item of items) {
                if (item.name === symbolName) {
                    // Convert DocumentSymbol to SymbolInformation if needed
                    if ('range' in item && !('location' in item)) {
                        return {
                            name: item.name,
                            kind: item.kind,
                            location: { uri, range: item.range },
                        };
                    }
                    return item as SymbolInformation;
                }
                // Check children for DocumentSymbol
                if ('children' in item && item.children) {
                    const found = findInSymbols(item.children);
                    if (found) return found;
                }
            }
            return null;
        };

        const result = findInSymbols(symbols);
        server.closeDocument(uri);
        return result;
    }

    /**
     * Find all references to a symbol
     */
    async findReferences(filePath: string, line: number, character: number): Promise<Location[]> {
        const language = this.detectLanguage(filePath);
        if (!language) return [];

        const server = await this.getServer(language);
        const uri = `file://${path.resolve(this.rootPath, filePath)}`;

        // Open the document
        const content = fs.readFileSync(filePath, 'utf-8');
        server.openDocument(uri, language, content);

        const references = await server.findReferences(uri, { line, character });

        server.closeDocument(uri);
        return references || [];
    }

    /**
     * Go to definition
     */
    async goToDefinition(filePath: string, line: number, character: number): Promise<Location | null> {
        const language = this.detectLanguage(filePath);
        if (!language) return null;

        const server = await this.getServer(language);
        const uri = `file://${path.resolve(this.rootPath, filePath)}`;

        // Open the document
        const content = fs.readFileSync(filePath, 'utf-8');
        server.openDocument(uri, language, content);

        const definition = await server.findDefinition(uri, { line, character });

        server.closeDocument(uri);

        if (Array.isArray(definition)) {
            return definition[0] || null;
        }
        return definition;
    }

    /**
     * Get all symbols in a file
     */
    async getSymbols(filePath: string): Promise<SymbolInformation[]> {
        const language = this.detectLanguage(filePath);
        if (!language) return [];

        const server = await this.getServer(language);
        const uri = `file://${path.resolve(this.rootPath, filePath)}`;

        // Open the document
        const content = fs.readFileSync(filePath, 'utf-8');
        server.openDocument(uri, language, content);

        const symbols = await server.getDocumentSymbols(uri);

        // Flatten symbols if they are DocumentSymbols
        const flatten = (items: (DocumentSymbol | SymbolInformation)[], result: SymbolInformation[] = []): SymbolInformation[] => {
            for (const item of items) {
                if ('range' in item && !('location' in item)) {
                    result.push({
                        name: item.name,
                        kind: item.kind,
                        location: { uri, range: item.range },
                    });
                    if (item.children) {
                        flatten(item.children, result);
                    }
                } else {
                    result.push(item as SymbolInformation);
                }
            }
            return result;
        };

        const flattened = flatten(symbols);
        server.closeDocument(uri);
        return flattened;
    }

    /**
     * Rename a symbol across the project
     */
    async renameSymbol(filePath: string, line: number, character: number, newName: string): Promise<unknown> {
        const language = this.detectLanguage(filePath);
        if (!language) return null;

        const server = await this.getServer(language);
        const uri = `file://${path.resolve(this.rootPath, filePath)}`;

        // Open the document
        const content = fs.readFileSync(filePath, 'utf-8');
        server.openDocument(uri, language, content);

        const result = await server.rename(uri, { line, character }, newName);

        server.closeDocument(uri);
        return result;
    }

    /**
     * Index all symbols in the project
     */
    async indexProject(extensions: string[] = ['.ts', '.js', '.py']): Promise<void> {
        const files = this.findFiles(this.rootPath, extensions);

        for (const file of files) {
            try {
                const symbols = await this.getSymbols(file);
                this.symbolIndex.set(file, symbols);
            } catch (error) {
                console.warn(`Failed to index ${file}:`, error);
            }
        }
    }

    /**
     * Find files with given extensions
     */
    private findFiles(dir: string, extensions: string[], files: string[] = []): string[] {
        const entries = fs.readdirSync(dir, { withFileTypes: true });

        for (const entry of entries) {
            const fullPath = path.join(dir, entry.name);

            // Skip common ignored directories
            if (entry.isDirectory()) {
                if (entry.name.startsWith('.') ||
                    entry.name === 'node_modules' ||
                    entry.name === 'dist' ||
                    entry.name === '__pycache__') {
                    continue;
                }
                this.findFiles(fullPath, extensions, files);
            } else if (extensions.some(ext => entry.name.endsWith(ext))) {
                files.push(fullPath);
            }
        }

        return files;
    }

    /**
     * Search symbols by name across indexed files
     */
    searchSymbols(query: string): SymbolInformation[] {
        const results: SymbolInformation[] = [];
        const lowerQuery = query.toLowerCase();

        for (const [, symbols] of this.symbolIndex) {
            for (const symbol of symbols) {
                if (symbol.name.toLowerCase().includes(lowerQuery)) {
                    results.push(symbol);
                }
            }
        }

        return results;
    }

    /**
     * Stop all language servers
     */
    async shutdown(): Promise<void> {
        const stopPromises: Promise<void>[] = [];
        for (const server of this.servers.values()) {
            stopPromises.push(server.stop());
        }
        await Promise.all(stopPromises);
        this.servers.clear();
    }
}

export default LSPService;
