export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface SupervisorConfig {
  name: string;
  provider: 'openai' | 'anthropic' | 'google' | 'xai' | 'moonshot' | 'deepseek' | 'qwen' | 'custom' | 'gemini' | 'grok' | 'kimi' | 'openrouter';
  apiKey?: string;
  model?: string;
  temperature?: number;
  baseURL?: string;
  systemPrompt?: string;
  weight?: number;
}

export type ConsensusMode = 
  | 'simple-majority'      // >50% approval
  | 'supermajority'        // >66% approval  
  | 'unanimous'            // 100% approval
  | 'weighted'             // Weighted by supervisor weight × confidence
  | 'ceo-override'         // Lead supervisor can override (head honcho)
  | 'ceo-veto'             // Lead supervisor can only veto, not force approve
  | 'hybrid-ceo-majority'  // CEO decides ties, majority otherwise
  | 'ranked-choice';       // Supervisors rank options, highest ranked wins

export interface SpecializedCouncilConfig extends CouncilConfig {
  id: string;
  name: string;
  description: string;
  specialties: TaskType[];
}

export interface CouncilConfig {
  supervisors: SupervisorConfig[];
  debateRounds?: number;
  consensusThreshold?: number;
  enabled?: boolean;
  smartPilot?: boolean;
  weightedVoting?: boolean;
  consensusMode?: ConsensusMode;
  leadSupervisor?: string;  // Name of the CEO/head honcho supervisor
  fallbackSupervisors?: string[];  // Ordered fallback chain
  specializedCouncils?: SpecializedCouncilConfig[];
}

export interface SessionTemplate {
  name: string;
  description?: string;
  supervisors: SupervisorConfig[];
  councilConfig?: Partial<CouncilConfig>;
  autoStart?: boolean;
  tags?: string[];
}

export interface PersistedSession {
  id: string;
  status: 'idle' | 'starting' | 'running' | 'paused' | 'stopped' | 'error' | 'completed';
  startedAt: number;
  lastActivity?: number;
  currentTask?: string;
  port: number;
  workingDirectory?: string;
  templateName?: string;
  tags?: string[];
  metadata?: Record<string, unknown>;
}

export interface Session {
  id: string;
  status: 'idle' | 'starting' | 'running' | 'paused' | 'stopped' | 'error' | 'completed';
  startedAt: number;
  lastActivity?: number;
  currentTask?: string;
  logs: LogEntry[];
  port?: number;
  workingDirectory?: string;
  templateName?: string;
  tags?: string[];
}

export interface LogEntry {
  timestamp: number;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  source?: string;
}

export interface DevelopmentTask {
  id: string;
  description: string;
  context: string;
  files: string[];
  timestamp?: number;
  cliType?: CLIType;
}

export interface SubTask {
  id: string;
  title: string;
  description: string;
  dependencies: string[];
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  context?: string;
  preferredCLI?: CLIType;
}

export interface TaskPlan {
  originalTaskId?: string;
  subtasks: SubTask[];
  reasoning: string;
}

export interface CouncilDecision {
  approved: boolean;
  consensus: number;
  weightedConsensus?: number;
  votes: Vote[];
  reasoning: string;
  dissent?: string[];
}

export interface Vote {
  supervisor: string;
  approved: boolean;
  confidence: number;
  weight: number;
  comment: string;
}

export interface Guidance {
  approved: boolean;
  feedback: string;
  suggestedNextSteps: string[];
}

export interface Supervisor {
  name: string;
  provider: string;
  chat(messages: Message[]): Promise<string>;
  isAvailable(): Promise<boolean>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface WebSocketMessage {
  type: 'session_update' | 'council_decision' | 'log' | 'error' | 'bulk_update' | 'supervisor_fallback' | 'rotation_room';
  payload: unknown;
  timestamp: number;
}

export interface BulkSessionRequest {
  count: number;
  template?: string;
  tags?: string[];
  staggerDelayMs?: number;
}

export interface BulkSessionResponse {
  sessions: Session[];
  failed: Array<{ index: number; error: string }>;
}

export interface SessionPersistenceConfig {
  enabled: boolean;
  filePath: string;
  autoSaveIntervalMs: number;
  autoResumeOnStart: boolean;
  maxPersistedSessions: number;
}

export type CLIType =
  | 'tormentnexus'
  | 'opencode'
  | 'antigravity'
  | 'claude'
  | 'aider'
  | 'cursor'
  | 'continue'
  | 'cody'
  | 'copilot'
  | 'custom'
  | 'adrenaline'
  | 'amazon-q'
  | 'amazon-q-developer'
  | 'amp-code'
  | 'auggie'
  | 'azure-openai'
  | 'bito'
  | 'byterover'
  | 'claude-code'
  | 'code-codex'
  | 'codebuff'
  | 'codemachine'
  | 'codex'
  | 'crush'
  | 'dolt'
  | 'factory'
  | 'factory-droid'
  | 'gemini'
  | 'goose'
  | 'grok'
  | 'jules'
  | 'kilo-code'
  | 'kimi'
  | 'llm'
  | 'litellm'
  | 'llamafile'
  | 'manus'
  | 'mistral-vibe'
  | 'ollama'
  | 'open-interpreter'
  | 'pi'
  | 'qwen-code'
  | 'rowboatx'
  | 'rovo'
  | 'shell-pilot'
  | 'smithery'
  | 'superai-cli'
  | 'trae'
  | 'warp';

export interface CLITool {
  type: CLIType;
  name: string;
  command: string;
  args: string[];
  healthEndpoint?: string;
  detectCommand?: string;
  available?: boolean;
  version?: string;
  capabilities?: string[];
  interactive?: boolean;
  promptRegex?: string;
}

export interface LogRotationConfig {
  maxLogsPerSession: number;
  maxLogAgeMs: number;
  pruneIntervalMs: number;
}

export interface HealthCheckConfig {
  enabled: boolean;
  intervalMs: number;
  timeoutMs: number;
  maxFailures: number;
}

export interface CrashRecoveryConfig {
  enabled: boolean;
  maxRestartAttempts: number;
  restartDelayMs: number;
  backoffMultiplier: number;
  maxBackoffMs: number;
}

export type SessionHealthStatus = 'healthy' | 'degraded' | 'unresponsive' | 'crashed';

export interface SessionHealth {
  status: SessionHealthStatus;
  lastCheck: number;
  consecutiveFailures: number;
  restartCount: number;
  lastRestartAt?: number;
  errorMessage?: string;
}

export type TaskType = 
  | 'security-audit'
  | 'ui-design'
  | 'api-design'
  | 'performance'
  | 'refactoring'
  | 'bug-fix'
  | 'testing'
  | 'documentation'
  | 'architecture'
  | 'code-review'
  | 'general';

export interface SupervisorProfile {
  name: string;
  provider: string;
  strengths: TaskType[];
  weaknesses?: TaskType[];
  specializations?: string[];
  preferredForLeadOn?: TaskType[];
}

export interface TeamTemplate {
  name: string;
  description: string;
  taskTypes: TaskType[];
  supervisors: string[];
  leadSupervisor?: string;
  consensusMode?: ConsensusMode;
  minSupervisors?: number;
}

export interface TeamSelectionResult {
  team: string[];
  leadSupervisor?: string;
  consensusMode: ConsensusMode;
  reasoning: string;
  taskType: TaskType;
  confidence: number;
}
