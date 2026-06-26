/**
 * MCP Tools v2.0 - Minimal API
 * 8 funzioni essenziali: add_node, search_nodes, delete_node, register_skill, suggest_skills, save_context_snapshot, restore_context, get_memory_report
 */

import memoryService from '../services/memory.js';
import { z } from 'zod';

/**
 * Tool: add_node
 * Aggiunge un nodo alla memoria
 */
async function addNode({ keywords, content, type, parentId, metadata, weight }) {
  try {
    const result = await memoryService.addNode({
      keywords: keywords || [],
      content: content || '',
      type: type || 'generic',
      parentId: parentId || null,
      metadata: metadata || {},
      weight: weight || 1.0
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: search_nodes
 * Cerca nodi per keywords
 */
async function searchNodes({ keywords, maxResults, minConfidence, type, sessionId }) {
  try {
    const result = await memoryService.searchNodes({
      keywords: keywords || [],
      maxResults: maxResults || 10,
      minConfidence: minConfidence || 0.1,
      type: type || null,
      sessionId: sessionId || null
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: delete_node
 * Elimina un nodo
 */
async function deleteNode({ nodeId, cascade }) {
  try {
    const result = await memoryService.deleteNode({
      nodeId,
      cascade: cascade || false
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: register_skill
 * Registra una skill con schema rigido (v2.0)
 * 
 * Schema richiesto:
 * - name: nome della skill
 * - framework: framework (es. "fastify", "react", "prisma")
 * - language: linguaggio (es. "javascript", "typescript", "python")
 * - filePattern: pattern file (es. "*.service.js", "**\/*.controller.ts")
 * - learnSteps: array di passi per imparare la skill
 * - useCases: array di casi d'uso
 * - implementation: implementazione dettagliata (opzionale)
 * - examples: esempi pratici (opzionale)
 * - prerequisites: prerequisiti (opzionale)
 * - keywords: keywords aggiuntive (opzionale)
 */
async function registerSkill({ name, framework, language, filePattern, learnSteps, useCases, implementation, examples, prerequisites, keywords, content }) {
  try {
    const result = await memoryService.registerSkill({
      name,
      framework: framework || '',
      language: language || '',
      filePattern: filePattern || '',
      learnSteps: learnSteps || [],
      useCases: useCases || [],
      implementation: implementation || '',
      examples: examples || [],
      prerequisites: prerequisites || [],
      keywords: keywords || [],
      content: content || ''
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: suggest_skills
 * Suggerisci skills basate su contesto
 */
async function suggestSkills({ currentKeywords, maxResults, domain }) {
  try {
    const result = await memoryService.suggestSkills({
      currentKeywords: currentKeywords || [],
      maxResults: maxResults || 5,
      domain: domain || null
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: save_context_snapshot
 * Salva snapshot contesto dettagliato
 */
async function saveContextSnapshot({ summary, workDone, pendingTasks, keyDecisions, blockers, learnings, nextSteps }) {
  try {
    const result = await memoryService.saveContextSnapshot({
      summary: summary || '',
      workDone: workDone || {},
      pendingTasks: pendingTasks || [],
      keyDecisions: keyDecisions || [],
      blockers: blockers || [],
      learnings: learnings || [],
      nextSteps: nextSteps || []
    });
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: restore_context
 * Recupera contesto compresso
 */
async function restoreContext({ snapshotId }) {
  try {
    const result = await memoryService.restoreContext(snapshotId);
    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * Tool: get_memory_report
 * Genera report memoria
 */
async function getMemoryReport({ format, keywords, includeStats, includeRecentWork, includeTopSkills }) {
  try {
    const result = await memoryService.getMemoryReport({
      format: format || 'json',
      keywords: keywords || [],
      includeStats: includeStats !== false,
      includeRecentWork: includeRecentWork !== false,
      includeTopSkills: includeTopSkills !== false
    });

    // Se il formato è HTML, restituisci come text
    if (format === 'html' && typeof result === 'string') {
      return { content: [{ type: 'text', text: result }] };
    }

    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: `Error: ${error.message}` }], isError: true };
  }
}

/**
 * TOOL HANDLERS MAP
 * Mappa tutti i tool handlers per il server MCP
 */
export const toolHandlers = {
  // === NODE MANAGEMENT ===
  add_node: addNode,
  search_nodes: searchNodes,
  delete_node: deleteNode,

  // === SKILLS FRAMEWORK ===
  register_skill: registerSkill,
  suggest_skills: suggestSkills,

  // === CONTEXT MANAGEMENT ===
  save_context_snapshot: saveContextSnapshot,
  restore_context: restoreContext,

  // === REPORTS ===
  get_memory_report: getMemoryReport
};

/**
 * TOOL DEFININITIONS
 * Definizioni JSON schema per ogni tool (per capability协商)
 */
export const toolDefinitions = [
  // === NODE MANAGEMENT ===
  {
    name: 'add_node',
    description: 'Aggiunge un nodo alla memoria neurale',
    inputSchema: {
      type: 'object',
      properties: {
        keywords: { type: 'array', items: { type: 'string' }, description: 'Keywords per ricerca' },
        content: { type: 'string', description: 'Contenuto long-text' },
        type: { type: 'string', enum: ['task', 'entity', 'file', 'concept', 'summary', 'action', 'generic', 'skill', 'error', 'edge_case', 'operation', 'convention', 'pattern', 'context_snapshot'], default: 'generic' },
        parentId: { type: 'string', description: 'ID nodo padre per gerarchia' },
        metadata: { type: 'object', description: 'Dati aggiuntivi' },
        weight: { type: 'number', default: 1.0 }
      },
      required: ['keywords']
    }
  },
  {
    name: 'search_nodes',
    description: 'Cerca nodi per keywords con confidence scoring',
    inputSchema: {
      type: 'object',
      properties: {
        keywords: { type: 'array', items: { type: 'string' } },
        maxResults: { type: 'number', default: 10 },
        minConfidence: { type: 'number', default: 0.1 },
        type: { type: 'string' },
        sessionId: { type: 'string' }
      },
      required: ['keywords']
    }
  },
  {
    name: 'delete_node',
    description: 'Elimina un nodo',
    inputSchema: {
      type: 'object',
      properties: {
        nodeId: { type: 'string' },
        cascade: { type: 'boolean', default: false, description: 'Elimina anche i figli' }
      },
      required: ['nodeId']
    }
  },

  // === SKILLS FRAMEWORK ===
  {
    name: 'register_skill',
    description: 'Registra una SKILL con schema rigido obbligatorio. Schema: name (obbligatorio), framework, language, filePattern, learnSteps[], useCases[].',
    inputSchema: {
      type: 'object',
      properties: {
        name: { type: 'string', description: 'Nome della skill (obbligatorio)' },
        framework: { type: 'string', description: 'Framework (es. fastify, react, prisma)' },
        language: { type: 'string', description: 'Linguaggio (es. javascript, typescript)' },
        filePattern: { type: 'string', description: 'Pattern file (es. *.service.js)' },
        learnSteps: { type: 'array', items: { type: 'string' }, description: 'Passi per imparare' },
        useCases: { type: 'array', items: { type: 'string' }, description: 'Casi d\'uso' },
        implementation: { type: 'string', description: 'Implementazione dettagliata' },
        examples: { type: 'array', items: { type: 'string' } },
        prerequisites: { type: 'array', items: { type: 'string' } },
        keywords: { type: 'array', items: { type: 'string' } },
        content: { type: 'string' }
      },
      required: ['name']
    }
  },
  {
    name: 'suggest_skills',
    description: 'Suggerisce skills rilevanti basate sul contesto corrente',
    inputSchema: {
      type: 'object',
      properties: {
        currentKeywords: { type: 'array', items: { type: 'string' } },
        maxResults: { type: 'number', default: 5 },
        domain: { type: 'string' }
      },
      required: ['currentKeywords']
    }
  },

  // === CONTEXT MANAGEMENT ===
  {
    name: 'save_context_snapshot',
    description: 'Salva snapshot dettagliato del contesto di lavoro',
    inputSchema: {
      type: 'object',
      properties: {
        summary: { type: 'string' },
        workDone: { type: 'object' },
        pendingTasks: { type: 'array', items: { type: 'string' } },
        keyDecisions: { type: 'array', items: { type: 'string' } },
        blockers: { type: 'array', items: { type: 'string' } },
        learnings: { type: 'array', items: { type: 'string' } },
        nextSteps: { type: 'array', items: { type: 'string' } }
      }
    }
  },
  {
    name: 'restore_context',
    description: 'Recupera un contesto salvato',
    inputSchema: {
      type: 'object',
      properties: {
        snapshotId: { type: 'string' }
      },
      required: ['snapshotId']
    }
  },
  // === REPORTS ===
  {
    name: 'get_memory_report',
    description: 'Genera report della memoria in formato JSON o HTML',
    inputSchema: {
      type: 'object',
      properties: {
        format: { type: 'string', enum: ['json', 'html'], default: 'json' },
        keywords: { type: 'array', items: { type: 'string' } },
        includeStats: { type: 'boolean', default: true },
        includeRecentWork: { type: 'boolean', default: true },
        includeTopSkills: { type: 'boolean', default: true }
      }
    }
  }
];

/**
 * Registra tutti i tool su un server McpServer
 * @param {McpServer} server - Istanza del server MCP
 * @returns {Array} Lista dei tool registrati
 */
export function registerAllTools(server) {
  const registeredTools = [];

  for (const tool of toolDefinitions) {
    const zodShape = convertJsonSchemaToZodShape(tool.inputSchema);

    server.tool(tool.name, tool.description, zodShape, async (args) => {
      try {
        const handler = toolHandlers[tool.name];
        if (!handler) {
          return { content: [{ type: "text", text: `Handler not found for tool: ${tool.name}` }], isError: true };
        }
        return await handler(args);
      } catch (error) {
        return { content: [{ type: "text", text: `Error: ${error.message}` }], isError: true };
      }
    });

    registeredTools.push({
      name: tool.name,
      description: tool.description,
      inputSchema: tool.inputSchema,
    });
  }

  return registeredTools;
}

/**
 * Helper: converte un JSON schema molto semplice in una shape Zod (oggetto di schemi)
 */
function convertJsonSchemaToZodShape(schema) {
  if (!schema || !schema.properties) {
    return {};
  }

  const shape = {};
  for (const [key, prop] of Object.entries(schema.properties)) {
    let zType;
    switch (prop.type) {
      case 'string':
        zType = z.string();
        if (prop.enum) zType = z.enum(prop.enum);
        break;
      case 'number':
        zType = z.number();
        break;
      case 'boolean':
        zType = z.boolean();
        break;
      case 'array':
        zType = z.array(prop.items?.type === 'string' ? z.string() : z.any());
        break;
      case 'object':
        zType = z.record(z.any());
        break;
      default:
        zType = z.any();
    }

    if (prop.description) {
      zType = zType.describe(prop.description);
    }

    if (schema.required && schema.required.includes(key)) {
      shape[key] = zType;
    } else {
      shape[key] = zType.optional();
    }
  }

  return shape;
}
