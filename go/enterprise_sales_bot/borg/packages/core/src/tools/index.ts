/**
 * Tools Index - Exports all MCP tools
 */

// LSP Tools (Phase 51)
export { LSPTools, LSP_TOOL_DEFINITIONS } from './LSPTools.js';
export {
    FindSymbolSchema,
    FindReferencesSchema,
    GoToDefinitionSchema,
    GetSymbolsSchema,
    RenameSymbolSchema,
    SearchSymbolsSchema,
} from './LSPTools.js';

// Existing Tools
export { DiagnosticTools } from './DiagnosticTools.js';
export { WebSearchTool } from './WebSearchTool.js';
