import { getLLMService } from '../../../lib/trpc-core.js';
import { wsManager } from './ws-manager.js';

export type RotationRole = 'planner' | 'implementer' | 'tester';
export type RotationMessageKind = 'system' | 'discussion' | 'handoff' | 'artifact';
export type RotationRoomStatus = 'draft' | 'active' | 'paused' | 'completed';
export type RotationRoomMode = 'plan' | 'execute';
export type RotationSupervisorEvaluationMode = 'manual' | 'after_turn' | 'after_cycle';
export type RotationSupervisorTrigger = 'manual' | 'turn_advanced' | 'cycle_completed';
export type RotationSupervisorVerdict = 'continue' | 'complete';
export type RotationSupervisorCompletionAction = 'pause' | 'complete';

const DEFAULT_ROLE_ORDER: RotationRole[] = ['planner', 'implementer', 'tester'];
const DEFAULT_SUPERVISOR_TRANSCRIPT_MESSAGES = 18;

export interface RotationParticipant {
  id: string;
  name: string;
  provider: string;
  model: string;
}

export interface RotationTurn {
  id: string;
  cycleNumber: number;
  turnIndex: number;
  role: RotationRole;
  participantId: string;
  startedAt: number;
  completedAt?: number;
  summary?: string;
}

export interface RotationMessage {
  id: string;
  roomId: string;
  participantId: string;
  participantName: string;
  role: RotationRole | 'system';
  kind: RotationMessageKind;
  content: string;
  cycleNumber: number;
  turnIndex: number;
  timestamp: number;
}

export interface RotationAgreement {
  participantId: string;
  agrees: boolean;
  note?: string;
  updatedAt: number;
}

export interface RotationSupervisorConfig {
  id: string;
  name: string;
  provider: string;
  model: string;
  completionCriteria: string;
  evaluationMode: RotationSupervisorEvaluationMode;
  completionAction: RotationSupervisorCompletionAction;
  enabled: boolean;
  maxTranscriptMessages: number;
  prompt?: string;
  temperature?: number;
}

export interface RotationSupervisorEvaluation {
  id: string;
  trigger: RotationSupervisorTrigger;
  verdict: RotationSupervisorVerdict;
  summary: string;
  confidence?: number;
  satisfiedCriteria: string[];
  remainingCriteria: string[];
  recommendedAction?: string;
  raw: string;
  evaluatedAt: number;
}

export interface RotationSupervisorState {
  status: 'inactive' | 'active' | 'satisfied';
  config?: RotationSupervisorConfig;
  evaluations: RotationSupervisorEvaluation[];
  lastEvaluationAt?: number;
}

export interface RotationRoom {
  id: string;
  title: string;
  objective: string;
  sharedContext: string;
  mode: RotationRoomMode;
  status: RotationRoomStatus;
  participants: RotationParticipant[];
  roleOrder: RotationRole[];
  participantOrder: string[];
  currentCycleNumber: number;
  currentTurnIndex: number;
  roleAssignments: Record<RotationRole, string>;
  agreements: Record<string, RotationAgreement>;
  transcript: RotationMessage[];
  turns: RotationTurn[];
  supervisor: RotationSupervisorState;
  lastCheckpoint?: string;
  consensusReachedAt?: number;
  createdAt: number;
  updatedAt: number;
}

type RotationSupervisorInput = {
  name?: string;
  provider: string;
  model: string;
  completionCriteria: string;
  evaluationMode?: RotationSupervisorEvaluationMode;
  completionAction?: RotationSupervisorCompletionAction;
  enabled?: boolean;
  maxTranscriptMessages?: number;
  prompt?: string;
  temperature?: number;
};

type ParsedSupervisorEvaluation = {
  verdict: RotationSupervisorVerdict;
  summary: string;
  confidence?: number;
  satisfiedCriteria: string[];
  remainingCriteria: string[];
  recommendedAction?: string;
};

type RotationSupervisorRunner = (input: {
  room: RotationRoom;
  config: RotationSupervisorConfig;
  trigger: RotationSupervisorTrigger;
  prompt: string;
}) => Promise<string>;

function rotateLeft<T>(items: T[], count: number): T[] {
  if (items.length === 0) {
    return [];
  }

  const shift = ((count % items.length) + items.length) % items.length;
  return items.slice(shift).concat(items.slice(0, shift));
}

function buildRoleAssignments(
  participantOrder: string[],
  roleOrder: RotationRole[],
  cycleNumber: number,
): Record<RotationRole, string> {
  const rotated = rotateLeft(participantOrder, cycleNumber % participantOrder.length);
  return roleOrder.reduce((assignments, role, index) => {
    assignments[role] = rotated[index % rotated.length];
    return assignments;
  }, {
    planner: rotated[0],
    implementer: rotated[0],
    tester: rotated[0],
  } as Record<RotationRole, string>);
}

function requireParticipant(room: RotationRoom, participantId: string): RotationParticipant {
  const participant = room.participants.find((entry) => entry.id === participantId);
  if (!participant) {
    throw new Error(`Participant '${participantId}' not found in room '${room.id}'`);
  }
  return participant;
}

class RotationRoomService {
  private readonly rooms = new Map<string, RotationRoom>();
  private supervisorRunner: RotationSupervisorRunner = async ({ config, prompt }) => {
    const llm = getLLMService();
    if (typeof llm.generateText !== 'function') {
      throw new Error('Rotation supervisor evaluation requires llmService.generateText()');
    }

    const response = await llm.generateText(
      config.provider,
      config.model,
      'You are a strict autonomous supervisor for a multi-model coding room. Return JSON only.',
      prompt,
      {
        temperature: config.temperature ?? 0,
        taskComplexity: 'high',
      },
    );

    const content = typeof response?.content === 'string'
      ? response.content
      : typeof response?.text === 'string'
        ? response.text
        : '';
    if (!content.trim()) {
      throw new Error(`Rotation supervisor '${config.name}' returned an empty evaluation`);
    }
    return content;
  };

  private makeId(prefix: string): string {
    return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
  }

  createRoom(input: {
    title: string;
    objective: string;
    sharedContext?: string;
    mode?: RotationRoomMode;
    participants: RotationParticipant[];
    roleOrder?: RotationRole[];
    supervisor?: RotationSupervisorInput;
  }): RotationRoom {
    if (input.participants.length < 2) {
      throw new Error('Rotation rooms need at least 2 participants');
    }

    const roleOrder = input.roleOrder && input.roleOrder.length > 0
      ? [...input.roleOrder]
      : [...DEFAULT_ROLE_ORDER];
    const uniqueParticipantIds = new Set(input.participants.map((participant) => participant.id));
    if (uniqueParticipantIds.size !== input.participants.length) {
      throw new Error('Participant IDs must be unique');
    }

    const id = this.makeId('rotation');
    const now = Date.now();
    const participantOrder = input.participants.map((participant) => participant.id);
    const roleAssignments = buildRoleAssignments(participantOrder, roleOrder, 0);
    const firstRole = roleOrder[0];
    const agreements = participantOrder.reduce((accumulator, participantId) => {
      accumulator[participantId] = {
        participantId,
        agrees: false,
        updatedAt: now,
      };
      return accumulator;
    }, {} as Record<string, RotationAgreement>);
    const firstTurn: RotationTurn = {
      id: this.makeId('turn'),
      cycleNumber: 0,
      turnIndex: 0,
      role: firstRole,
      participantId: roleAssignments[firstRole],
      startedAt: now,
    };

    const room: RotationRoom = {
      id,
      title: input.title,
      objective: input.objective,
      sharedContext: input.sharedContext?.trim() ?? '',
      mode: input.mode ?? 'execute',
      status: 'active',
      participants: input.participants.map((participant) => ({ ...participant })),
      roleOrder,
      participantOrder,
      currentCycleNumber: 0,
      currentTurnIndex: 0,
      roleAssignments,
      agreements,
      transcript: [],
      turns: [firstTurn],
      supervisor: this.createSupervisorState(input.supervisor),
      createdAt: now,
      updatedAt: now,
    };

    this.rooms.set(id, room);
    this.addSystemMessage(room, `Rotation room created. Cycle 1 assignments: ${this.describeAssignments(room)}`);
    if (room.supervisor.config?.enabled) {
      this.addSystemMessage(
        room,
        `Supervisor configured: ${room.supervisor.config.name} watching for completion criteria "${room.supervisor.config.completionCriteria}"`,
      );
    }
    this.broadcastRoom(room, 'created');
    return room;
  }

  listRooms(): RotationRoom[] {
    return Array.from(this.rooms.values()).sort((left, right) => right.updatedAt - left.updatedAt);
  }

  getRoom(roomId: string): RotationRoom {
    const room = this.rooms.get(roomId);
    if (!room) {
      throw new Error(`Rotation room '${roomId}' not found`);
    }
    return room;
  }

  postMessage(input: {
    roomId: string;
    participantId: string;
    content: string;
    kind?: RotationMessageKind;
  }): RotationMessage {
    const room = this.getRoom(input.roomId);
    const participant = requireParticipant(room, input.participantId);
    const currentTurn = room.turns[room.turns.length - 1];
    const message: RotationMessage = {
      id: this.makeId('msg'),
      roomId: room.id,
      participantId: participant.id,
      participantName: participant.name,
      role: this.findAssignedRole(room, participant.id) ?? 'system',
      kind: input.kind ?? 'discussion',
      content: input.content,
      cycleNumber: currentTurn.cycleNumber,
      turnIndex: currentTurn.turnIndex,
      timestamp: Date.now(),
    };

    room.transcript.push(message);
    room.updatedAt = message.timestamp;
    this.broadcastRoom(room, 'message_added', message);
    return message;
  }

  setAgreement(input: {
    roomId: string;
    participantId: string;
    agrees: boolean;
    note?: string;
  }): RotationRoom {
    const room = this.getRoom(input.roomId);
    const participant = requireParticipant(room, input.participantId);
    room.agreements[input.participantId] = {
      participantId: input.participantId,
      agrees: input.agrees,
      note: input.note?.trim() || undefined,
      updatedAt: Date.now(),
    };
    room.updatedAt = Date.now();
    room.consensusReachedAt = this.hasConsensus(room) ? Date.now() : undefined;
    this.addSystemMessage(
      room,
      `${participant.name} marked plan ${input.agrees ? 'agreed' : 'not agreed'}${input.note?.trim() ? `: ${input.note.trim()}` : ''}`,
    );
    this.broadcastRoom(room, 'agreement_updated');
    return room;
  }

  async advanceTurn(input: {
    roomId: string;
    summary?: string;
    checkpoint?: string;
  }): Promise<RotationRoom> {
    const room = this.getRoom(input.roomId);
    if (room.status !== 'active') {
      throw new Error(`Cannot advance room in '${room.status}' state`);
    }

    const currentTurn = room.turns[room.turns.length - 1];
    if (!currentTurn.completedAt) {
      currentTurn.completedAt = Date.now();
      currentTurn.summary = input.summary?.trim() || undefined;
    }

    if (input.summary?.trim()) {
      this.addSystemMessage(room, `Turn complete: ${currentTurn.role} summary — ${input.summary.trim()}`);
    }

    if (input.checkpoint?.trim()) {
      room.lastCheckpoint = input.checkpoint.trim();
      this.addSystemMessage(room, `Checkpoint updated: ${room.lastCheckpoint}`);
    }

    let supervisorTrigger: RotationSupervisorTrigger = 'turn_advanced';
    const nextTurnIndex = currentTurn.turnIndex + 1;
    if (nextTurnIndex < room.roleOrder.length) {
      const nextRole = room.roleOrder[nextTurnIndex];
      room.currentTurnIndex = nextTurnIndex;
      room.turns.push({
        id: this.makeId('turn'),
        cycleNumber: room.currentCycleNumber,
        turnIndex: nextTurnIndex,
        role: nextRole,
        participantId: room.roleAssignments[nextRole],
        startedAt: Date.now(),
      });
    } else {
      supervisorTrigger = 'cycle_completed';
      if (room.mode === 'plan' && this.hasConsensus(room)) {
        room.status = 'paused';
        room.consensusReachedAt = Date.now();
        this.addSystemMessage(room, 'Plan-mode consensus reached. Debate is paused until execution resumes.');
      } else {
        room.currentCycleNumber += 1;
        room.currentTurnIndex = 0;
        room.roleAssignments = buildRoleAssignments(room.participantOrder, room.roleOrder, room.currentCycleNumber);
        room.turns.push({
          id: this.makeId('turn'),
          cycleNumber: room.currentCycleNumber,
          turnIndex: 0,
          role: room.roleOrder[0],
          participantId: room.roleAssignments[room.roleOrder[0]],
          startedAt: Date.now(),
        });
        if (room.mode === 'plan') {
          this.addSystemMessage(
            room,
            `Plan-mode consensus not reached. Continue debating. Waiting on: ${this.describePendingAgreements(room)}`,
          );
        }
        this.addSystemMessage(room, `Cycle ${room.currentCycleNumber + 1} assignments: ${this.describeAssignments(room)}`);
      }
    }

    room.updatedAt = Date.now();
    if (this.shouldAutoEvaluateSupervisor(room, supervisorTrigger)) {
      await this.runSupervisorCheckInternal(room, supervisorTrigger);
    }
    this.broadcastRoom(room, 'turn_advanced');
    return room;
  }

  pauseRoom(roomId: string): RotationRoom {
    const room = this.getRoom(roomId);
    room.status = 'paused';
    room.updatedAt = Date.now();
    this.addSystemMessage(room, 'Rotation room paused.');
    this.broadcastRoom(room, 'paused');
    return room;
  }

  resumeRoom(roomId: string): RotationRoom {
    const room = this.getRoom(roomId);
    room.status = 'active';
    room.updatedAt = Date.now();
    this.addSystemMessage(room, 'Rotation room resumed.');
    this.broadcastRoom(room, 'resumed');
    return room;
  }

  addParticipant(input: {
    roomId: string;
    participant: RotationParticipant;
  }): RotationRoom {
    const room = this.getRoom(input.roomId);
    if (room.status === 'completed') {
      throw new Error(`Cannot add participants to completed room '${room.id}'`);
    }
    if (room.participants.some((participant) => participant.id === input.participant.id)) {
      throw new Error(`Participant '${input.participant.id}' already exists in room '${room.id}'`);
    }

    room.participants.push({ ...input.participant });
    room.participantOrder.push(input.participant.id);
    room.agreements[input.participant.id] = {
      participantId: input.participant.id,
      agrees: false,
      updatedAt: Date.now(),
    };
    room.consensusReachedAt = undefined;
    if (room.mode === 'plan' && room.status === 'paused') {
      room.status = 'active';
      this.addSystemMessage(room, 'Plan consensus reopened because a new participant joined the room.');
    }
    room.updatedAt = Date.now();
    this.addSystemMessage(
      room,
      `Participant added: ${input.participant.name} (${input.participant.model}) will join the rotation on the next cycle.`,
    );
    this.broadcastRoom(room, 'participant_added');
    return room;
  }

  configureSupervisor(input: {
    roomId: string;
    supervisor: RotationSupervisorInput;
  }): RotationRoom {
    const room = this.getRoom(input.roomId);
    const existing = room.supervisor.config;
    const config = this.normalizeSupervisorConfig(input.supervisor, existing?.id);
    room.supervisor = {
      ...room.supervisor,
      status: (input.supervisor.enabled ?? true) ? 'active' : 'inactive',
      config,
    };
    room.updatedAt = Date.now();
    this.addSystemMessage(
      room,
      config.enabled
        ? `Supervisor configured: ${config.name} (${config.provider}/${config.model})`
        : `Supervisor disabled: ${config.name}`,
    );
    this.broadcastRoom(room, 'supervisor_configured');
    return room;
  }

  async runSupervisorCheck(input: {
    roomId: string;
    trigger?: RotationSupervisorTrigger;
  }): Promise<RotationRoom> {
    const room = this.getRoom(input.roomId);
    const trigger = input.trigger ?? 'manual';
    await this.runSupervisorCheckInternal(room, trigger);
    this.broadcastRoom(room, 'supervisor_evaluated');
    return room;
  }

  startExecution(roomId: string, checkpoint?: string): RotationRoom {
    const room = this.getRoom(roomId);
    room.mode = 'execute';
    room.status = 'active';
    room.roleOrder = ['implementer'];
    room.currentCycleNumber += 1;
    room.currentTurnIndex = 0;
    room.roleAssignments = buildRoleAssignments(room.participantOrder, room.roleOrder, room.currentCycleNumber);
    room.turns.push({
      id: this.makeId('turn'),
      cycleNumber: room.currentCycleNumber,
      turnIndex: 0,
      role: 'implementer',
      participantId: room.roleAssignments.implementer,
      startedAt: Date.now(),
    });
    if (checkpoint?.trim()) {
      room.lastCheckpoint = checkpoint.trim();
    }
    room.updatedAt = Date.now();
    this.addSystemMessage(room, 'Execution mode started. Implementation now rotates one model per turn.');
    this.broadcastRoom(room, 'execution_started');
    return room;
  }

  completeRoom(roomId: string, summary?: string): RotationRoom {
    const room = this.getRoom(roomId);
    room.status = 'completed';
    room.updatedAt = Date.now();
    if (summary?.trim()) {
      room.lastCheckpoint = summary.trim();
    }
    this.addSystemMessage(room, summary?.trim() ? `Rotation room completed: ${summary.trim()}` : 'Rotation room completed.');
    this.broadcastRoom(room, 'completed');
    return room;
  }

  updateSharedContext(roomId: string, sharedContext: string): RotationRoom {
    const room = this.getRoom(roomId);
    room.sharedContext = sharedContext.trim();
    room.updatedAt = Date.now();
    this.addSystemMessage(room, 'Shared context updated.');
    this.broadcastRoom(room, 'context_updated');
    return room;
  }

  setSupervisorRunner(runner: RotationSupervisorRunner | null): void {
    this.supervisorRunner = runner ?? this.supervisorRunner;
  }

  reset(): void {
    this.rooms.clear();
    this.supervisorRunner = async ({ config, prompt }) => {
      const llm = getLLMService();
      if (typeof llm.generateText !== 'function') {
        throw new Error('Rotation supervisor evaluation requires llmService.generateText()');
      }

      const response = await llm.generateText(
        config.provider,
        config.model,
        'You are a strict autonomous supervisor for a multi-model coding room. Return JSON only.',
        prompt,
        {
          temperature: config.temperature ?? 0,
          taskComplexity: 'high',
        },
      );

      const content = typeof response?.content === 'string'
        ? response.content
        : typeof response?.text === 'string'
          ? response.text
          : '';
      if (!content.trim()) {
        throw new Error(`Rotation supervisor '${config.name}' returned an empty evaluation`);
      }
      return content;
    };
  }

  private addSystemMessage(room: RotationRoom, content: string): void {
    const currentTurn = room.turns[room.turns.length - 1];
    room.transcript.push({
      id: this.makeId('msg'),
      roomId: room.id,
      participantId: 'system',
      participantName: 'System',
      role: 'system',
      kind: 'system',
      content,
      cycleNumber: currentTurn?.cycleNumber ?? room.currentCycleNumber,
      turnIndex: currentTurn?.turnIndex ?? room.currentTurnIndex,
      timestamp: Date.now(),
    });
  }

  private broadcastRoom(room: RotationRoom, event: string, message?: RotationMessage): void {
    wsManager.broadcast({
      type: 'rotation_room',
      payload: {
        event,
        room,
        message,
      },
      timestamp: Date.now(),
    });
  }

  private describeAssignments(room: RotationRoom): string {
    return room.roleOrder
      .map((role) => {
        const participantId = room.roleAssignments[role];
        const participant = requireParticipant(room, participantId);
        return `${role}: ${participant.name} (${participant.model})`;
      })
      .join(' | ');
  }

  private hasConsensus(room: RotationRoom): boolean {
    return room.participantOrder.every((participantId) => room.agreements[participantId]?.agrees === true);
  }

  private describePendingAgreements(room: RotationRoom): string {
    const pending = room.participantOrder
      .filter((participantId) => room.agreements[participantId]?.agrees !== true)
      .map((participantId) => requireParticipant(room, participantId).name);
    return pending.length > 0 ? pending.join(', ') : 'none';
  }

  private findAssignedRole(room: RotationRoom, participantId: string): RotationRole | undefined {
    return room.roleOrder.find((role) => room.roleAssignments[role] === participantId);
  }

  private createSupervisorState(input?: RotationSupervisorInput): RotationSupervisorState {
    return {
      status: input && (input.enabled ?? true) ? 'active' : 'inactive',
      config: input ? this.normalizeSupervisorConfig(input) : undefined,
      evaluations: [],
    };
  }

  private normalizeSupervisorConfig(input: RotationSupervisorInput, existingId?: string): RotationSupervisorConfig {
    return {
      id: existingId ?? this.makeId('supervisor'),
      name: input.name?.trim() || `${input.provider}/${input.model}`,
      provider: input.provider,
      model: input.model,
      completionCriteria: input.completionCriteria.trim(),
      evaluationMode: input.evaluationMode ?? 'manual',
      completionAction: input.completionAction ?? 'complete',
      enabled: input.enabled ?? true,
      maxTranscriptMessages: input.maxTranscriptMessages ?? DEFAULT_SUPERVISOR_TRANSCRIPT_MESSAGES,
      prompt: input.prompt?.trim() || undefined,
      temperature: input.temperature,
    };
  }

  private shouldAutoEvaluateSupervisor(room: RotationRoom, trigger: RotationSupervisorTrigger): boolean {
    const config = room.supervisor.config;
    if (!config?.enabled || room.status === 'completed') {
      return false;
    }
    if (config.evaluationMode === 'manual') {
      return false;
    }
    if (config.evaluationMode === 'after_turn') {
      return trigger === 'turn_advanced' || trigger === 'cycle_completed';
    }
    return trigger === 'cycle_completed';
  }

  private async runSupervisorCheckInternal(room: RotationRoom, trigger: RotationSupervisorTrigger): Promise<void> {
    const config = room.supervisor.config;
    if (!config?.enabled) {
      throw new Error(`Room '${room.id}' does not have an enabled supervisor`);
    }

    const raw = await this.supervisorRunner({
      room,
      config,
      trigger,
      prompt: this.buildSupervisorPrompt(room, config, trigger),
    });
    const parsed = this.parseSupervisorEvaluation(raw);
    const evaluatedAt = Date.now();
    const evaluation: RotationSupervisorEvaluation = {
      id: this.makeId('supervisor_eval'),
      trigger,
      verdict: parsed.verdict,
      summary: parsed.summary,
      confidence: parsed.confidence,
      satisfiedCriteria: parsed.satisfiedCriteria,
      remainingCriteria: parsed.remainingCriteria,
      recommendedAction: parsed.recommendedAction,
      raw,
      evaluatedAt,
    };

    room.supervisor.evaluations.push(evaluation);
    room.supervisor.lastEvaluationAt = evaluatedAt;
    room.supervisor.status = evaluation.verdict === 'complete' ? 'satisfied' : 'active';
    room.updatedAt = evaluatedAt;
    this.addSystemMessage(
      room,
      `Supervisor ${config.name} verdict: ${evaluation.verdict}. ${evaluation.summary}`,
    );

    if (evaluation.verdict === 'complete') {
      if (config.completionAction === 'pause') {
        room.status = 'paused';
        this.addSystemMessage(room, `Room paused by supervisor ${config.name} after completion criteria were satisfied.`);
      } else {
        room.status = 'completed';
        room.lastCheckpoint = evaluation.summary;
        this.addSystemMessage(room, `Rotation room completed by supervisor ${config.name}.`);
      }
    }
  }

  private buildSupervisorPrompt(
    room: RotationRoom,
    config: RotationSupervisorConfig,
    trigger: RotationSupervisorTrigger,
  ): string {
    const transcript = room.transcript
      .slice(-config.maxTranscriptMessages)
      .map((entry) => `[${entry.participantName}][${entry.role}][${entry.kind}] ${entry.content}`)
      .join('\n');
    const recentTurns = room.turns
      .slice(-Math.max(4, room.roleOrder.length))
      .map((turn) => {
        const participant = requireParticipant(room, turn.participantId);
        return `- cycle ${turn.cycleNumber + 1}, turn ${turn.turnIndex + 1}, ${turn.role}, ${participant.name} (${participant.model})${turn.summary ? `, summary: ${turn.summary}` : ''}`;
      })
      .join('\n');

    return [
      'Review this multi-model room and decide whether it has satisfied the completion criteria.',
      'Return JSON only with this shape:',
      '{"verdict":"continue|complete","summary":"string","confidence":0.0,"satisfiedCriteria":["..."],"remainingCriteria":["..."],"recommendedAction":"string"}',
      `Trigger: ${trigger}`,
      `Room title: ${room.title}`,
      `Objective: ${room.objective}`,
      `Mode: ${room.mode}`,
      `Status: ${room.status}`,
      `Shared context: ${room.sharedContext || '(none)'}`,
      `Last checkpoint: ${room.lastCheckpoint || '(none)'}`,
      `Completion criteria: ${config.completionCriteria}`,
      config.prompt ? `Additional supervisor instructions: ${config.prompt}` : '',
      `Participants: ${room.participants.map((participant) => `${participant.name} (${participant.provider}/${participant.model})`).join(', ')}`,
      'Recent turns:',
      recentTurns || '(none)',
      'Recent transcript:',
      transcript || '(none)',
    ].filter(Boolean).join('\n\n');
  }

  private parseSupervisorEvaluation(raw: string): ParsedSupervisorEvaluation {
    const cleaned = raw
      .trim()
      .replace(/^```json\s*/i, '')
      .replace(/^```\s*/i, '')
      .replace(/\s*```$/i, '')
      .trim();

    let parsed: unknown;
    try {
      parsed = JSON.parse(cleaned);
    } catch (error) {
      throw new Error(`Rotation supervisor returned invalid JSON: ${(error as Error).message}`);
    }

    if (!parsed || typeof parsed !== 'object') {
      throw new Error('Rotation supervisor must return a JSON object');
    }

    const record = parsed as Record<string, unknown>;
    if (record.verdict !== 'continue' && record.verdict !== 'complete') {
      throw new Error("Rotation supervisor JSON must include verdict 'continue' or 'complete'");
    }
    if (typeof record.summary !== 'string' || !record.summary.trim()) {
      throw new Error('Rotation supervisor JSON must include a non-empty summary');
    }

    return {
      verdict: record.verdict,
      summary: record.summary.trim(),
      confidence: typeof record.confidence === 'number' ? record.confidence : undefined,
      satisfiedCriteria: this.parseStringArray(record.satisfiedCriteria),
      remainingCriteria: this.parseStringArray(record.remainingCriteria),
      recommendedAction: typeof record.recommendedAction === 'string' && record.recommendedAction.trim()
        ? record.recommendedAction.trim()
        : undefined,
    };
  }

  private parseStringArray(value: unknown): string[] {
    if (!Array.isArray(value)) {
      return [];
    }
    return value.filter((entry): entry is string => typeof entry === 'string' && entry.trim().length > 0);
  }
}

export const rotationRoomService = new RotationRoomService();
