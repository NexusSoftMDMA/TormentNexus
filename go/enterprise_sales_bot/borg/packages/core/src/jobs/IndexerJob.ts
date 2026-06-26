
import { MemoryManager } from "../services/MemoryManager.js";

export class IndexerJob {
    private intervalId: NodeJS.Timeout | null = null;
    private isRunning: boolean = false;
    private isIndexing: boolean = false;

    constructor(
        private memoryManager: MemoryManager,
        private rootDir: string = process.cwd(),
        private intervalMs: number = 300000 // 5 minutes
    ) { }

    public start() {
        if (this.isRunning) return;
        this.isRunning = true;
        console.log(`[IndexerJob] Started (Interval: ${this.intervalMs}ms)`);

        // Initial run immediately
        this.runIndex();

        this.intervalId = setInterval(() => {
            this.runIndex();
        }, this.intervalMs);
    }

    public stop() {
        if (!this.isRunning) return;
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
        this.isRunning = false;
        console.log(`[IndexerJob] Stopped.`);
    }

    private async runIndex() {
        if (this.isIndexing) {
            console.log(`[IndexerJob] Skip: Indexing already in progress.`);
            return;
        }

        try {
            this.isIndexing = true;
            console.log(`[IndexerJob] 🔄 Running Scheduled Indexing...`);

            // Index Symbols
            const symbolsCount = await this.memoryManager.indexSymbols(this.rootDir);
            console.log(`[IndexerJob] ✅ Indexed ${symbolsCount} symbols.`);

            // Optionally Index Text (Full Codebase)
            // const filesCount = await this.memoryManager.indexCodebase(this.rootDir);
            // console.log(`[IndexerJob] ✅ Indexed ${filesCount} files.`);

        } catch (e: any) {
            console.error(`[IndexerJob] ❌ Failed: ${e.message}`);
        } finally {
            this.isIndexing = false;
        }
    }

    public getStatus() {
        return {
            running: this.isRunning,
            indexing: this.isIndexing
        };
    }
}
