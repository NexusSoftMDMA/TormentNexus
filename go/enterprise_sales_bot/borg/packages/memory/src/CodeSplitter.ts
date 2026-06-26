
import ts from 'typescript';

/**
 * CodeSplitter
 * Semantic chunking for code files.
 * Uses AST analysis for TypeScript/JavaScript and heuristic indentation for others.
 */
export class CodeSplitter {
    /**
     * Splits code into semantic chunks.
     * @param code The source code
     * @param extension File extension (e.g. .ts, .py)
     * @param maxChunkSize Max chars per chunk (soft limit)
     */
    static split(code: string, extension: string, maxChunkSize: number = 1000): string[] {
        if (['.ts', '.tsx', '.js', '.jsx'].includes(extension)) {
            try {
                return this.splitTS(code, extension, maxChunkSize);
            } catch (e) {
                console.error("CodeSplitter AST failure, falling back to line-based:", e);
            }
        }
        return this.splitGeneric(code, maxChunkSize);
    }

    private static splitTS(code: string, extension: string, maxChunkSize: number): string[] {
        const chunks: string[] = [];
        const sourceFile = ts.createSourceFile(
            `temp${extension}`,
            code,
            ts.ScriptTarget.Latest,
            true
        );

        const chunkNodes: ts.Node[] = [];

        // Top-level traversal
        ts.forEachChild(sourceFile, (node) => {
            // Check for top-level constructs we want to keep whole
            if (
                ts.isFunctionDeclaration(node) ||
                ts.isClassDeclaration(node) ||
                ts.isInterfaceDeclaration(node) ||
                ts.isTypeAliasDeclaration(node) ||
                ts.isModuleDeclaration(node) ||
                ts.isEnumDeclaration(node)
            ) {
                // If the node is massive, we might need to split it (TODO: Sub-chunking)
                // For now, accept it as one semantic unit.
                chunks.push(node.getText(sourceFile));
            } else {
                // Statements, imports, exports, variables...
                // Group them until we hit a declaration or size limit
                // Ideally, we group strictly related imports/vars, but sequential grouping is MVP.
                // Simple approach: Just push them as text immediately if small, or buffer?
                // Let's buffer 'loose' nodes and flush when big enough.
                const text = node.getText(sourceFile);
                if (text.trim().length > 0) {
                    // Check if last chunk is "open" / "statement cluster"?
                    // Simplified: Just make every top-level statement a chunk if meaningful size?
                    // Better: Accumulate misc top-level stuff.
                    chunks.push(text);
                }
            }
        });

        // Post-process logic could merge small adjacent chunks here
        return this.mergeSmallChunks(chunks, maxChunkSize);
    }

    private static mergeSmallChunks(chunks: string[], maxSize: number): string[] {
        const merged: string[] = [];
        let buffer: string[] = [];
        let currentSize = 0;

        for (const chunk of chunks) {
            const size = chunk.length;
            if (currentSize + size > maxSize && buffer.length > 0) {
                merged.push(buffer.join('\n\n'));
                buffer = [];
                currentSize = 0;
            }
            buffer.push(chunk);
            currentSize += size;
        }
        if (buffer.length > 0) {
            merged.push(buffer.join('\n\n'));
        }
        return merged;
    }


    private static splitGeneric(code: string, maxChunkSize: number): string[] {
        // Naive line-based splitting for now, enhanced with Block Detection
        const lines = code.split('\n');
        const chunks: string[] = [];
        let currentChunk: string[] = [];
        let currentSize = 0;

        for (const line of lines) {
            currentChunk.push(line);
            currentSize += line.length + 1; // +1 for newline

            // Heuristic using indentation:
            // If line starts with NO spaces (top level), it might be a good break point
            // IF current chunk is big enough.
            const isTopLevel = !line.startsWith(' ') && !line.startsWith('\t') && line.trim().length > 0;
            const isBlockEnd = line.trim() === '}' || line.trim() === '};';

            if (currentSize >= maxChunkSize) {
                // Try to find a good break point
                if (isTopLevel || isBlockEnd) {
                    chunks.push(currentChunk.join('\n'));
                    currentChunk = [];
                    currentSize = 0;
                }
            } else if (currentSize >= maxChunkSize * 2) {
                // Hard limit, force split
                chunks.push(currentChunk.join('\n'));
                currentChunk = [];
                currentSize = 0;
            }
        }

        if (currentChunk.length > 0) {
            chunks.push(currentChunk.join('\n'));
        }

        return chunks;
    }
}
