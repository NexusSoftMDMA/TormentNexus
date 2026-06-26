import { accessDeniedResult, jsonResult, requireString } from './helpers.js';
import { ContextToolModule } from './types.js';

export const removeTool: ContextToolModule = {
  name: 'remove',
  definition: {
    namespacedName: 'remove',
    originalName: 'remove',
    connectorId: 'mvmt',
    sourceId: 'mvmt',
    requiredAction: 'write',
    toolKind: 'semantic',
    description: 'Use only when the user explicitly asks to delete a specific permitted text file. Protected paths are always blocked.',
    inputSchema: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'Path such as /workspace/old-note.md' },
      },
      required: ['path'],
    },
  },
  async handle(args, { index, access }) {
    const inputPath = requireString(args.path, 'path');
    const mountName = index.mountNameForPath(inputPath);
    if (!mountName || !access.pathAllowed(inputPath, 'write')) {
      return accessDeniedResult(`missing_permission path=${inputPath} action=write`);
    }
    return jsonResult(await index.remove(inputPath));
  },
};
