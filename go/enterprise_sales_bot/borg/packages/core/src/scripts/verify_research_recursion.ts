
import { DeepResearchService } from '../services/DeepResearchService.js';
import { LLMService } from '@tormentnexus/ai';
import { MemoryManager } from '../services/MemoryManager.js';

type RecursiveResearchResult = {
    topic: string;
    summary: string;
    sources: Array<{ title: string; url: string; keyPoints: string[] }>;
    relatedTopics: string[];
    subTopics?: RecursiveResearchResult[];
};

// Mock dependencies
const mockLLM = {
    modelSelector: { selectModel: async () => ({ provider: 'mock', modelId: 'mock-model' }) },
    generateText: async (p: any, m: any, sys: any, prompt: string) => {
        if (prompt.includes("Generate 3 specific web search queries")) {
            return { content: "Topic Analysis\nLatest News\nDeep Dive" };
        }
        if (prompt.includes("Summarize the key findings")) {
            // Return JSON structure
            return {
                content: JSON.stringify({
                    topic: "Mock Topic",
                    summary: "This is a summary.",
                    sources: [{ title: "Source 1", url: "http://example.com" }],
                    relatedTopics: ["Subtopic A", "Subtopic B"]
                })
            };
        }
        return { content: "" };
    }
} as unknown as LLMService;

const mockMemory = {
    saveContext: async () => 123
} as unknown as MemoryManager;

const mockSearch = {} as unknown as import('@tormentnexus/search').SearchService;

async function verifyRecursion() {
    console.log("🔍 Verifying Recursive Research...");

    const service = new DeepResearchService({ executeTool: async () => ({}) as any }, mockLLM, mockSearch, mockMemory);

    // Override internal researchTopic to avoid actual web calls/import issues during simple test
    // We just want to test recursion logic
    service.researchTopic = async (topic: string, depth: number) => {
        console.log(`  -> Researching: ${topic} (Depth: ${depth})`);
        return {
            topic,
            summary: `Summary of ${topic}`,
            sources: [],
            relatedTopics: depth > 1 ? [`${topic}-Sub1`, `${topic}-Sub2`] : []
        };
    };

    console.log("--- Starting Recursive Run (Depth 2, Breadth 2) ---");
    const result = await service.recursiveResearch("Root", 2, 2) as RecursiveResearchResult;

    console.log("\n--- Result Tree ---");
    printTree(result);

    if (result.subTopics && result.subTopics.length === 2 && result.subTopics[0].subTopics?.length === 2) {
        console.log("\n✅ Recursion Logic Verified!");
    } else {
        console.error(`\n❌ Recursion Logic Failed (Structure mismatch). Expected 2 children each with 2 children. Got children: ${result.subTopics?.length}, Grandchild[0]: ${result.subTopics?.[0]?.subTopics?.length}`);
        process.exit(1);
    }
}

function printTree(node: any, level = 0) {
    const indent = "  ".repeat(level);
    console.log(`${indent}- ${node.topic} (Related: ${node.relatedTopics.length})`);
    if (node.subTopics) {
        node.subTopics.forEach((sub: any) => printTree(sub, level + 1));
    }
}

verifyRecursion().catch(console.error);
