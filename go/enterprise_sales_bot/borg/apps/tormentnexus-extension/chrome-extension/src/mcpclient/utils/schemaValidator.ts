import { z } from 'zod';
import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('ToolSchemaValidator');

/**
 * Standard MCP Tool input schema validation.
 * Ensures the schema is a valid JSON Schema object.
 */
const McpToolInputSchemaValidator = z.object({
  type: z.literal('object'),
  properties: z.record(z.string(), z.any()).optional(),
  required: z.array(z.string()).optional(),
}).passthrough();

export function validateToolSchema(toolName: string, schema: unknown): boolean {
  try {
    // Empty objects or missing schemas are valid for tools with no mandatory arguments
    if (!schema || (typeof schema === 'object' && Object.keys(schema).length === 0)) {
      return true;
    }
    
    McpToolInputSchemaValidator.parse(schema);
    return true;
  } catch (error) {
    logger.error(`[SchemaValidator] Tool '${toolName}' rejected due to malformed schema:`, error);
    // In a real scenario we could emit an event here to notify the UI that a tool was rejected
    if (typeof window !== 'undefined' && window.dispatchEvent) {
       window.dispatchEvent(new CustomEvent('mcp:tool-rejected', {
           detail: { toolName, error }
       }));
    }
    return false;
  }
}
