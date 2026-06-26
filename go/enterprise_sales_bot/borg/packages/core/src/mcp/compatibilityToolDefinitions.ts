import type { Tool } from '@modelcontextprotocol/sdk/types.js';

type CompatibilityToolName =
    | 'run_code'
    | 'run_python'
    | 'run_agent'
    | 'save_memory'
    | 'search_memory'
    | 'save_script'
    | 'save_tool_set'
    | 'load_tool_set'
    | 'toolset_list'
    | 'import_mcp_config'
    | 'auto_call_tool';

interface CompatibilityToolDefinitionOverrides {
    descriptions?: Partial<Record<CompatibilityToolName, string>>;
}

const baseDefinitions: Record<CompatibilityToolName, Tool> = {
    run_code: {
        name: 'run_code',
        description: 'Execute TypeScript/JavaScript code in a secure sandbox. Use this to chain multiple tool calls, process data, or perform logic. You can call other tools from within this code using `await mcp.call(\'tool_name\', args)`.',
        inputSchema: {
            type: 'object',
            properties: {
                code: {
                    type: 'string',
                    description: 'The TypeScript/JavaScript code to execute. Top-level await is supported.',
                },
            },
            required: ['code'],
        },
    },
    run_python: {
        name: 'run_python',
        description: 'Execute Python 3 code. Suitable for data processing or simple scripts. No direct tool calling integration yet (use run_code for tool chaining).',
        inputSchema: {
            type: 'object',
            properties: {
                code: {
                    type: 'string',
                    description: 'The Python 3 code to execute.',
                },
            },
            required: ['code'],
        },
    },
    run_agent: {
        name: 'run_agent',
        description: 'Run an autonomous AI agent to perform a task. The agent will analyze your request, find relevant tools, write its own code, and execute it.',
        inputSchema: {
            type: 'object',
            properties: {
                task: {
                    type: 'string',
                    description: 'The natural language description of the task (e.g., \'Find the latest issue in repo X and summarize it\').',
                },
                policyId: {
                    type: 'string',
                    description: 'Optional UUID of a Policy to restrict the agent\'s tool access.',
                },
            },
            required: ['task'],
        },
    },
    save_memory: {
        name: 'save_memory',
        description: 'Save a fact, instruction, or knowledge snippet to your long-term memory. Use this to remember user preferences, project details, or learnings for future sessions.',
        inputSchema: {
            type: 'object',
            properties: {
                content: {
                    type: 'string',
                    description: 'The information to remember.',
                },
                type: {
                    type: 'string',
                    enum: ['working', 'long_term'],
                    description: 'Type of memory. \'long_term\' persists across sessions. \'working\' is for the current task context.',
                    default: 'long_term',
                },
            },
            required: ['content'],
        },
    },
    search_memory: {
        name: 'search_memory',
        description: 'Search your memory for relevant facts or context.',
        inputSchema: {
            type: 'object',
            properties: {
                query: {
                    type: 'string',
                    description: 'The search query.',
                },
                limit: {
                    type: 'number',
                    description: 'Max results.',
                    default: 5,
                },
            },
            required: ['query'],
        },
    },
    save_script: {
        name: 'save_script',
        description: 'Save a successful code snippet as a reusable tool (Saved Script). The script will be available as a tool in future sessions.',
        inputSchema: {
            type: 'object',
            properties: {
                name: {
                    type: 'string',
                    description: 'The name of the new tool (must be unique, alphanumeric).',
                },
                description: {
                    type: 'string',
                    description: 'Description of what this script does.',
                },
                code: {
                    type: 'string',
                    description: 'The code to save.',
                },
            },
            required: ['name', 'code'],
        },
    },
    save_tool_set: {
        name: 'save_tool_set',
        description: 'Save the currently loaded tools as a \"Tool Set\" (Profile). This allows you to restore this working environment later.',
        inputSchema: {
            type: 'object',
            properties: {
                name: {
                    type: 'string',
                    description: 'Name of the tool set (e.g., \'web_dev\', \'data_analysis\').',
                },
                description: {
                    type: 'string',
                    description: 'Description of the tool set.',
                },
            },
            required: ['name'],
        },
    },
    load_tool_set: {
        name: 'load_tool_set',
        description: 'Load a previously saved Tool Set (Profile). This will add all tools in the set to your current context.',
        inputSchema: {
            type: 'object',
            properties: {
                name: {
                    type: 'string',
                    description: 'Name of the tool set to load.',
                },
            },
            required: ['name'],
        },
    },
    toolset_list: {
        name: 'toolset_list',
        description: 'List saved Tool Sets (Profiles) available to the current session.',
        inputSchema: {
            type: 'object',
            properties: {},
        },
    },
    import_mcp_config: {
        name: 'import_mcp_config',
        description: 'Import MCP servers from a JSON configuration file content (e.g., claude_desktop_config.json).',
        inputSchema: {
            type: 'object',
            properties: {
                configJson: {
                    type: 'string',
                    description: 'The content of the JSON configuration file.',
                },
            },
            required: ['configJson'],
        },
    },
    auto_call_tool: {
        name: 'auto_call_tool',
        description: 'One-shot discovery and execution. Automatically searches for the best tool to accomplish an objective, maps the parameters using the best available schema and metadata, and executes it immediately.',
        inputSchema: {
            type: 'object',
            properties: {
                objective: {
                    type: 'string',
                    description: 'The objective or task you want to accomplish using a tool.',
                },
                context: {
                    type: 'string',
                    description: 'Optional variables, file paths, snippets, or structured notes required to fill tool arguments.',
                },
            },
            required: ['objective'],
        },
    },
};

const compatibilityToolOrder: CompatibilityToolName[] = [
    'run_code',
    'run_python',
    'run_agent',
    'save_memory',
    'search_memory',
    'save_script',
    'save_tool_set',
    'load_tool_set',
    'toolset_list',
    'import_mcp_config',
    'auto_call_tool',
];

export function getCompatibilityToolDefinitions(
    overrides: CompatibilityToolDefinitionOverrides = {},
): Tool[] {
    return compatibilityToolOrder.map((name) => ({
        ...baseDefinitions[name],
        description: overrides.descriptions?.[name] ?? baseDefinitions[name].description,
    }));
}
