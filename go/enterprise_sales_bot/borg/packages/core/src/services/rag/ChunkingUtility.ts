/**
 * ChunkingUtility.ts
 * 
 * Provides various strategies for splitting large documents into smaller chunks
 * suitable for vector embedding and retrieval (RAG).
 */

export interface ChunkingOptions {
    chunkSize?: number;
    chunkOverlap?: number;
    strategy?: 'sliding_window' | 'recursive' | 'semantic';
}

export class ChunkingUtility {

    /**
     * Main entry point for chunking text based on the provided strategy.
     */
    static chunkText(text: string, options: ChunkingOptions = {}): string[] {
        const { strategy = 'recursive', chunkSize = 1000, chunkOverlap = 200 } = options;

        switch (strategy) {
            case 'sliding_window':
                return this.slidingWindowChunk(text, chunkSize, chunkOverlap);
            case 'semantic':
                return this.semanticChunk(text, chunkSize);
            case 'recursive':
            default:
                return this.recursiveChunk(text, chunkSize, chunkOverlap);
        }
    }

    /**
     * Basic sliding window over characters.
     */
    private static slidingWindowChunk(text: string, size: number, overlap: number): string[] {
        if (text.length <= size) return [text];

        const chunks: string[] = [];
        let i = 0;

        while (i < text.length) {
            chunks.push(text.slice(i, i + size));
            i += (size - overlap);
        }

        return chunks;
    }

    /**
     * Recursive character text splitter.
     * Tries to split on paragraphs, then sentences, then words, then characters.
     */
    private static recursiveChunk(text: string, size: number, overlap: number): string[] {
        const separators = ['\n\n', '\n', '. ', ' ', ''];

        function splitRecursively(textToSplit: string, sepIndex: number): string[] {
            if (textToSplit.length <= size) {
                return [textToSplit];
            }

            if (sepIndex >= separators.length) {
                // Fallback to strict sliding window if we can't split nicely
                return ChunkingUtility.slidingWindowChunk(textToSplit, size, overlap);
            }

            const separator = separators[sepIndex];
            const splits = textToSplit.split(separator);

            const chunks: string[] = [];
            let currentChunk = '';

            for (const split of splits) {
                const combined = currentChunk ? currentChunk + separator + split : split;
                if (combined.length <= size) {
                    currentChunk = combined;
                } else {
                    if (currentChunk) {
                        chunks.push(currentChunk);
                    }
                    if (split.length > size) {
                        // The split itself is too big, recurse on it with next separator
                        chunks.push(...splitRecursively(split, sepIndex + 1));
                        currentChunk = '';
                    } else {
                        currentChunk = split;
                    }
                }
            }

            if (currentChunk) {
                chunks.push(currentChunk);
            }

            // Note: True recursive overlapping can be complex. 
            // For simplicity, we just return the cleanly split joined parts here.
            return chunks;
        }

        return splitRecursively(text, 0);
    }

    /**
     * A pseudo-semantic splitter that tries to respect markdown headers and double line breaks exclusively.
     */
    private static semanticChunk(text: string, maxSize: number): string[] {
        // Split by major sections (e.g. headers or double newlines)
        const sections = text.split(/(?=\n#{1,6} )|\n\n/);

        const chunks: string[] = [];
        let currentChunk = '';

        for (const section of sections) {
            if ((currentChunk.length + section.length) > maxSize && currentChunk.length > 0) {
                chunks.push(currentChunk.trim());
                currentChunk = '';
            }
            currentChunk += section + '\n\n';
        }

        if (currentChunk.trim().length > 0) {
            chunks.push(currentChunk.trim());
        }

        return chunks;
    }
}
