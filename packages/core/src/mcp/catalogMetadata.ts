type CatalogToolSeed = {
    name: string;
    title?: string | null;
    description?: string | null;
    inputSchema?: Record<string, unknown> | null;
    alwaysOn?: boolean;
};

type CatalogServerSeed = {
    serverName: string;
    description?: string | null;
    alwaysOn?: boolean;
    tools: CatalogToolSeed[];
};

export type DerivedServerCatalogMetadata = {
    serverName: string;
    serverDisplayName: string;
    serverTags: string[];
    alwaysOn: boolean;
};

export type DerivedToolCatalogMetadata = {
    name: string;
    title?: string | null;
    description?: string | null;
    inputSchema?: Record<string, unknown> | null;
    alwaysOn: boolean;
    serverName: string;
    serverDisplayName: string;
    serverTags: string[];
    toolTags: string[];
    semanticGroup: string;
    semanticGroupLabel: string;
    advertisedName: string;
    keywords: string[];
};

export type DerivedServerCatalog = DerivedServerCatalogMetadata & {
    tools: DerivedToolCatalogMetadata[];
};

type TagRule = {
    tag: string;
    keywords: string[];
    semanticGroup: string;
    semanticLabel: string;
};

const CAPABILITY_RULES: TagRule[] = [
    {
        tag: 'browser',
        keywords: ['browser', 'tab', 'page', 'dom', 'playwright', 'screenshot', 'viewport', 'click', 'hover', 'type', 'navigate', 'url'],
        semanticGroup: 'browser-automation',
        semanticLabel: 'browser automation',
    },
    {
        tag: 'filesystem',
        keywords: ['file', 'folder', 'directory', 'path', 'read_file', 'write', 'mkdir', 'patch', 'workspace'],
        semanticGroup: 'filesystem-operations',
        semanticLabel: 'filesystem operations',
    },
    {
        tag: 'terminal',
        keywords: ['terminal', 'shell', 'command', 'powershell', 'process', 'stdout', 'stderr', 'task runner'],
        semanticGroup: 'terminal-execution',
        semanticLabel: 'terminal execution',
    },
    {
        tag: 'git',
        keywords: ['git', 'commit', 'branch', 'pull request', 'worktree', 'diff', 'merge', 'repository'],
        semanticGroup: 'source-control',
        semanticLabel: 'source control',
    },
    {
        tag: 'github',
        keywords: ['github', 'issue', 'pull request', 'pr', 'repo', 'workflow', 'actions'],
        semanticGroup: 'github-collaboration',
        semanticLabel: 'GitHub collaboration',
    },
    {
        tag: 'search',
        keywords: ['search', 'find', 'grep', 'query', 'lookup', 'discover', 'rank', 'semantic'],
        semanticGroup: 'search-discovery',
        semanticLabel: 'search and discovery',
    },
    {
        tag: 'memory',
        keywords: ['memory', 'recall', 'remember', 'knowledge', 'context', 'notes', 'summary'],
        semanticGroup: 'memory-context',
        semanticLabel: 'memory and context',
    },
    {
        tag: 'documentation',
        keywords: ['docs', 'documentation', 'markdown', 'readme', 'pdf', 'article', 'content', 'text extract'],
        semanticGroup: 'documentation-content',
        semanticLabel: 'documentation and content',
    },
    {
        tag: 'network',
        keywords: ['http', 'https', 'fetch', 'request', 'api', 'webhook', 'endpoint', 'sse', 'websocket'],
        semanticGroup: 'network-api',
        semanticLabel: 'network and API',
    },
    {
        tag: 'database',
        keywords: ['database', 'db', 'sql', 'table', 'query', 'record', 'row', 'schema'],
        semanticGroup: 'data-storage',
        semanticLabel: 'data storage',
    },
    {
        tag: 'data',
        keywords: ['json', 'csv', 'yaml', 'xml', 'parse', 'transform', 'dataset'],
        semanticGroup: 'data-processing',
        semanticLabel: 'data processing',
    },
    {
        tag: 'agent',
        keywords: ['agent', 'subagent', 'autonomous', 'workflow', 'orchestrate', 'delegate'],
        semanticGroup: 'agent-orchestration',
        semanticLabel: 'agent orchestration',
    },
    {
        tag: 'ui',
        keywords: ['ui', 'dashboard', 'screen', 'page', 'view', 'dialog', 'notebook'],
        semanticGroup: 'ui-surface',
        semanticLabel: 'UI surface',
    },
];

const ACTION_RULES = [
    { tag: 'read', keywords: ['read', 'view', 'list', 'show', 'inspect', 'summarize', 'get'] },
    { tag: 'write', keywords: ['write', 'save', 'create', 'insert', 'append', 'record'] },
    { tag: 'update', keywords: ['update', 'edit', 'modify', 'patch', 'rename', 'hydrate'] },
    { tag: 'delete', keywords: ['delete', 'remove', 'clear', 'destroy', 'unload'] },
    { tag: 'execute', keywords: ['run', 'execute', 'start', 'launch', 'build', 'compile', 'test'] },
    { tag: 'analyze', keywords: ['analyze', 'rank', 'score', 'diagnose', 'debug', 'explain'] },
];

const GENERIC_TOKENS = new Set([
    'tool',
    'tools',
    'server',
    'mcp',
    'tormentnexus',
    'meta',
    'data',
    'service',
    'function',
    'functions',
]);

function normalizeText(value: string | null | undefined): string {
    return value?.trim().toLowerCase() ?? '';
}

function splitIdentifier(value: string | null | undefined): string[] {
    return normalizeText(value)
        .replace(/__/g, ' ')
        .replace(/[^a-z0-9]+/g, ' ')
        .split(/\s+/)
        .filter(Boolean);
}

function unique<T>(values: T[]): T[] {
    return Array.from(new Set(values));
}

function titleCase(value: string): string {
    return value
        .split(/[-_\s]+/)
        .filter(Boolean)
        .map((token) => token.charAt(0).toUpperCase() + token.slice(1))
        .join(' ');
}

function collectSearchText(server: CatalogServerSeed, tool?: CatalogToolSeed): string {
    const schemaKeys = tool?.inputSchema && typeof tool.inputSchema === 'object'
        ? Object.keys(tool.inputSchema)
        : [];

    return [
        server.serverName,
        server.description,
        tool?.name,
        tool?.title,
        tool?.description,
        ...schemaKeys,
    ]
        .filter((value): value is string => typeof value === 'string' && value.length > 0)
        .join(' ')
        .toLowerCase();
}

function scoreRuleMatches(text: string, keywords: string[]): number {
    return keywords.reduce((score, keyword) => (
        text.includes(keyword.toLowerCase()) ? score + 1 : score
    ), 0);
}

function inferCapabilityTags(text: string): string[] {
    return CAPABILITY_RULES
        .map((rule) => ({ rule, score: scoreRuleMatches(text, rule.keywords) }))
        .filter((entry) => entry.score > 0)
        .sort((left, right) => right.score - left.score || left.rule.tag.localeCompare(right.rule.tag))
        .map((entry) => entry.rule.tag)
        .slice(0, 4);
}

function inferActionTags(text: string): string[] {
    return ACTION_RULES
        .filter((rule) => rule.keywords.some((keyword) => text.includes(keyword)))
        .map((rule) => rule.tag)
        .slice(0, 3);
}

function fallbackTagsFromIdentifier(...values: Array<string | null | undefined>): string[] {
    return unique(values.flatMap((value) => splitIdentifier(value)))
        .filter((token) => token.length > 2 && !GENERIC_TOKENS.has(token))
        .slice(0, 3);
}

function selectSemanticGroup(tags: string[]): { key: string; label: string } {
    for (const tag of tags) {
        const rule = CAPABILITY_RULES.find((candidate) => candidate.tag === tag);
        if (rule) {
            return {
                key: rule.semanticGroup,
                label: rule.semanticLabel,
            };
        }
    }

    return {
        key: 'general-utility',
        label: 'general utility',
    };
}

function buildServerDisplayName(serverName: string, serverTags: string[]): string {
    if (serverTags.length === 0) {
        return serverName;
    }

    return `${serverName} [${serverTags.slice(0, 2).join(', ')}]`;
}

function buildAdvertisedToolName(serverDisplayName: string, toolName: string, toolTags: string[]): string {
    if (toolTags.length === 0) {
        return `${serverDisplayName} -> ${toolName}`;
    }

    return `${serverDisplayName} -> ${toolName} [${toolTags.slice(0, 2).join(', ')}]`;
}

function buildKeywords(serverName: string, serverTags: string[], toolName: string, toolTags: string[], semanticLabel: string): string[] {
    return unique([
        ...splitIdentifier(serverName),
        ...serverTags,
        ...splitIdentifier(toolName),
        ...toolTags,
        ...splitIdentifier(semanticLabel),
    ]).slice(0, 12);
}

export function deriveSemanticCatalogForServer(server: CatalogServerSeed): DerivedServerCatalog {
    const serverText = [
        server.serverName,
        server.description,
        ...server.tools.flatMap((tool) => [tool.name, tool.title, tool.description]),
    ]
        .filter((value): value is string => typeof value === 'string' && value.length > 0)
        .join(' ')
        .toLowerCase();

    const inferredServerTags = inferCapabilityTags(serverText);
    const serverTags = inferredServerTags.length > 0
        ? inferredServerTags
        : fallbackTagsFromIdentifier(server.serverName, server.description);
    const serverDisplayName = buildServerDisplayName(server.serverName, serverTags);

    const tools = server.tools.map((tool) => {
        const toolText = collectSearchText(server, tool);
        const toolTags = unique([
            ...inferCapabilityTags(toolText),
            ...inferActionTags(toolText),
            ...serverTags.slice(0, 2),
        ]);
        const stabilizedToolTags = toolTags.length > 0
            ? toolTags.slice(0, 6)
            : fallbackTagsFromIdentifier(tool.name, tool.description, server.serverName);
        const semanticGroup = selectSemanticGroup(stabilizedToolTags);

        return {
            name: tool.name,
            title: tool.title ?? null,
            description: tool.description ?? null,
            inputSchema: tool.inputSchema ?? null,
            alwaysOn: Boolean(tool.alwaysOn || server.alwaysOn),
            serverName: server.serverName,
            serverDisplayName,
            serverTags,
            toolTags: stabilizedToolTags,
            semanticGroup: semanticGroup.key,
            semanticGroupLabel: semanticGroup.label,
            advertisedName: buildAdvertisedToolName(serverDisplayName, tool.name, stabilizedToolTags),
            keywords: buildKeywords(server.serverName, serverTags, tool.name, stabilizedToolTags, semanticGroup.label),
        } satisfies DerivedToolCatalogMetadata;
    });

    return {
        serverName: server.serverName,
        serverDisplayName,
        serverTags,
        alwaysOn: Boolean(server.alwaysOn),
        tools,
    };
}

export function describeSemanticGroup(groupKey: string): string {
    if (!groupKey) {
        return 'general utility';
    }

    const capabilityRule = CAPABILITY_RULES.find((rule) => rule.semanticGroup === groupKey);
    return capabilityRule?.semanticLabel ?? titleCase(groupKey);
}
