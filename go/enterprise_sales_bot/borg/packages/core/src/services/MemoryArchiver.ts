/**
 * @file MemoryArchiver.ts
 * @module packages/core/src/services/MemoryArchiver
 *
 * WHAT: Advanced session transcript archiving and memory extraction system.
 * Converts raw JSON sessions into compressed plaintext archives and extracts 
 * high-value semantic memories via LLM.
 *
 * WHY: Retaining thousands of raw JSON sessions is storage-heavy and context-noisy.
 * Archiving them as compressed plaintext preserves the history while extraction 
 * keeps the core knowledge accessible and lightweight.
 *
 * HOW:
 * 1. Read raw session files (JSON/JSONL).
 * 2. Transform into structured plaintext (Speaker: Message).
 * 3. Store in a compressed ZIP archive (using adm-zip or similar).
 * 4. Run extraction turn: LLM summarizes the session into key facts/decisions.
 * 5. Persist extracted memories into the AgentMemoryService.
 */

import fs from 'fs/promises';
import path from 'path';
import AdmZip from 'adm-zip';
import { LLMService } from "@tormentnexus/ai";
import { AgentMemoryService } from "./AgentMemoryService.js";
import { A2ALogger } from "@tormentnexus/agents";

export interface ArchivedSessionMetadata {
    originalId: string;
    sourceTool: string;
    title: string;
    timestamp: number;
    compressedSize: number;
}

export class MemoryArchiver {
    private archivePath: string;

    constructor(
        private workspaceRoot: string,
        private llmService: LLMService,
        private agentMemory: AgentMemoryService,
        private a2aLogger?: A2ALogger
    ) {
        this.archivePath = path.join(this.workspaceRoot, 'data', 'archives', 'sessions.zip');
    }

    /**
     * Archives a session transcript and extracts memories.
     */
    public async archiveAndExtract(sessionData: any): Promise<ArchivedSessionMetadata | null> {
        const sessionId = sessionData.id || `session-${Date.now()}`;
        const transcript = this.formatToPlaintext(sessionData);
        
        if (!transcript) return null;

        // 1. Extract Valuable Memories via LLM
        await this.extractValuableMemories(transcript, sessionData.title || sessionId);

        // 2. Add to Compressed Archive
        const zip = new AdmZip(await this.ensureArchiveExists());
        const entryName = `${sessionId}.txt`;
        zip.addFile(entryName, Buffer.from(transcript, 'utf-8'), `Source: ${sessionData.sourceTool}`);

        // 3. Add A2A Logs if available
        if (this.a2aLogger) {
            const logs = await this.a2aLogger.getRecentLogs(200);
            if (logs.length > 0) {
                const logsContent = logs.map((l: any) => JSON.stringify(l)).join('\n');
                zip.addFile(`${sessionId}.a2a.jsonl`, Buffer.from(logsContent, 'utf-8'));
            }
        }

        zip.writeZip(this.archivePath);

        const entry = zip.getEntry(entryName);
        
        return {
            originalId: sessionId,
            sourceTool: sessionData.sourceTool || 'unknown',
            title: sessionData.title || sessionId,
            timestamp: Date.now(),
            compressedSize: entry?.header.compressedSize || 0
        };
    }

    private formatToPlaintext(sessionData: any): string {
        // Handle different formats (OpenAI-like, Claude-like)
        const messages = sessionData.messages || sessionData.conversation || [];
        if (!Array.isArray(messages)) return "";

        return messages.map((m: any) => {
            const role = m.role || 'unknown';
            const content = typeof m.content === 'string' ? m.content : JSON.stringify(m.content);
            return `${role.toUpperCase()}: ${content}`;
        }).join("\n\n---\n\n");
    }

    private async extractValuableMemories(transcript: string, title: string) {
        const prompt = `
            Analyze the following session transcript titled "${title}".
            Identify the MOST VALUABLE pieces of knowledge, decisions made, or technical discoveries.
            Ignore small talk, errors, or redundant steps.
            
            Return a JSON array of strings, each being a concise "Memory" to be stored long-term.
            Example: ["The project uses port 4000 for the API", "Model X performs better on task Y"]

            TRANSCRIPT:
            ${transcript.slice(0, 8000)}
        `;

        try {
            const model = await this.llmService.modelSelector.selectModel({ taskComplexity: 'medium' });
            const response = await this.llmService.generateText(model.provider, model.modelId, "Memory Extraction", prompt);
            
            let memories: string[] = [];
            try {
                memories = JSON.parse(response.content.match(/\[.*\]/s)?.[0] || "[]");
            } catch {
                memories = response.content.split('\n').filter(l => l.startsWith('-')).map(l => l.substring(1).trim());
            }

            for (const memory of memories) {
                await this.agentMemory.add(memory, 'long_term', 'project', {
                    source: 'archiver_extraction',
                    sessionTitle: title
                });
            }
            
            console.log(`[MemoryArchiver] Extracted ${memories.length} memories from "${title}"`);
        } catch (e: any) {
            console.warn(`[MemoryArchiver] Memory extraction failed: ${e.message}`);
        }
    }

    private async ensureArchiveExists(): Promise<string> {
        const dir = path.dirname(this.archivePath);
        try {
            await fs.access(dir);
        } catch {
            await fs.mkdir(dir, { recursive: true });
        }

        try {
            await fs.access(this.archivePath);
        } catch {
            const zip = new AdmZip();
            zip.writeZip(this.archivePath);
        }
        return this.archivePath;
    }
}
