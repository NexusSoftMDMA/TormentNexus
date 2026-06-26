
import fs from 'fs/promises';
import path from 'path';
import { VectorStore } from './VectorStore.js';
import { CodeSplitter } from './CodeSplitter.js';
import crypto from 'crypto';

// Basic list of extensions to index
const EXTENSIONS = new Set(['.ts', '.tsx', '.js', '.jsx', '.md', '.json', '.css', '.html']);

export interface IndexerStorage {
    initialize(): Promise<void>;
    addDocuments(docs: any[]): Promise<void>;
}

export class Indexer {
    private vectorStore: IndexerStorage;
    private maxChunkSize: number = 500; // chars approx for now? or tokens. Simple chars for speed.

    constructor(vectorStore: IndexerStorage) {
        this.vectorStore = vectorStore;
    }

    async indexDirectory(rootDir: string) {
        console.log(`[Indexer] Scanning ${rootDir}...`);

        await this.vectorStore.initialize();

        const files = await this.walk(rootDir);
        const codeDocs: any[] = [];

        for (const file of files) {
            // Compute relative path for ID
            const relPath = path.relative(rootDir, file);

            try {
                const content = await fs.readFile(file, 'utf-8');
                const fileHash = crypto.createHash('md5').update(content).digest('hex');

                // Check if already indexed?
                // For MVP, we overwrite or just append.
                // Real impl should check hash in DB. (Skipped for speed in MVP)

                // Chunking
                const chunks = CodeSplitter.split(content, path.extname(file));

                chunks.forEach((chunk, index) => {
                    codeDocs.push({
                        id: `${relPath}#${index}`,
                        file_path: relPath,
                        content: chunk,
                        hash: fileHash
                    });
                });

            } catch (e: any) {
                console.error(`[Indexer] Error processing ${relPath}: ${e.message}`);
            }
        }

        if (codeDocs.length > 0) {
            // Batch add
            // Lancedb handles batching, but we can do it in chunks of 100 too to prevent OOM
            const BATCH_SIZE = 50;
            for (let i = 0; i < codeDocs.length; i += BATCH_SIZE) {
                const batch = codeDocs.slice(i, i + BATCH_SIZE);
                await this.vectorStore.addDocuments(batch);
                console.log(`[Indexer] Indexed batch ${i}-${i + BATCH_SIZE} / ${codeDocs.length}`);
            }
        }

        return codeDocs.length;
    }

    async indexSymbols(rootDir: string) {
        console.log(`[Indexer] Indexing Symbols in ${rootDir}...`);

        // Dynamic import to avoid load-time dependency if unused
        const ts = await import('typescript');

        const fileNames = await this.walk(rootDir);
        const tsFiles = fileNames.filter(f => f.endsWith('.ts') || f.endsWith('.tsx'));

        const program = ts.createProgram(tsFiles, {
            target: ts.ScriptTarget.ESNext,
            module: ts.ModuleKind.CommonJS,
            allowJs: true
        });

        const checker = program.getTypeChecker();
        const symbols: any[] = [];

        // Normalize paths for comparison (Windows fix)
        const normalizedFileNames = new Set(fileNames.map(f => f.replace(/\\/g, '/').toLowerCase()));

        for (const sourceFile of program.getSourceFiles()) {
            const normalizedSourcePath = sourceFile.fileName.replace(/\\/g, '/').toLowerCase();
            // Check if this source file is one of our target files (ignoring libs)
            // Note: TS might use absolute paths with forward slashes
            const isTarget = normalizedFileNames.has(normalizedSourcePath);

            if (!isTarget) continue;

            ts.forEachChild(sourceFile, (node) => {
                if (ts.isFunctionDeclaration(node) && node.name) {
                    const symbol = checker.getSymbolAtLocation(node.name);
                    if (symbol) {
                        symbols.push(this.extractSymbol(ts, node, symbol, sourceFile, 'function'));
                    }
                } else if (ts.isClassDeclaration(node) && node.name) {
                    const symbol = checker.getSymbolAtLocation(node.name);
                    if (symbol) {
                        symbols.push(this.extractSymbol(ts, node, symbol, sourceFile, 'class'));
                        // Methods
                        node.members.forEach((member) => {
                            if (ts.isMethodDeclaration(member) && member.name) {
                                const memSymbol = checker.getSymbolAtLocation(member.name);
                                if (memSymbol) {
                                    symbols.push(this.extractSymbol(ts, member, memSymbol, sourceFile, 'method', node.name?.text));
                                }
                            }
                        });
                    }
                } else if (ts.isInterfaceDeclaration(node) && node.name) {
                    const symbol = checker.getSymbolAtLocation(node.name);
                    if (symbol) {
                        symbols.push(this.extractSymbol(ts, node, symbol, sourceFile, 'interface'));
                    }
                }
            });
        }

        console.log(`[Indexer] Found ${symbols.length} symbols.`);

        if (symbols.length > 0) {
            const BATCH_SIZE = 50;
            for (let i = 0; i < symbols.length; i += BATCH_SIZE) {
                const batch = symbols.slice(i, i + BATCH_SIZE);
                await this.vectorStore.addDocuments(batch);
            }
        }

        return symbols.length;
    }

    private extractSymbol(ts: any, node: any, symbol: any, sourceFile: any, kind: string, parentName?: string) {
        const name = symbol.getName();
        const fullName = parentName ? `${parentName}.${name}` : name;

        // Extract DocBlock
        const docComment = ts.displayPartsToString(symbol.getDocumentationComment(undefined));

        // Extract Signature
        const signature = node.getText(sourceFile).split('{')[0].trim(); // Rough signature

        const relPath = path.relative(process.cwd(), sourceFile.fileName); // Use process.cwd as root for consistency

        return {
            id: `symbol:${relPath}:${fullName}`,
            file_path: relPath,
            content: `${kind} ${fullName}\n${docComment}\n${signature}`, // Searchable content
            hash: 'symbol',
            metadata: {
                type: 'symbol',
                kind,
                name: fullName,
                signature,
                line: sourceFile.getLineAndCharacterOfPosition(node.getStart()).line + 1
            }
        };
    }

    private async walk(dir: string): Promise<string[]> {
        let results: string[] = [];
        try {
            const list = await fs.readdir(dir);
            for (const file of list) {
                const filepath = path.join(dir, file);
                const stat = await fs.stat(filepath);

                if (stat && stat.isDirectory()) {
                    // Ignore common junk
                    if (['node_modules', '.git', 'dist', 'build', '.next'].includes(file)) continue;
                    results = results.concat(await this.walk(filepath));
                } else {
                    if (EXTENSIONS.has(path.extname(filepath))) {
                        results.push(filepath);
                    }
                }
            }
        } catch (e) { /* ignore access errors */ }
        return results;
    }

}
