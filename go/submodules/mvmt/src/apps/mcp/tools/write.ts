import { accessDeniedResult, jsonResult, optionalString, requireString, requireText } from './helpers.js';
import { ContextToolModule } from './types.js';

export const writeTool: ContextToolModule = {
  name: 'write',
  definition: {
    namespacedName: 'write',
    originalName: 'write',
    connectorId: 'mvmt',
    sourceId: 'mvmt',
    requiredAction: 'write',
    toolKind: 'semantic',
    description: 'Use only when the user explicitly asks to create or overwrite a specific permitted text file. Optionally pass expected_hash to avoid stale writes.',
    inputSchema: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'Path such as /workspace/notes.md' },
        content: { type: 'string', description: 'Full file content to write' },
        expected_hash: { type: 'string', description: 'Optional SHA-256 hash from a previous read' },
      },
      required: ['path', 'content'],
    },
  },
  async handle(args, { index, access }) {
    const inputPath = requireString(args.path, 'path');
    const content = requireText(args.content, 'content');
    const expectedHash = optionalString(args.expected_hash);
    const mountName = index.mountNameForPath(inputPath);
    if (!mountName || !access.pathAllowed(inputPath, 'write')) {
      return accessDeniedResult(`missing_permission path=${inputPath} action=write`);
    }
    return jsonResult(await index.write(inputPath, content, expectedHash));
  },
};
