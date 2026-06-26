import { IAgent, LLMService } from "@tormentnexus/ai";
import { PromptRegistry } from "../prompts/PromptRegistry.js";
import { ClaudeAgent } from "./ClaudeAgent.js";

export class MetaArchitectAgent extends ClaudeAgent {
    constructor(llmService: LLMService, promptRegistry: PromptRegistry) {
        super(llmService, promptRegistry);
    }

    // Override send to inject specialized system prompt
    async send(message: string, context?: any): Promise<string> {
        let systemPrompt = `You are the META-ARCHITECT of the TormentNexus System.
Your purpose is SELF-EVOLUTION. You verify modifications to the system itself.
- You analyze code structure for cleanliness, modularity, and safety.
- You design new tools and capabilities.
- You are EXTREMELY conservative about breaking changes.
- You prioritize interfaces (SOLID principles) and type safety.

When reviewing or writing code, ensure it aligns with the 'packages/' monorepo structure.`;

        // Prepend the meta-system prompt if it's the start of a conversation, 
        // or just rely on the Director/Council to pass strict instructions.
        // For now, let's just use the super.send but maybe prefix the message?
        // Actually, the Council usually constructs the system prompt in 'consult'.
        // Functioning primarily as a distinct subsystem marker. Specific tool availability can be enforced here.

        return super.send(message, context);
    }
}
