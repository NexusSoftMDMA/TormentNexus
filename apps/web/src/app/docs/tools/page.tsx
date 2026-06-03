"use client";
import React, { useState } from 'react';
import Link from 'next/link';

type ToolCategory = {
    id: string;
    title: string;
    icon: string;
    description: string;
    tools: Tool[];
};

type Tool = {
    name: string;
    description: string;
    parameters?: { name: string; type: string; required: boolean; description: string }[];
    example?: string;
    returns?: string;
};

const TOOL_CATEGORIES: ToolCategory[] = [
    {
        id: 'lsp',
        title: 'LSP Tools',
        icon: '🔍',
        description: 'Language Server Protocol integration for semantic code understanding',
        tools: [
            {
                name: 'find_symbol',
                description: 'Find a symbol definition in the codebase',
                parameters: [
                    { name: 'symbol', type: 'string', required: true, description: 'Symbol name to find' },
                    { name: 'path', type: 'string', required: false, description: 'Optional path to search in' }
                ],
                example: 'find_symbol({ symbol: "MCPServer" })',
                returns: 'Location of the symbol definition'
            },
            {
                name: 'find_references',
                description: 'Find all references to a symbol',
                parameters: [
                    { name: 'symbol', type: 'string', required: true, description: 'Symbol to find references for' },
                    { name: 'path', type: 'string', required: true, description: 'File path containing the symbol' }
                ],
                example: 'find_references({ symbol: "executeTool", path: "src/MCPServer.ts" })',
                returns: 'Array of locations where the symbol is referenced'
            },
            {
                name: 'go_to_definition',
                description: 'Navigate to the definition of a symbol',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'File path' },
                    { name: 'line', type: 'number', required: true, description: 'Line number' },
                    { name: 'column', type: 'number', required: true, description: 'Column number' }
                ],
                example: 'go_to_definition({ path: "src/index.ts", line: 10, column: 5 })',
                returns: 'Definition location'
            },
            {
                name: 'get_symbols',
                description: 'Get all symbols in a file',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'File path' }
                ],
                example: 'get_symbols({ path: "src/services/LSPService.ts" })',
                returns: 'Array of symbol information'
            },
            {
                name: 'rename_symbol',
                description: 'Rename a symbol across the codebase',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'File path' },
                    { name: 'line', type: 'number', required: true, description: 'Line number' },
                    { name: 'column', type: 'number', required: true, description: 'Column number' },
                    { name: 'newName', type: 'string', required: true, description: 'New name for the symbol' }
                ],
                example: 'rename_symbol({ path: "src/utils.ts", line: 5, column: 10, newName: "newFunctionName" })',
                returns: 'List of files modified'
            },
            {
                name: 'search_symbols',
                description: 'Search for symbols by pattern',
                parameters: [
                    { name: 'query', type: 'string', required: true, description: 'Search query' },
                    { name: 'limit', type: 'number', required: false, description: 'Max results (default: 20)' }
                ],
                example: 'search_symbols({ query: "Service", limit: 10 })',
                returns: 'Matching symbols'
            }
        ]
    },
    {
        id: 'plan',
        title: 'Plan/Build Tools',
        icon: '📝',
        description: 'Structured development workflow with Plan mode (review) and Build mode (execute)',
        tools: [
            {
                name: 'plan_mode',
                description: 'Enter Plan mode - changes are staged but not applied',
                parameters: [],
                example: 'plan_mode()',
                returns: 'Confirmation of Plan mode activation'
            },
            {
                name: 'build_mode',
                description: 'Enter Build mode - changes are applied immediately',
                parameters: [],
                example: 'build_mode()',
                returns: 'Confirmation of Build mode activation'
            },
            {
                name: 'propose_change',
                description: 'Propose a code change (staged in Plan mode)',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'File path' },
                    { name: 'original', type: 'string', required: true, description: 'Original content' },
                    { name: 'modified', type: 'string', required: true, description: 'Modified content' },
                    { name: 'description', type: 'string', required: false, description: 'Change description' }
                ],
                example: 'propose_change({ path: "src/utils.ts", original: "const x = 1;", modified: "const x = 2;", description: "Update constant" })',
                returns: 'Diff ID for the proposed change'
            },
            {
                name: 'review_changes',
                description: 'Review all staged changes',
                parameters: [],
                example: 'review_changes()',
                returns: 'List of staged diffs with details'
            },
            {
                name: 'approve_change',
                description: 'Approve a specific staged change',
                parameters: [
                    { name: 'diff_id', type: 'string', required: true, description: 'ID of the diff to approve' }
                ],
                example: 'approve_change({ diff_id: "diff-abc123" })',
                returns: 'Confirmation'
            },
            {
                name: 'apply_changes',
                description: 'Apply all approved changes to the filesystem',
                parameters: [],
                example: 'apply_changes()',
                returns: 'List of applied changes'
            },
            {
                name: 'plan_status',
                description: 'Get current mode and pending changes status',
                parameters: [],
                example: 'plan_status()',
                returns: 'Current mode, staged count, approved count'
            },
            {
                name: 'create_checkpoint',
                description: 'Create a rollback checkpoint',
                parameters: [
                    { name: 'name', type: 'string', required: true, description: 'Checkpoint name' },
                    { name: 'description', type: 'string', required: false, description: 'Description' }
                ],
                example: 'create_checkpoint({ name: "before-refactor", description: "Pre-refactor state" })',
                returns: 'Checkpoint ID'
            },
            {
                name: 'rollback',
                description: 'Rollback to a checkpoint',
                parameters: [
                    { name: 'checkpoint_id', type: 'string', required: true, description: 'Checkpoint ID' }
                ],
                example: 'rollback({ checkpoint_id: "cp-abc123" })',
                returns: 'Rollback result'
            }
        ]
    },
    {
        id: 'memory',
        title: 'Memory Tools',
        icon: '🧠',
        description: 'Tiered memory system for persistent agent context',
        tools: [
            {
                name: 'add_memory',
                description: 'Store content in the memory system',
                parameters: [
                    { name: 'content', type: 'string', required: true, description: 'Memory content' },
                    { name: 'type', type: 'session|working|long_term', required: false, description: 'Memory tier (default: working)' },
                    { name: 'namespace', type: 'user|agent|project', required: false, description: 'Memory namespace (default: project)' },
                    { name: 'tags', type: 'string[]', required: false, description: 'Tags for categorization' }
                ],
                example: 'add_memory({ content: "User prefers TypeScript", type: "long_term", namespace: "user" })',
                returns: 'Memory ID'
            },
            {
                name: 'search_memory',
                description: 'Search memories by content similarity (TF-IDF)',
                parameters: [
                    { name: 'query', type: 'string', required: true, description: 'Search query' },
                    { name: 'type', type: 'string', required: false, description: 'Filter by tier' },
                    { name: 'namespace', type: 'string', required: false, description: 'Filter by namespace' },
                    { name: 'limit', type: 'number', required: false, description: 'Max results (default: 10)' }
                ],
                example: 'search_memory({ query: "TypeScript preferences", limit: 5 })',
                returns: 'Matching memories with scores'
            },
            {
                name: 'get_recent_memories',
                description: 'Get recently accessed memories',
                parameters: [
                    { name: 'limit', type: 'number', required: false, description: 'Max results (default: 10)' },
                    { name: 'type', type: 'string', required: false, description: 'Filter by tier' }
                ],
                example: 'get_recent_memories({ limit: 5 })',
                returns: 'Recent memories'
            },
            {
                name: 'memory_stats',
                description: 'Get memory system statistics',
                parameters: [],
                example: 'memory_stats()',
                returns: 'Total counts by tier, namespace, access patterns'
            },
            {
                name: 'clear_session_memory',
                description: 'Clear all ephemeral session memories',
                parameters: [],
                example: 'clear_session_memory()',
                returns: 'Confirmation'
            }
        ]
    },
    {
        id: 'workflow',
        title: 'Workflow Tools',
        icon: '⚙️',
        description: 'Graph-based workflow orchestration with human-in-the-loop support',
        tools: [
            {
                name: 'run_workflow',
                description: 'Execute a registered workflow',
                parameters: [
                    { name: 'workflow_id', type: 'string', required: true, description: 'ID of the workflow' },
                    { name: 'input', type: 'object', required: false, description: 'Input data for the workflow' }
                ],
                example: 'run_workflow({ workflow_id: "code-review-v1", input: { code: "function foo() {}" } })',
                returns: 'Execution result with run ID and state'
            },
            {
                name: 'list_workflows',
                description: 'List all workflow executions',
                parameters: [],
                example: 'list_workflows()',
                returns: 'List of execution IDs, statuses, current nodes'
            },
            {
                name: 'workflow_status',
                description: 'Get detailed status of a workflow run',
                parameters: [
                    { name: 'run_id', type: 'string', required: true, description: 'Run ID' }
                ],
                example: 'workflow_status({ run_id: "abc123" })',
                returns: 'Workflow ID, status, current node, last updated'
            },
            {
                name: 'approve_workflow',
                description: 'Approve or reject a workflow at a HITL checkpoint',
                parameters: [
                    { name: 'run_id', type: 'string', required: true, description: 'Run ID' },
                    { name: 'approved', type: 'boolean', required: false, description: 'Approve (true) or reject (false)' }
                ],
                example: 'approve_workflow({ run_id: "abc123", approved: true })',
                returns: 'Confirmation'
            }
        ]
    },
    {
        id: 'codemode',
        title: 'Code Mode Tools',
        icon: '🖥️',
        description: 'Execute code to call tools instead of structured JSON (94% context reduction)',
        tools: [
            {
                name: 'execute_code',
                description: 'Execute JavaScript/TypeScript in a sandboxed environment',
                parameters: [
                    { name: 'code', type: 'string', required: true, description: 'Code to execute' },
                    { name: 'context', type: 'object', required: false, description: 'Additional context variables' }
                ],
                example: `execute_code({ code: \`
  const files = await list_files("src");
  console.log("Found", files.length, "files");
  return files;
\` })`,
                returns: 'Execution result, output, tools called'
            },
            {
                name: 'enable_code_mode',
                description: 'Enable Code Mode for tool calling via code',
                parameters: [],
                example: 'enable_code_mode()',
                returns: 'Confirmation'
            },
            {
                name: 'disable_code_mode',
                description: 'Disable Code Mode',
                parameters: [],
                example: 'disable_code_mode()',
                returns: 'Confirmation'
            },
            {
                name: 'code_mode_status',
                description: 'Get Code Mode status and context reduction stats',
                parameters: [],
                example: 'code_mode_status()',
                returns: 'Enabled status, tool count, context reduction percentage'
            },
            {
                name: 'list_code_tools',
                description: 'List tools available in Code Mode',
                parameters: [],
                example: 'list_code_tools()',
                returns: 'List of tool names and descriptions'
            }
        ]
    },
    {
        id: 'filesystem',
        title: 'Filesystem Tools',
        icon: '📁',
        description: 'File system operations for reading, writing, and managing files',
        tools: [
            {
                name: 'read_file',
                description: 'Read the contents of a file',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Path to the file' },
                    { name: 'encoding', type: 'string', required: false, description: 'File encoding (default: utf-8)' }
                ],
                example: 'read_file({ path: "src/index.ts" })',
                returns: 'File contents as string'
            },
            {
                name: 'write_file',
                description: 'Write content to a file (creates or overwrites)',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Path to the file' },
                    { name: 'content', type: 'string', required: true, description: 'Content to write' }
                ],
                example: 'write_file({ path: "output.txt", content: "Hello World" })',
                returns: 'Confirmation with bytes written'
            },
            {
                name: 'list_files',
                description: 'List files in a directory',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Directory path' },
                    { name: 'recursive', type: 'boolean', required: false, description: 'Include subdirectories' },
                    { name: 'pattern', type: 'string', required: false, description: 'Glob pattern to filter' }
                ],
                example: 'list_files({ path: "src", recursive: true, pattern: "*.ts" })',
                returns: 'Array of file paths'
            },
            {
                name: 'create_directory',
                description: 'Create a new directory',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Directory path' },
                    { name: 'recursive', type: 'boolean', required: false, description: 'Create parent directories' }
                ],
                example: 'create_directory({ path: "new/nested/dir", recursive: true })',
                returns: 'Confirmation'
            },
            {
                name: 'delete_file',
                description: 'Delete a file or directory',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Path to delete' },
                    { name: 'recursive', type: 'boolean', required: false, description: 'Delete directories recursively' }
                ],
                example: 'delete_file({ path: "temp.txt" })',
                returns: 'Confirmation'
            },
            {
                name: 'copy_file',
                description: 'Copy a file or directory',
                parameters: [
                    { name: 'source', type: 'string', required: true, description: 'Source path' },
                    { name: 'destination', type: 'string', required: true, description: 'Destination path' }
                ],
                example: 'copy_file({ source: "template.ts", destination: "new-file.ts" })',
                returns: 'Confirmation'
            },
            {
                name: 'move_file',
                description: 'Move/rename a file or directory',
                parameters: [
                    { name: 'source', type: 'string', required: true, description: 'Source path' },
                    { name: 'destination', type: 'string', required: true, description: 'Destination path' }
                ],
                example: 'move_file({ source: "old-name.ts", destination: "new-name.ts" })',
                returns: 'Confirmation'
            },
            {
                name: 'file_info',
                description: 'Get metadata about a file',
                parameters: [
                    { name: 'path', type: 'string', required: true, description: 'Path to the file' }
                ],
                example: 'file_info({ path: "package.json" })',
                returns: 'Size, modified date, type, permissions'
            }
        ]
    },
    {
        id: 'terminal',
        title: 'Terminal Tools',
        icon: '💻',
        description: 'Execute shell commands and manage terminal sessions',
        tools: [
            {
                name: 'run_command',
                description: 'Execute a shell command',
                parameters: [
                    { name: 'command', type: 'string', required: true, description: 'Command to execute' },
                    { name: 'cwd', type: 'string', required: false, description: 'Working directory' },
                    { name: 'timeout', type: 'number', required: false, description: 'Timeout in ms' }
                ],
                example: 'run_command({ command: "npm test", cwd: "packages/core" })',
                returns: 'Exit code, stdout, stderr'
            },
            {
                name: 'spawn_process',
                description: 'Start a long-running background process',
                parameters: [
                    { name: 'command', type: 'string', required: true, description: 'Command to execute' },
                    { name: 'name', type: 'string', required: false, description: 'Process name/ID' }
                ],
                example: 'spawn_process({ command: "npm run dev", name: "dev-server" })',
                returns: 'Process ID'
            },
            {
                name: 'kill_process',
                description: 'Terminate a running process',
                parameters: [
                    { name: 'pid', type: 'string', required: true, description: 'Process ID or name' }
                ],
                example: 'kill_process({ pid: "dev-server" })',
                returns: 'Confirmation'
            },
            {
                name: 'list_processes',
                description: 'List running managed processes',
                parameters: [],
                example: 'list_processes()',
                returns: 'Array of process info (pid, name, status, uptime)'
            }
        ]
    },
    {
        id: 'search',
        title: 'Search Tools',
        icon: '🔎',
        description: 'Code search, grep, and semantic search capabilities',
        tools: [
            {
                name: 'grep',
                description: 'Search for patterns in files (ripgrep)',
                parameters: [
                    { name: 'pattern', type: 'string', required: true, description: 'Search pattern (regex)' },
                    { name: 'path', type: 'string', required: false, description: 'Directory to search' },
                    { name: 'filePattern', type: 'string', required: false, description: 'File glob pattern' }
                ],
                example: 'grep({ pattern: "TODO:", path: "src", filePattern: "*.ts" })',
                returns: 'Array of matches with file, line, content'
            },
            {
                name: 'semantic_search',
                description: 'AI-powered semantic code search',
                parameters: [
                    { name: 'query', type: 'string', required: true, description: 'Natural language query' },
                    { name: 'limit', type: 'number', required: false, description: 'Max results' }
                ],
                example: 'semantic_search({ query: "authentication middleware", limit: 5 })',
                returns: 'Ranked results with relevance scores'
            },
            {
                name: 'find_files',
                description: 'Find files by name pattern',
                parameters: [
                    { name: 'pattern', type: 'string', required: true, description: 'File name pattern' },
                    { name: 'path', type: 'string', required: false, description: 'Directory to search' }
                ],
                example: 'find_files({ pattern: "*.test.ts", path: "src" })',
                returns: 'Array of matching file paths'
            },
            {
                name: 'web_search',
                description: 'Search the web for information',
                parameters: [
                    { name: 'query', type: 'string', required: true, description: 'Search query' },
                    { name: 'limit', type: 'number', required: false, description: 'Max results' }
                ],
                example: 'web_search({ query: "TypeScript generics tutorial", limit: 5 })',
                returns: 'Search results with title, url, snippet'
            }
        ]
    },
    {
        id: 'git',
        title: 'Git Tools',
        icon: '🌿',
        description: 'Version control operations and repository management',
        tools: [
            {
                name: 'git_status',
                description: 'Get the status of the git repository',
                parameters: [],
                example: 'git_status()',
                returns: 'Modified, staged, untracked files'
            },
            {
                name: 'git_diff',
                description: 'Show changes in the working directory',
                parameters: [
                    { name: 'path', type: 'string', required: false, description: 'Specific file to diff' },
                    { name: 'staged', type: 'boolean', required: false, description: 'Show staged changes' }
                ],
                example: 'git_diff({ path: "src/index.ts" })',
                returns: 'Unified diff output'
            },
            {
                name: 'git_commit',
                description: 'Create a commit with staged changes',
                parameters: [
                    { name: 'message', type: 'string', required: true, description: 'Commit message' }
                ],
                example: 'git_commit({ message: "feat: add new feature" })',
                returns: 'Commit hash'
            },
            {
                name: 'git_log',
                description: 'Show commit history',
                parameters: [
                    { name: 'limit', type: 'number', required: false, description: 'Number of commits' },
                    { name: 'path', type: 'string', required: false, description: 'Filter by file path' }
                ],
                example: 'git_log({ limit: 10 })',
                returns: 'Array of commits (hash, message, author, date)'
            },
            {
                name: 'git_branch',
                description: 'List, create, or switch branches',
                parameters: [
                    { name: 'action', type: 'list|create|switch|delete', required: true, description: 'Branch action' },
                    { name: 'name', type: 'string', required: false, description: 'Branch name (for create/switch/delete)' }
                ],
                example: 'git_branch({ action: "create", name: "feature/new-feature" })',
                returns: 'Branch list or confirmation'
            },
            {
                name: 'git_stash',
                description: 'Stash or restore uncommitted changes',
                parameters: [
                    { name: 'action', type: 'push|pop|list', required: true, description: 'Stash action' },
                    { name: 'message', type: 'string', required: false, description: 'Stash message (for push)' }
                ],
                example: 'git_stash({ action: "push", message: "WIP: refactoring" })',
                returns: 'Confirmation or stash list'
            }
        ]
    }
];

export default function ToolsPage() {
    const [selectedCategory, setSelectedCategory] = useState<string>('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [expandedTool, setExpandedTool] = useState<string | null>(null);

    const filteredCategories = TOOL_CATEGORIES.filter(cat =>
        selectedCategory === 'all' || cat.id === selectedCategory
    ).map(cat => ({
        ...cat,
        tools: cat.tools.filter(tool =>
            tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            tool.description.toLowerCase().includes(searchQuery.toLowerCase())
        )
    })).filter(cat => cat.tools.length > 0);

    const totalTools = TOOL_CATEGORIES.reduce((sum, cat) => sum + cat.tools.length, 0);

    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4 sticky top-0 z-10">
                <div className="max-w-5xl mx-auto flex items-center justify-between">
                    <div>
                        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">🛠️ MCP Tools Reference</h1>
                        <p className="text-sm text-zinc-500">{totalTools} tools available</p>
                    </div>
                    <Link href="/docs" className="text-blue-500 hover:text-blue-400 font-medium">← Back to Docs</Link>
                </div>
            </header>

            <main className="max-w-5xl mx-auto px-6 py-8">
                {/* Filters */}
                <div className="flex flex-col md:flex-row gap-4 mb-8">
                    <input
                        type="text"
                        placeholder="Search tools..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="flex-1 px-4 py-2 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg text-zinc-900 dark:text-white placeholder-zinc-500 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <div className="flex gap-2 flex-wrap">
                        <button
                            onClick={() => setSelectedCategory('all')}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${selectedCategory === 'all'
                                ? 'bg-blue-500 text-white'
                                : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-200 dark:hover:bg-zinc-700'
                                }`}
                        >
                            All
                        </button>
                        {TOOL_CATEGORIES.map(cat => (
                            <button
                                key={cat.id}
                                onClick={() => setSelectedCategory(cat.id)}
                                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${selectedCategory === cat.id
                                    ? 'bg-blue-500 text-white'
                                    : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-200 dark:hover:bg-zinc-700'
                                    }`}
                            >
                                {cat.icon} {cat.title}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Tool Categories */}
                {filteredCategories.map(category => (
                    <section key={category.id} className="mb-10">
                        <div className="mb-4">
                            <h2 className="text-xl font-bold text-zinc-900 dark:text-white flex items-center gap-2">
                                <span>{category.icon}</span>
                                {category.title}
                                <span className="text-sm font-normal text-zinc-500">({category.tools.length} tools)</span>
                            </h2>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">{category.description}</p>
                        </div>

                        <div className="space-y-3">
                            {category.tools.map(tool => (
                                <article
                                    key={tool.name}
                                    className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg overflow-hidden"
                                >
                                    <button
                                        onClick={() => setExpandedTool(expandedTool === tool.name ? null : tool.name)}
                                        className="w-full px-5 py-4 flex items-center justify-between text-left hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
                                    >
                                        <div className="flex items-center gap-3">
                                            <code className="text-blue-500 font-mono font-bold">{tool.name}</code>
                                            <span className="text-zinc-600 dark:text-zinc-400 text-sm">{tool.description}</span>
                                        </div>
                                        <span className="text-zinc-400">{expandedTool === tool.name ? '−' : '+'}</span>
                                    </button>

                                    {expandedTool === tool.name && (
                                        <div className="px-5 pb-5 border-t border-zinc-200 dark:border-zinc-800">
                                            {/* Parameters */}
                                            {tool.parameters && tool.parameters.length > 0 && (
                                                <div className="mt-4">
                                                    <h4 className="text-sm font-bold text-zinc-500 uppercase mb-2">Parameters</h4>
                                                    <table className="w-full text-sm">
                                                        <thead>
                                                            <tr className="text-left text-zinc-500">
                                                                <th className="pb-2">Name</th>
                                                                <th className="pb-2">Type</th>
                                                                <th className="pb-2">Required</th>
                                                                <th className="pb-2">Description</th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            {tool.parameters.map(param => (
                                                                <tr key={param.name} className="border-t border-zinc-100 dark:border-zinc-800">
                                                                    <td className="py-2 font-mono text-blue-400">{param.name}</td>
                                                                    <td className="py-2 text-zinc-500">{param.type}</td>
                                                                    <td className="py-2">
                                                                        {param.required ? (
                                                                            <span className="text-red-400">required</span>
                                                                        ) : (
                                                                            <span className="text-zinc-400">optional</span>
                                                                        )}
                                                                    </td>
                                                                    <td className="py-2 text-zinc-600 dark:text-zinc-400">{param.description}</td>
                                                                </tr>
                                                            ))}
                                                        </tbody>
                                                    </table>
                                                </div>
                                            )}

                                            {/* Example */}
                                            {tool.example && (
                                                <div className="mt-4">
                                                    <h4 className="text-sm font-bold text-zinc-500 uppercase mb-2">Example</h4>
                                                    <pre className="bg-zinc-100 dark:bg-zinc-800 p-3 rounded text-sm overflow-x-auto">
                                                        <code className="text-zinc-700 dark:text-zinc-300">{tool.example}</code>
                                                    </pre>
                                                </div>
                                            )}

                                            {/* Returns */}
                                            {tool.returns && (
                                                <div className="mt-4">
                                                    <h4 className="text-sm font-bold text-zinc-500 uppercase mb-2">Returns</h4>
                                                    <p className="text-sm text-zinc-600 dark:text-zinc-400">{tool.returns}</p>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </article>
                            ))}
                        </div>
                    </section>
                ))}

                {filteredCategories.length === 0 && (
                    <div className="text-center py-16 text-zinc-500">
                        <p className="text-lg">No tools match your search.</p>
                        <button
                            onClick={() => { setSearchQuery(''); setSelectedCategory('all'); }}
                            className="mt-4 text-blue-500 hover:text-blue-400"
                        >
                            Clear filters
                        </button>
                    </div>
                )}

                {/* Footer */}
                <footer className="mt-12 pt-8 border-t border-zinc-200 dark:border-zinc-800 text-center text-sm text-zinc-500">
                    <p>TormentNexus Mission Control • {totalTools} MCP Tools</p>
                    <div className="mt-2 flex justify-center gap-4">
                        <Link href="/docs" className="text-blue-500 hover:text-blue-400">Feature Docs</Link>
                        <Link href="/docs/api" className="text-blue-500 hover:text-blue-400">API Reference</Link>
                        <Link href="/" className="text-blue-500 hover:text-blue-400">Dashboard</Link>
                    </div>
                </footer>
            </main>
        </div>
    );
}
