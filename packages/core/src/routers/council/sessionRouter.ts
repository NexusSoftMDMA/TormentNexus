import { z } from 'zod';
import { t, publicProcedure } from '../../lib/trpc-core.js';
import { sessionManager } from '../../orchestrator/council/services/session-manager.js';
import { wsManager } from '../../orchestrator/council/services/ws-manager.js';
import { loadConfig } from '../../orchestrator/council/services/config.js';

const cliTypeSchema = z.enum([
  'tormentnexus', 'opencode', 'claude', 'aider', 'cursor', 'continue', 'cody', 'copilot', 'custom',
  'adrenaline', 'amazon-q', 'amazon-q-developer', 'amp-code', 'auggie', 'azure-openai',
  'bito', 'byterover', 'claude-code', 'code-codex', 'codebuff', 'codemachine', 'codex',
  'crush', 'dolt', 'factory', 'factory-droid', 'gemini', 'goose', 'grok', 'jules', 'kilo-code', 'kimi',
  'llm', 'litellm', 'llamafile', 'manus', 'mistral-vibe', 'ollama', 'open-interpreter',
  'pi', 'qwen-code', 'rowboatx', 'rovo', 'shell-pilot', 'smithery', 'superai-cli', 'trae', 'warp'
]);

const developmentTaskSchema = z.object({
  id: z.string(),
  description: z.string(),
  context: z.string(),
  files: z.array(z.string()),
  timestamp: z.number().optional(),
  cliType: cliTypeSchema.optional(),
});

export const sessionRouter = t.router({
  list: publicProcedure.query(async () => {
    return sessionManager.getAllSessions();
  }),

  active: publicProcedure.query(async () => {
    return sessionManager.getActiveSessions();
  }),

  stats: publicProcedure.query(async () => {
    return sessionManager.getSessionStats();
  }),

  get: publicProcedure.input(z.object({ id: z.string() })).query(async ({ input }) => {
    const session = sessionManager.getSession(input.id);
    if (!session) throw new Error('Session not found');
    return session;
  }),

  start: publicProcedure.input(z.object({
    task: developmentTaskSchema.optional(),
    tags: z.array(z.string()).optional(),
    template: z.string().optional(),
    workingDirectory: z.string().optional(),
    cliType: cliTypeSchema.optional(),
    env: z.record(z.string()).optional(),
  })).mutation(async ({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return sessionManager.startSession(input.task, {
      tags: input.tags,
      templateName: input.template,
      workingDirectory: input.workingDirectory,
      cliType: input.cliType,
      env: input.env,
    });
  }),

  bulkStart: publicProcedure.input(z.object({
    count: z.number(),
    template: z.string().optional(),
    tags: z.array(z.string()).optional(),
    staggerDelayMs: z.number().optional(),
    cliType: cliTypeSchema.optional(),
  })).mutation(async ({ input }) => {
    const result = await sessionManager.startBulkSessions(input.count, {
      tags: input.tags,
      templateName: input.template,
      cliType: input.cliType,
      staggerDelayMs: input.staggerDelayMs ?? 500,
    });

    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'start', count: result.sessions.length, failed: result.failed.length },
      timestamp: Date.now(),
    });

    return result;
  }),

  bulkStop: publicProcedure.mutation(async () => {
    const result = await sessionManager.stopAllSessions();

    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'stop', stopped: result.stopped, failed: result.failed },
      timestamp: Date.now(),
    });

    return result;
  }),

  bulkResume: publicProcedure.mutation(async () => {
    const result = await sessionManager.resumeAllSessions();

    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'resume', count: result.sessions.length, failed: result.failed.length },
      timestamp: Date.now(),
    });

    return result;
  }),

  persisted: publicProcedure.query(async () => {
    return sessionManager.getPersistedSessions();
  }),

  byTag: publicProcedure.input(z.object({ tag: z.string() })).query(async ({ input }) => {
    return sessionManager.getSessionsByTag(input.tag);
  }),

  byTemplate: publicProcedure.input(z.object({ template: z.string() })).query(async ({ input }) => {
    return sessionManager.getSessionsByTemplate(input.template);
  }),

  stop: publicProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    await sessionManager.stopSession(input.id);
    return sessionManager.getSession(input.id) || null;
  }),

  resume: publicProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    return sessionManager.resumeSession(input.id);
  }),

  delete: publicProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    await sessionManager.deleteSession(input.id);
    return { success: true };
  }),

  sendGuidance: publicProcedure.input(z.object({
    id: z.string(),
    approved: z.boolean().default(true),
    feedback: z.string().optional(),
    suggestedNextSteps: z.array(z.string()).optional(),
  })).mutation(async ({ input }) => {
    await sessionManager.sendGuidance(input.id, {
      approved: input.approved,
      feedback: input.feedback || '',
      suggestedNextSteps: input.suggestedNextSteps || [],
    });
    return sessionManager.getSession(input.id) || null;
  }),

  getLogs: publicProcedure.input(z.object({ id: z.string() })).query(async ({ input }) => {
    const session = sessionManager.getSession(input.id);
    if (!session) throw new Error('Session not found');
    return session.logs;
  }),

  templates: publicProcedure.query(async () => {
    const config = loadConfig();
    return config.templates;
  }),

  startFromTemplate: publicProcedure.input(z.object({ name: z.string() })).mutation(async ({ input }) => {
    const config = loadConfig();
    const template = config.templates.find(t => t.name === input.name);
    if (!template) throw new Error(`Template '${input.name}' not found`);

    return sessionManager.startSession(undefined, {
      templateName: template.name,
      tags: template.tags,
    });
  }),

  byCLI: publicProcedure.input(z.object({ cliType: cliTypeSchema })).query(async ({ input }) => {
    return sessionManager.getSessionsByCLI(input.cliType);
  }),

  updateTags: publicProcedure.input(z.object({
    id: z.string(),
    tags: z.array(z.string()),
  })).mutation(async ({ input }) => {
    sessionManager.updateSessionTags(input.id, input.tags);
    const session = sessionManager.getSession(input.id);
    if (!session) throw new Error('Session not found');

    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: input.id, action: 'tags_updated', tags: input.tags },
      timestamp: Date.now(),
    });

    return session;
  }),

  addTag: publicProcedure.input(z.object({
    id: z.string(),
    tag: z.string(),
  })).mutation(async ({ input }) => {
    sessionManager.addSessionTag(input.id, input.tag);
    const session = sessionManager.getSession(input.id);
    if (!session) throw new Error('Session not found');

    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: input.id, action: 'tag_added', tag: input.tag },
      timestamp: Date.now(),
    });

    return session;
  }),

  removeTag: publicProcedure.input(z.object({
    id: z.string(),
    tag: z.string(),
  })).mutation(async ({ input }) => {
    sessionManager.removeSessionTag(input.id, input.tag);
    const session = sessionManager.getSession(input.id);
    if (!session) throw new Error('Session not found');

    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: input.id, action: 'tag_removed', tag: input.tag },
      timestamp: Date.now(),
    });

    return session;
  }),
});
