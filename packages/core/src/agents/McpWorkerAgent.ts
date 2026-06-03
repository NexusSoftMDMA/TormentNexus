import { SpecializedAgent } from '../mesh/SpecializedAgent.js';
import { LLMService } from '@tormentnexus/ai';
import { MCPServer } from '../MCPServer.js';

export class McpWorkerAgent extends SpecializedAgent {
    private llm: LLMService;
    private mcpServer: MCPServer;
    private maxSteps: number = 10;

    constructor(llm: LLMService, mcpServer: MCPServer) {
        // Expose capabilities bridging MCP tools and general reasoning
        super('McpWorker', ['mcp_worker', 'general_intelligence']);
        this.llm = llm;
        this.mcpServer = mcpServer;
    }

    public async handleTask(offer: any): Promise<any> {
        console.log(`[McpWorkerAgent] 🧠 Analyzing task: "${offer.task}"`);

        // 0. Predictive Tool Disclosure (Phase 112)
        // Preemptively fetch relevant tools based on task goal to reduce discovery turns
        const predictedToolNames = await this.mcpServer.getPredictedToolAds('', offer.task);
        if (predictedToolNames.length > 0) {
            console.log(`[McpWorkerAgent] 🔮 Sidecar predicted: ${predictedToolNames.join(', ')}`);
        }

        // 1. Resolve requested tools and mission-level policy (Phase 96)
        const requestedTools: string[] = Array.isArray(offer.tools) ? offer.tools : [];
        const policyAllow: string[] = Array.isArray(offer?.toolPolicy?.allow) ? offer.toolPolicy.allow : [];
        const policyDeny: string[] = Array.isArray(offer?.toolPolicy?.deny) ? offer.toolPolicy.deny : [];
        const deniedToolEvents: Array<{ tool: string; reason: string; timestamp: number }> = [];

        const nativeTools = await this.mcpServer.getNativeTools();

        const explicitAllowList = [...new Set([
            ...(requestedTools || []),
            ...(policyAllow || []),
            ...(predictedToolNames || [])
        ])];
        const hasExplicitAllowList = explicitAllowList.length > 0;
        const denySet = new Set(policyDeny);

        const availableTools = nativeTools.filter((tool) => {
            if (denySet.has(tool.name)) return false;
            if (!hasExplicitAllowList) return true;
            return explicitAllowList.includes(tool.name);
        });

        const toolDescriptions = availableTools.map(t => ({
            name: t.name,
            description: t.description,
            parameters: t.inputSchema
        }));

        let systemPrompt = `You are a Swarm Worker Agent with access to external tools.
Your task is: ${offer.task}

You have access to the following tools:
${JSON.stringify(toolDescriptions, null, 2)}

Mission Context (Global State shared across the swarm):
${JSON.stringify(offer.context || offer.missionContext || {}, null, 2)}

To use a tool, you MUST reply with a JSON object in this exact format, and nothing else:
{ "tool": "tool_name", "args": { ... } }

When you have completely solved the task, you MUST reply with a JSON object in this format:
{ "result": "your final detailed answer" }

If you discovered facts that other agents in this mission need to know, you can update the global context:
{ "result": "your final detailed answer", "_contextUpdate": { "key": "value" } }

Think step by step, but ALWAYS ensure your entire response is a single valid JSON object representing either a tool call or the final result.`;

        let history: { role: string, content: string }[] = [];
        let currentPrompt = "Begin execution.";
        let step = 0;

        while (step < this.maxSteps) {
            step++;
            console.log(`[McpWorkerAgent] 🔄 Step ${step}/${this.maxSteps}`);

            try {
                const model = await this.llm.modelSelector.selectModel({ taskComplexity: 'high', routingTaskType: 'worker' });
                const completion = await this.llm.generateText(model.provider, model.modelId, systemPrompt, currentPrompt, {
                    history,
                    taskComplexity: 'high',
                    routingTaskType: 'worker',
                });
                let responseText = completion.content.trim();

                // Strip markdown code blocks if present
                if (responseText.startsWith('\`\`\`json') && responseText.endsWith('\`\`\`')) {
                    responseText = responseText.substring(7, responseText.length - 3).trim();
                } else if (responseText.startsWith('\`\`\`') && responseText.endsWith('\`\`\`')) {
                    responseText = responseText.substring(3, responseText.length - 3).trim();
                }

                let action;
                try {
                    action = JSON.parse(responseText);
                } catch (e) {
                    console.warn(`[McpWorkerAgent] ⚠️ Invalid JSON from LLM: ${responseText}`);
                    history.push({ role: "assistant", content: responseText });
                    currentPrompt = "Your response MUST be a valid JSON object. Please fix your formatting and try again. E.g. { \"tool\": \"name\", \"args\": {} }";
                    continue;
                }

                if (action.result) {
                    console.log(`[McpWorkerAgent] ✅ Task Complete.`);

                    if (deniedToolEvents.length > 0) {
                        action._swarmTelemetry = {
                            ...(action._swarmTelemetry || {}),
                            deniedTools: deniedToolEvents
                        };
                    }

                    return action; // Returns { result: "...", _contextUpdate: "..." }
                } else if (action.tool) {
                    console.log(`[McpWorkerAgent] 🛠️ Calling tool ${action.tool}...`);
                    history.push({ role: "assistant", content: JSON.stringify(action) });

                    // Verify tool is in the allowed list
                    if (!availableTools.find(t => t.name === action.tool)) {
                        const deniedReason = denySet.has(action.tool)
                            ? `Tool '${action.tool}' denied by mission tool policy (deny list).`
                            : `Tool '${action.tool}' is not allowed for this task (allow list constraint).`;

                        deniedToolEvents.push({
                            tool: action.tool,
                            reason: deniedReason,
                            timestamp: Date.now()
                        });

                        currentPrompt = `${deniedReason} Available tools: ${availableTools.map(t => t.name).join(', ')}`;
                        continue;
                    }

                    try {
                        // Execute the tool locally via the MCP Server abstraction
                        let toolResult = await this.mcpServer.executeTool(action.tool, action.args);

                        let resultText = typeof toolResult === 'string' ? toolResult : JSON.stringify(toolResult);
                        // Truncate massive tool returns
                        if (resultText.length > 8000) {
                            resultText = resultText.substring(0, 8000) + "\n... [TRUNCATED]";
                        }

                        currentPrompt = `Tool returned:\n${resultText}`;
                    } catch (toolErr: any) {
                        console.error(`[McpWorkerAgent] ❌ Tool execution failed: ${toolErr.message}`);
                        currentPrompt = `Tool execution failed: ${toolErr.message}`;
                    }
                } else {
                    history.push({ role: "assistant", content: JSON.stringify(action) });
                    currentPrompt = "Unrecognized JSON format. You must specify either 'tool' or 'result'.";
                }

            } catch (error: any) {
                console.error(`[McpWorkerAgent] 💥 Fatal Loop Error:`, error);
                throw error;
            }
        }

        throw new Error(`McpWorkerAgent exhausted max steps (${this.maxSteps}) without returning a final {"result": "..."}.`);
    }
}
