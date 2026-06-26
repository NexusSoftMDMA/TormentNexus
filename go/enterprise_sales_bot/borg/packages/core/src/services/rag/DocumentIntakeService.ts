/**
 * DocumentIntakeService.ts
 * 
 * Ingests documents (PDF, DOCX, Markdown, TXT), extracts text, chunks it,
 * generates embeddings for each chunk, and saves it to the vector store memory.
 */

import fs from 'fs';
import path from 'path';
import { ChunkingUtility, ChunkingOptions } from './ChunkingUtility.js';
import { EmbeddingService } from './EmbeddingService.js';
import { IMemoryProvider } from '../../interfaces/IMemoryProvider.js';

// We dynamically import text extraction libraries to avoid heavy dependencies 
// if they aren't used or haven't been installed yet.
export class DocumentIntakeService {
    private embeddingService: EmbeddingService;
    private memoryProvider: IMemoryProvider;

    constructor(memoryProvider: IMemoryProvider, embeddingService?: EmbeddingService) {
        this.memoryProvider = memoryProvider;
        this.embeddingService = embeddingService || new EmbeddingService('local'); // Default to local transformers
    }

    /**
     * Ingest a file from disk into the memory system.
     */
    async ingestFile(filePath: string, userId: string, options: ChunkingOptions = {}): Promise<{ chunks: number, success: boolean }> {
        if (!fs.existsSync(filePath)) {
            throw new Error(`File not found: ${filePath}`);
        }

        const ext = path.extname(filePath).toLowerCase();
        let text = '';

        if (ext === '.pdf') {
            text = await this.extractPdf(filePath);
        } else if (ext === '.docx') {
            text = await this.extractDocx(filePath);
        } else if (ext === '.txt' || ext === '.md' || ext === '.json' || ext === '.csv') {
            text = fs.readFileSync(filePath, 'utf-8');
        } else {
            throw new Error(`Unsupported file type: ${ext}`);
        }

        return this.ingestText(text, filePath, userId, options);
    }

    /**
     * Ingest raw text directly.
     */
    async ingestText(text: string, source: string, userId: string, options: ChunkingOptions = {}): Promise<{ chunks: number, success: boolean }> {
        if (!text || text.trim().length === 0) {
            return { chunks: 0, success: false };
        }

        // 1. Chunk Text
        const chunks = ChunkingUtility.chunkText(text, options);

        // 2. Embed & Save each chunk
        let saved = 0;

        // We could batch embed, but for now we do it sequentially or promise.all
        // to avoid huge memory spikes with local transformers.
        for (let i = 0; i < chunks.length; i++) {
            const chunk = chunks[i];

            // Note: The actual memory provider might handle the embedding internally 
            // if configured (like lancedb vector column), but if it expects embedding vectors directly:
            const vector = await this.embeddingService.embed(chunk);

            await this.memoryProvider.saveMemory(
                chunk,
                {
                    source: source,
                    chunkIndex: i,
                    totalChunks: chunks.length,
                    type: 'document_intake',
                    vector: vector // Pass vector explicitly if the provider supports it
                },
                userId
            );
            saved++;
        }

        return { chunks: saved, success: true };
    }

    // ---- Extractors ----

    private async extractPdf(filePath: string): Promise<string> {
        try {
            const pdf = await import('pdf-parse');
            const dataBuffer = fs.readFileSync(filePath);
            const data = await pdf.default(dataBuffer);
            return data.text;
        } catch (e) {
            console.error("Failed to extract PDF. Is 'pdf-parse' installed?", e);
            throw new Error("PDF extraction failed: pdf-parse not available.");
        }
    }

    private async extractDocx(filePath: string): Promise<string> {
        try {
            const mammoth = await import('mammoth');
            const result = await mammoth.extractRawText({ path: filePath });
            return result.value;
        } catch (e) {
            console.error("Failed to extract DOCX. Is 'mammoth' installed?", e);
            throw new Error("DOCX extraction failed: mammoth not available.");
        }
    }
}
