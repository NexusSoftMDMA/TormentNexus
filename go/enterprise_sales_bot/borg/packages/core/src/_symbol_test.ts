
import { MemoryManager } from './services/MemoryManager.js';
import path from 'path';

async function main() {
    console.log("Testing Symbol Indexer...");

    const mem = new MemoryManager(process.cwd());
    const targetDir = path.join(process.cwd(), 'packages/core/src/services');

    console.log(`Target: ${targetDir}`);

    try {
        const count = await mem.indexSymbols(targetDir);
        console.log(`Indexed ${count} symbols.`);

        console.log("Searching for 'MemoryManager'...");
        const results = await mem.searchSymbols("MemoryManager class");

        results.forEach(r => {
            console.log(`[${r.score.toFixed(2)}] ${r.metadata?.name || r.id}`);
            console.log(r.content.substring(0, 100).replace(/\n/g, ' '));
        });

    } catch (e: any) {
        console.error("Test Failed:", e);
    }
}

main().catch(console.error);
