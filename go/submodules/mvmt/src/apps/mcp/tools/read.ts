import { accessDeniedResult, jsonResult, requireString } from './helpers.js';
import { ContextToolModule } from './types.js';

export const readTool: ContextToolModule = {
  name: 'read',
  definition: {
    namespacedName: 'read',
    originalName: 'read',
    connectorId: 'mvmt',
    sourceId: 'mvmt',
    requiredAction: 'read',
    toolKind: 'semantic',
    description: 'Use after search or list when you have a specific permitted file path and need the full file contents.',
    inputSchema: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'Path such as /workspace/README.md' },
      },
      required: ['path'],
    },
  },
  async handle(args, { index, access }) {
    const inputPath = requireString(args.path, 'path');
    const mountName = index.mountNameForPath(inputPath);
    if (!mountName || !access.pathAllowed(inputPath, 'read')) {
      return accessDeniedResult(`missing_permission path=${inputPath} action=read`);
    }
    return jsonResult(await index.read(inputPath));
  },
};
