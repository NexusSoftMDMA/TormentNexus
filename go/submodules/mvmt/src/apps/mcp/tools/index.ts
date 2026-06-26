import { listTool } from './list.js';
import { readTool } from './read.js';
import { removeTool } from './remove.js';
import { searchTool } from './search.js';
import { ContextToolModule, ContextToolName } from './types.js';
import { writeTool } from './write.js';

export const CONTEXT_TOOLS: ContextToolModule[] = [
  searchTool,
  listTool,
  readTool,
  writeTool,
  removeTool,
];

export const CONTEXT_TOOL_BY_NAME = new Map(
  CONTEXT_TOOLS.map((tool) => [tool.name, tool]),
);

export function isContextToolName(value: string): value is ContextToolName {
  return CONTEXT_TOOL_BY_NAME.has(value as ContextToolName);
}

export type { ContextToolName, NamespacedTool, PermissionAction } from './types.js';
