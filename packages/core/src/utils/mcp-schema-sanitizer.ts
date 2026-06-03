/**
 * Recursively replaces $ref and $defs/definitions pointers with their literal definitions
 * and prunes deep or circular structures to satisfy Gemini API constraints.
 */
export function sanitizeSchema(schema: any, maxDepth = 10, currentDepth = 0, defs: any = {}): any {
    if (currentDepth > maxDepth) {
        return { type: 'string', description: '[Max Depth Exceeded]' };
    }

    if (!schema || typeof schema !== 'object') {
        return schema;
    }

    // Capture definitions from this level if they exist
    const localDefs = schema.$defs || schema.definitions || {};
    const currentDefs = { ...defs, ...localDefs };

    // Clone the schema but remove the definition blocks to keep it clean
    let result = Array.isArray(schema) ? [...schema] : { ...schema };
    
    if (typeof result === 'object' && !Array.isArray(result)) {
        delete result.$defs;
        delete result.definitions;

        // Resolve $ref if present
        if (result.$ref) {
            const refPath = result.$ref;
            const refKey = refPath.split('/').pop();
            if (currentDefs[refKey]) {
                // Important: we don't increment depth here as we are resolving a pointer to a sibling/parent
                // but we should be careful about infinite recursion.
                // We'll increment depth anyway to be safe against circular refs.
                return sanitizeSchema(currentDefs[refKey], maxDepth, currentDepth + 1, currentDefs);
            }
            // If we can't resolve it, Gemini will likely fail, but we've done our best.
            // We could replace it with a generic object.
            return { type: 'object', description: `[Unresolved $ref: ${refPath}]`, additionalProperties: true };
        }

        // Recursively sanitize all properties
        const processed: any = {};
        for (const [key, value] of Object.entries(result)) {
            processed[key] = sanitizeSchema(value, maxDepth, currentDepth + 1, currentDefs);
        }
        return processed;
    }

    if (Array.isArray(result)) {
        return result.map(item => sanitizeSchema(item, maxDepth, currentDepth + 1, currentDefs));
    }

    return result;
}

/**
 * Specifically flattens a tool's inputSchema to ensure no $ref or $defs remain.
 */
export function flattenToolSchema(tool: any): any {
    if (!tool || !tool.inputSchema) return tool;
    return {
        ...tool,
        inputSchema: sanitizeSchema(tool.inputSchema)
    };
}
