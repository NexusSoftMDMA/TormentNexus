// @tormentnexus/adk — runtime-safe stubs with proper type/exports
// Provides interfaces and const enums that @tormentnexus/core uses.
// Real implementations live in @tormentnexus/core runtime.

// ── A2A Protocol Types ──────────────────────────────────────────────────────

export interface A2AMessage {
  id: string;
  sender: string;
  recipient: string;
  content: string;
  timestamp: number;
  type: string;
  payload?: any;
}

export const A2AMessageType = {
  HEARTBEAT: 'heartbeat',
  TASK_REQUEST: 'task_request',
  TASK_RESPONSE: 'task_response',
  TASK_NEGOTIATION: 'task_negotiation',
  CAPABILITY_REPORT: 'capability_report',
  STATE_UPDATE: 'state_update',
} as const;

export type A2AMessageType = typeof A2AMessageType[keyof typeof A2AMessageType];

export interface A2ATask {
  id: string;
  type: string;
  payload: any;
  status: string;
  description?: string;
  priority?: string;
}

export interface IA2AClient {
  sendMessage(message: A2AMessage): Promise<void>;
  onMessage(callback: (message: A2AMessage) => void): void;
  delegateTask(task: A2ATask, recipient: string): Promise<A2ATask>;
}

// ── MCP Server Interface ────────────────────────────────────────────────────

export interface IMCPServer {
  name: string;
  version: string;
  modelSelector?: any;
  permissionManager?: any;
  directorConfig?: any;
  getStatus(): any;
  getTools(): any[];
  handleRequest(request: any): Promise<any>;
  executeTool(toolName: string, args: any): Promise<any>;
  start(): Promise<void>;
  stop(): Promise<void>;
}
