/**
 * LSP Tools - MCP Tools for Language Server Protocol Operations
 * 
 * Provides MCP tool definitions for LSP operations, enabling LLMs to:
 * - Find symbols by name
 * - Find references to symbols
 * - Go to definition
 * - Get all symbols in a file
 * - Rename symbols across the project
 */

import { z } from 'zod';
import { LSPService, SymbolInformation, SymbolKind, Location } from '../services/LSPService.js';

// Schema definitions for tool inputs
export const FindSymbolSchema = z.object({
    file_path: z.string().describe('Relative path to the file'),
    symbol_name: z.string().describe('Name of the symbol to find'),
});

export const FindReferencesSchema = z.object({
    file_path: z.string().describe('Relative path to the file'),
    line: z.number().describe('Line number (0-indexed)'),
    character: z.number().describe('Character position (0-indexed)'),
});

export const GoToDefinitionSchema = z.object({
    file_path: z.string().describe('Relative path to the file'),
    line: z.number().describe('Line number (0-indexed)'),
    character: z.number().describe('Character position (0-indexed)'),
});

export const GetSymbolsSchema = z.object({
    file_path: z.string().describe('Relative path to the file'),
});

export const RenameSymbolSchema = z.object({
    file_path: z.string().describe('Relative path to the file'),
    line: z.number().describe('Line number (0-indexed)'),
    character: z.number().describe('Character position (0-indexed)'),
    new_name: z.string().describe('New name for the symbol'),
});

export const SearchSymbolsSchema = z.object({
    query: z.string().describe('Search query for symbol names'),
});

// Helper to convert SymbolKind to string
const symbolKindToString = (kind: SymbolKind): string => {
    const kindMap: Record<SymbolKind, string> = {
        [SymbolKind.File]: 'file',
        [SymbolKind.Module]: 'module',
        [SymbolKind.Namespace]: 'namespace',
        [SymbolKind.Package]: 'package',
        [SymbolKind.Class]: 'class',
        [SymbolKind.Method]: 'method',
        [SymbolKind.Property]: 'property',
        [SymbolKind.Field]: 'field',
        [SymbolKind.Constructor]: 'constructor',
        [SymbolKind.Enum]: 'enum',
        [SymbolKind.Interface]: 'interface',
        [SymbolKind.Function]: 'function',
        [SymbolKind.Variable]: 'variable',
        [SymbolKind.Constant]: 'constant',
        [SymbolKind.String]: 'string',
        [SymbolKind.Number]: 'number',
        [SymbolKind.Boolean]: 'boolean',
        [SymbolKind.Array]: 'array',
        [SymbolKind.Object]: 'object',
        [SymbolKind.Key]: 'key',
        [SymbolKind.Null]: 'null',
        [SymbolKind.EnumMember]: 'enum_member',
        [SymbolKind.Struct]: 'struct',
        [SymbolKind.Event]: 'event',
        [SymbolKind.Operator]: 'operator',
        [SymbolKind.TypeParameter]: 'type_parameter',
    };
    return kindMap[kind] || 'unknown';
};

// Format symbol for output
const formatSymbol = (symbol: SymbolInformation): string => {
    const loc = symbol.location;
    const range = loc.range;
    return `${symbol.name} (${symbolKindToString(symbol.kind)}) at ${loc.uri.replace('file://', '')}:${range.start.line + 1}:${range.start.character + 1}`;
};

// Format location for output
const formatLocation = (loc: Location): string => {
    const range = loc.range;
    return `${loc.uri.replace('file://', '')}:${range.start.line + 1}:${range.start.character + 1}`;
};

/**
 * LSP Tools class - provides MCP tool implementations
 */
export class LSPTools {
    private lspService: LSPService | null = null;

    constructor(private rootPath: string) { }

    private async getService(): Promise<LSPService> {
        if (!this.lspService) {
            this.lspService = new LSPService(this.rootPath);
        }
        return this.lspService;
    }

    /**
     * Find a symbol by name in a file
     */
    async findSymbol(input: z.infer<typeof FindSymbolSchema>): Promise<string> {
        const service = await this.getService();
        const symbol = await service.findSymbol(input.file_path, input.symbol_name);

        if (!symbol) {
            return `Symbol "${input.symbol_name}" not found in ${input.file_path}`;
        }

        return formatSymbol(symbol);
    }

    /**
     * Find all references to a symbol at a position
     */
    async findReferences(input: z.infer<typeof FindReferencesSchema>): Promise<string> {
        const service = await this.getService();
        const references = await service.findReferences(input.file_path, input.line, input.character);

        if (references.length === 0) {
            return `No references found at ${input.file_path}:${input.line + 1}:${input.character + 1}`;
        }

        const formatted = references.map(formatLocation);
        return `Found ${references.length} references:\n${formatted.join('\n')}`;
    }

    /**
     * Go to the definition of a symbol at a position
     */
    async goToDefinition(input: z.infer<typeof GoToDefinitionSchema>): Promise<string> {
        const service = await this.getService();
        const definition = await service.goToDefinition(input.file_path, input.line, input.character);

        if (!definition) {
            return `No definition found at ${input.file_path}:${input.line + 1}:${input.character + 1}`;
        }

        return `Definition: ${formatLocation(definition)}`;
    }

    /**
     * Get all symbols in a file
     */
    async getSymbols(input: z.infer<typeof GetSymbolsSchema>): Promise<string> {
        const service = await this.getService();
        const symbols = await service.getSymbols(input.file_path);

        if (symbols.length === 0) {
            return `No symbols found in ${input.file_path}`;
        }

        // Group by kind
        const grouped = new Map<string, SymbolInformation[]>();
        for (const symbol of symbols) {
            const kind = symbolKindToString(symbol.kind);
            if (!grouped.has(kind)) {
                grouped.set(kind, []);
            }
            grouped.get(kind)!.push(symbol);
        }

        const output: string[] = [`Symbols in ${input.file_path} (${symbols.length} total):`];
        for (const [kind, items] of grouped) {
            output.push(`\n## ${kind}s (${items.length})`);
            for (const item of items) {
                const line = item.location.range.start.line + 1;
                output.push(`  - ${item.name} (line ${line})`);
            }
        }

        return output.join('\n');
    }

    /**
     * Rename a symbol across the project
     */
    async renameSymbol(input: z.infer<typeof RenameSymbolSchema>): Promise<string> {
        const service = await this.getService();
        const result = await service.renameSymbol(
            input.file_path,
            input.line,
            input.character,
            input.new_name
        );

        if (!result) {
            return `Failed to rename symbol at ${input.file_path}:${input.line + 1}:${input.character + 1}`;
        }

        // WorkspaceEdit has documentChanges or changes
        const edit = result as { changes?: Record<string, unknown[]>; documentChanges?: unknown[] };

        if (edit.changes) {
            const fileCount = Object.keys(edit.changes).length;
            const editCount = Object.values(edit.changes).reduce((sum, edits) => sum + (edits as unknown[]).length, 0);
            return `Renamed to "${input.new_name}" with ${editCount} edit(s) across ${fileCount} file(s)`;
        }

        if (edit.documentChanges) {
            return `Renamed to "${input.new_name}" with ${edit.documentChanges.length} document change(s)`;
        }

        return `Renamed to "${input.new_name}"`;
    }

    /**
     * Search symbols by name across indexed files
     */
    async searchSymbols(input: z.infer<typeof SearchSymbolsSchema>): Promise<string> {
        const service = await this.getService();

        // First, ensure project is indexed
        await service.indexProject();

        const symbols = service.searchSymbols(input.query);

        if (symbols.length === 0) {
            return `No symbols found matching "${input.query}"`;
        }

        const formatted = symbols.slice(0, 50).map(formatSymbol);
        const moreCount = symbols.length > 50 ? `\n... and ${symbols.length - 50} more` : '';

        return `Found ${symbols.length} symbols matching "${input.query}":\n${formatted.join('\n')}${moreCount}`;
    }

    /**
     * Shutdown the LSP service
     */
    async shutdown(): Promise<void> {
        if (this.lspService) {
            await this.lspService.shutdown();
            this.lspService = null;
        }
    }
}

// Tool definitions for MCP
export const LSP_TOOL_DEFINITIONS = [
    {
        name: 'find_symbol',
        description: 'Find a symbol (function, class, variable, etc.) by name in a file. Returns the symbol location and type.',
        inputSchema: FindSymbolSchema,
    },
    {
        name: 'find_references',
        description: 'Find all references to a symbol at a specific position. Use this to see everywhere a function, class, or variable is used.',
        inputSchema: FindReferencesSchema,
    },
    {
        name: 'go_to_definition',
        description: 'Go to the definition of a symbol at a specific position. Use this to find where a function, class, or variable is defined.',
        inputSchema: GoToDefinitionSchema,
    },
    {
        name: 'get_symbols',
        description: 'Get all symbols (functions, classes, variables, etc.) in a file. Returns a structured list grouped by symbol type.',
        inputSchema: GetSymbolsSchema,
    },
    {
        name: 'rename_symbol',
        description: 'Rename a symbol across the entire project. This performs a semantic rename, updating all references.',
        inputSchema: RenameSymbolSchema,
    },
    {
        name: 'search_symbols',
        description: 'Search for symbols by name across the entire project. First indexes the project, then searches for matching symbols.',
        inputSchema: SearchSymbolsSchema,
    },
];

export default LSPTools;
