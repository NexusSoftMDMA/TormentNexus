// @tormentnexus/agents — comprehensive stub package
// Provides runtime-safe class stubs with full method surface area
// that @tormentnexus/core actually calls. Real implementations live in @tormentnexus/core runtime.

export * from './orchestration/RiskEvaluator.js';
export * from './orchestration/DebateEngine.js';
export * from './orchestration/ConferenceManager.js';

// ── Type declarations ──────────────────────────────────────────────────────

export interface A2AMessage {
  id: string;
  sender: string;
  recipient: string;
  content: string;
  timestamp: number;
}

export interface A2ATask {
  id: string;
  type: string;
  payload: any;
  status: string;
}

export interface IA2AClient {
  sendMessage(message: A2AMessage): Promise<void>;
  onMessage(callback: (message: A2AMessage) => void): void;
  delegateTask(task: A2ATask, recipient: string): Promise<A2ATask>;
}

export interface CouncilRoleDefinition {
  name: string;
  systemPrompt: string;
  model?: string;
  provider?: string;
}

// ── Const enums (used as values) ────────────────────────────────────────────

export const CouncilRole = {
  CRITIC: 'critic',
  ARCHITECT: 'architect',
  META_ARCHITECT: 'meta_architect',
  PLANNER: 'planner',
  IMPLEMENTER: 'implementer',
  TESTER: 'tester',
  COORDINATOR: 'coordinator',
} as const;

export type CouncilRole = typeof CouncilRole[keyof typeof CouncilRole];

// ── Class stubs (runtime-safe, full surface area) ───────────────────────────

export class A2ALogger {
  constructor(_dir?: string) {}
  logEvent(_event: any): void {}
  getEvents(): any[] { return []; }
  getRecentLogs(_limit?: number): any[] { return []; }
  on(_event: string, _cb: Function): void {}
}

export class Council {
  private _agents: any[] = [];
  private _server: any = null;
  constructor(_modelSelector?: any) {}
  setServer(server: any): void { this._server = server; }
  registerAgent(roleOrAgent: any, agent?: any): void {
    this._agents.push(agent ?? roleOrAgent);
  }
  async runDebate(_topic: string, _roles?: CouncilRoleDefinition[]): Promise<any> {
    return { summary: '[Council stub]', consensus: null };
  }
  async runConsensusSession(_topic: string, _opts?: any): Promise<any> {
    return { summary: '[Council consensus stub]', consensus: null };
  }
  listAgents(): any[] { return this._agents; }
}

export class Director {
  constructor(_server?: any) {}
  async direct(_task: string): Promise<any> { return { result: '[Director stub]' }; }
  async executeTask(_task: string, _opts?: any): Promise<any> { return { result: '[Director executeTask stub]' }; }
  async startAutoDrive(_opts?: any): Promise<void> {}
  async stopAutoDrive(): Promise<void> {}
  getIsActive(): boolean { return false; }
  getStatus(): any { return { active: false, task: null }; }
  async handleUserMessage(_msg: string): Promise<any> { return { response: '[Director handleUserMessage stub]' }; }
  broadcast(_event: string, _data: any): void {}
}

export class PairOrchestrator {
  constructor(_server?: any, _llm?: any) {}
  async run(_config?: any): Promise<any> { return { result: '[PairOrchestrator stub]' }; }
  async runTask(_task: string, _opts?: any): Promise<any> { return { result: '[PairOrchestrator runTask stub]' }; }
  async rotateRoles(): Promise<void> {}
  async setupFrontierSquad(_a?: any, _b?: any): Promise<void> {}
}

export class Supervisor {
  constructor(_server?: any) {}
  async supervise(_task: string): Promise<any> { return { result: '[Supervisor stub]' }; }
}

export class SwarmController {
  constructor(_server?: any, _llm?: any) {}
  private _members: any[] = [];
  async launch(_config?: any): Promise<any> { return { result: '[SwarmController stub]' }; }
  addMember(_member: any): void { this._members.push(_member); }
  async startSession(_goal?: any, _opts?: any): Promise<any> { return { result: '[SwarmController startSession stub]' }; }
  getTranscript(): any[] { return []; }
}

export class ToolPredictor {
  async predict(_query: string): Promise<string[]> { return []; }
}

export const SwarmRole = {
  COORDINATOR: 'coordinator',
  WORKER: 'worker',
  OBSERVER: 'observer',
  PLANNER: 'planner',
  IMPLEMENTER: 'implementer',
  TESTER: 'tester',
  CRITIC: 'critic',
} as const;

export const a2aBroker = {
  routeMessage: async (msg: any) => { if (msg?.type !== 'HEARTBEAT') console.warn('[A2ABroker:Stub] routeMessage (ignored):', msg); },
  on: (_event: string, _cb: Function) => {},
  listAgents: () => [] as any[],
  getHistory: () => [] as any[],
  getNegotiations: () => [] as any[],
  registerAgent: (_id: any, _agent?: any) => {},
};

export const taskQueue = {
  push: async (_task: any) => {},
  pop: () => null as any,
  size: 0,
  listTasks: () => [] as any[],
};
