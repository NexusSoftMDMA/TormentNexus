import { EventEmitter } from 'events';
import type { Message, Vote, TaskType } from './types.js';

// ============ Types ============

export type ParticipantRole = 'owner' | 'admin' | 'voter' | 'observer';

export type ParticipantStatus = 'invited' | 'joined' | 'voted' | 'left' | 'kicked';

export type CollaborativeDebateStatus = 
  | 'draft'
  | 'inviting'
  | 'waiting_for_participants'
  | 'in_progress'
  | 'voting'
  | 'completed'
  | 'cancelled'
  | 'expired';

export interface Participant {
  id: string;
  name: string;
  email?: string;
  role: ParticipantRole;
  status: ParticipantStatus;
  joinedAt?: number;
  votedAt?: number;
  vote?: HumanVote;
  invitedBy: string;
  invitedAt: number;
}

export interface HumanVote {
  participantId: string;
  approved: boolean;
  confidence: number;
  reasoning: string;
  concerns?: string[];
  suggestions?: string[];
  timestamp: number;
}

export interface CollaborativeDebateConfig {
  minParticipants: number;
  maxParticipants: number;
  votingDeadlineMs: number;
  requireAllVotes: boolean;
  allowObservers: boolean;
  allowLateJoins: boolean;
  autoStartOnMinParticipants: boolean;
  consensusThreshold: number;
  allowRevoting: boolean;
  anonymousVoting: boolean;
}

export interface CollaborativeDebate {
  id: string;
  title: string;
  description: string;
  taskType: TaskType;
  status: CollaborativeDebateStatus;
  config: CollaborativeDebateConfig;
  owner: string;
  participants: Map<string, Participant>;
  messages: DebateMessage[];
  aiSupervisorVotes?: Vote[];
  humanVotes: HumanVote[];
  finalDecision?: CollaborativeDecision;
  createdAt: number;
  startedAt?: number;
  votingStartedAt?: number;
  completedAt?: number;
  expiresAt?: number;
  tags?: string[];
  metadata?: Record<string, unknown>;
}

export interface DebateMessage {
  id: string;
  participantId: string;
  participantName: string;
  content: string;
  timestamp: number;
  replyTo?: string;
  reactions?: Map<string, string[]>;
  edited?: boolean;
  editedAt?: number;
}

export interface CollaborativeDecision {
  approved: boolean;
  humanApprovalRate: number;
  aiApprovalRate?: number;
  combinedApprovalRate: number;
  consensus: 'strong' | 'moderate' | 'weak' | 'none';
  summary: string;
  participantCount: number;
  voteCount: number;
  abstainedCount: number;
  decidedAt: number;
}

export interface InviteToken {
  token: string;
  debateId: string;
  role: ParticipantRole;
  email?: string;
  expiresAt: number;
  usedBy?: string;
  usedAt?: number;
}

export interface CollaborativeDebateStats {
  totalDebates: number;
  activeDebates: number;
  completedDebates: number;
  averageParticipants: number;
  averageVotingTime: number;
  consensusRate: number;
  byTaskType: Record<string, { count: number; consensusRate: number }>;
}

// ============ Collaborative Debate Manager Service ============

const DEFAULT_CONFIG: CollaborativeDebateConfig = {
  minParticipants: 2,
  maxParticipants: 10,
  votingDeadlineMs: 24 * 60 * 60 * 1000,
  requireAllVotes: false,
  allowObservers: true,
  allowLateJoins: true,
  autoStartOnMinParticipants: false,
  consensusThreshold: 0.6,
  allowRevoting: true,
  anonymousVoting: false,
};

export class CollaborativeDebateManagerService extends EventEmitter {
  private debates: Map<string, CollaborativeDebate> = new Map();
  private inviteTokens: Map<string, InviteToken> = new Map();
  private participantDebates: Map<string, Set<string>> = new Map();
  private expirationTimers: Map<string, NodeJS.Timeout> = new Map();

  constructor() {
    super();
  }

  // ============ Debate Lifecycle ============

  createDebate(params: {
    title: string;
    description: string;
    taskType: TaskType;
    ownerId: string;
    ownerName: string;
    config?: Partial<CollaborativeDebateConfig>;
    tags?: string[];
    metadata?: Record<string, unknown>;
  }): CollaborativeDebate {
    const id = `collab_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const owner: Participant = {
      id: params.ownerId,
      name: params.ownerName,
      role: 'owner',
      status: 'joined',
      joinedAt: Date.now(),
      invitedBy: params.ownerId,
      invitedAt: Date.now(),
    };

    const debate: CollaborativeDebate = {
      id,
      title: params.title,
      description: params.description,
      taskType: params.taskType,
      status: 'draft',
      config: { ...DEFAULT_CONFIG, ...params.config },
      owner: params.ownerId,
      participants: new Map([[params.ownerId, owner]]),
      messages: [],
      humanVotes: [],
      createdAt: Date.now(),
      tags: params.tags,
      metadata: params.metadata,
    };

    this.debates.set(id, debate);
    this.trackParticipantDebate(params.ownerId, id);
    this.emit('debate:created', debate);
    return debate;
  }

  getDebate(id: string): CollaborativeDebate | undefined {
    return this.debates.get(id);
  }

  listDebates(filter?: {
    status?: CollaborativeDebateStatus;
    participantId?: string;
    taskType?: TaskType;
    tag?: string;
  }): CollaborativeDebate[] {
    let debates = Array.from(this.debates.values());

    if (filter?.status) {
      debates = debates.filter(d => d.status === filter.status);
    }
    if (filter?.participantId) {
      debates = debates.filter(d => d.participants.has(filter.participantId!));
    }
    if (filter?.taskType) {
      debates = debates.filter(d => d.taskType === filter.taskType);
    }
    if (filter?.tag) {
      debates = debates.filter(d => d.tags?.includes(filter.tag!));
    }

    return debates.sort((a, b) => b.createdAt - a.createdAt);
  }

  startInviting(debateId: string): CollaborativeDebate {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');
    if (debate.status !== 'draft') throw new Error('Debate already started');

    debate.status = 'inviting';
    this.emit('debate:inviting', debate);
    return debate;
  }

  startDebate(debateId: string): CollaborativeDebate {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');
    if (!['draft', 'inviting', 'waiting_for_participants'].includes(debate.status)) {
      throw new Error(`Cannot start debate in ${debate.status} status`);
    }

    const joinedCount = this.getJoinedParticipantCount(debate);
    if (joinedCount < debate.config.minParticipants) {
      throw new Error(`Need at least ${debate.config.minParticipants} participants, have ${joinedCount}`);
    }

    debate.status = 'in_progress';
    debate.startedAt = Date.now();
    this.emit('debate:started', debate);
    return debate;
  }

  startVoting(debateId: string): CollaborativeDebate {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');
    if (debate.status !== 'in_progress') throw new Error('Debate not in progress');

    debate.status = 'voting';
    debate.votingStartedAt = Date.now();
    debate.expiresAt = Date.now() + debate.config.votingDeadlineMs;

    this.setExpirationTimer(debateId, debate.config.votingDeadlineMs);
    this.emit('debate:voting_started', debate);
    return debate;
  }

  private setExpirationTimer(debateId: string, delayMs: number): void {
    const existing = this.expirationTimers.get(debateId);
    if (existing) clearTimeout(existing);

    const timer = setTimeout(() => {
      this.handleVotingExpired(debateId);
    }, delayMs);

    this.expirationTimers.set(debateId, timer);
  }

  private handleVotingExpired(debateId: string): void {
    const debate = this.debates.get(debateId);
    if (!debate || debate.status !== 'voting') return;

    if (debate.humanVotes.length > 0) {
      this.finalizeDebate(debateId);
    } else {
      debate.status = 'expired';
      this.emit('debate:expired', debate);
    }
  }

  finalizeDebate(debateId: string): CollaborativeDebate {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');
    if (!['voting', 'in_progress'].includes(debate.status)) {
      throw new Error('Debate not in voting phase');
    }

    const decision = this.calculateDecision(debate);
    debate.finalDecision = decision;
    debate.status = 'completed';
    debate.completedAt = Date.now();

    const timer = this.expirationTimers.get(debateId);
    if (timer) {
      clearTimeout(timer);
      this.expirationTimers.delete(debateId);
    }

    this.emit('debate:completed', { debate, decision });
    return debate;
  }

  cancelDebate(debateId: string, reason?: string): CollaborativeDebate {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');
    if (debate.status === 'completed') throw new Error('Cannot cancel completed debate');

    debate.status = 'cancelled';
    debate.metadata = { ...debate.metadata, cancelReason: reason };

    const timer = this.expirationTimers.get(debateId);
    if (timer) {
      clearTimeout(timer);
      this.expirationTimers.delete(debateId);
    }

    this.emit('debate:cancelled', { debate, reason });
    return debate;
  }

  // ============ Participant Management ============

  createInvite(params: {
    debateId: string;
    inviterId: string;
    role: ParticipantRole;
    email?: string;
    expiresInMs?: number;
  }): InviteToken {
    const debate = this.debates.get(params.debateId);
    if (!debate) throw new Error('Debate not found');

    const inviter = debate.participants.get(params.inviterId);
    if (!inviter || !['owner', 'admin'].includes(inviter.role)) {
      throw new Error('Only owner or admin can create invites');
    }

    if (params.role === 'owner') {
      throw new Error('Cannot create invite for owner role');
    }

    const token: InviteToken = {
      token: `inv_${Date.now()}_${Math.random().toString(36).substr(2, 16)}`,
      debateId: params.debateId,
      role: params.role,
      email: params.email,
      expiresAt: Date.now() + (params.expiresInMs ?? 7 * 24 * 60 * 60 * 1000),
    };

    this.inviteTokens.set(token.token, token);
    this.emit('invite:created', { token, debate });
    return token;
  }

  joinWithToken(params: {
    token: string;
    participantId: string;
    participantName: string;
    email?: string;
  }): { debate: CollaborativeDebate; participant: Participant } {
    const invite = this.inviteTokens.get(params.token);
    if (!invite) throw new Error('Invalid invite token');
    if (invite.expiresAt < Date.now()) throw new Error('Invite token expired');
    if (invite.usedBy) throw new Error('Invite token already used');
    if (invite.email && invite.email !== params.email) {
      throw new Error('Email does not match invite');
    }

    const debate = this.debates.get(invite.debateId);
    if (!debate) throw new Error('Debate not found');

    if (debate.participants.has(params.participantId)) {
      throw new Error('Already a participant');
    }

    if (!debate.config.allowLateJoins && debate.status === 'voting') {
      throw new Error('Late joins not allowed');
    }

    const currentCount = this.getJoinedParticipantCount(debate);
    if (currentCount >= debate.config.maxParticipants) {
      throw new Error('Maximum participants reached');
    }

    const participant: Participant = {
      id: params.participantId,
      name: params.participantName,
      email: params.email,
      role: invite.role,
      status: 'joined',
      joinedAt: Date.now(),
      invitedBy: debate.owner,
      invitedAt: invite.expiresAt - (7 * 24 * 60 * 60 * 1000),
    };

    debate.participants.set(params.participantId, participant);
    invite.usedBy = params.participantId;
    invite.usedAt = Date.now();

    this.trackParticipantDebate(params.participantId, debate.id);

    if (debate.config.autoStartOnMinParticipants && 
        debate.status === 'waiting_for_participants' &&
        this.getJoinedParticipantCount(debate) >= debate.config.minParticipants) {
      debate.status = 'in_progress';
      debate.startedAt = Date.now();
      this.emit('debate:auto_started', debate);
    }

    this.emit('participant:joined', { debate, participant });
    return { debate, participant };
  }

  addParticipantDirectly(params: {
    debateId: string;
    addedBy: string;
    participantId: string;
    participantName: string;
    email?: string;
    role: ParticipantRole;
  }): Participant {
    const debate = this.debates.get(params.debateId);
    if (!debate) throw new Error('Debate not found');

    const adder = debate.participants.get(params.addedBy);
    if (!adder || !['owner', 'admin'].includes(adder.role)) {
      throw new Error('Only owner or admin can add participants');
    }

    if (params.role === 'owner') {
      throw new Error('Cannot add as owner');
    }

    if (debate.participants.has(params.participantId)) {
      throw new Error('Already a participant');
    }

    const currentCount = this.getJoinedParticipantCount(debate);
    if (currentCount >= debate.config.maxParticipants) {
      throw new Error('Maximum participants reached');
    }

    const participant: Participant = {
      id: params.participantId,
      name: params.participantName,
      email: params.email,
      role: params.role,
      status: 'joined',
      joinedAt: Date.now(),
      invitedBy: params.addedBy,
      invitedAt: Date.now(),
    };

    debate.participants.set(params.participantId, participant);
    this.trackParticipantDebate(params.participantId, debate.id);
    this.emit('participant:added', { debate, participant, addedBy: params.addedBy });
    return participant;
  }

  removeParticipant(debateId: string, participantId: string, removedBy: string): void {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    const remover = debate.participants.get(removedBy);
    const participant = debate.participants.get(participantId);
    if (!participant) throw new Error('Participant not found');

    if (participant.role === 'owner') {
      throw new Error('Cannot remove owner');
    }

    const canRemove = remover?.role === 'owner' || 
                      (remover?.role === 'admin' && participant.role !== 'admin') ||
                      participantId === removedBy;

    if (!canRemove) {
      throw new Error('Not authorized to remove this participant');
    }

    participant.status = participantId === removedBy ? 'left' : 'kicked';
    this.untrackParticipantDebate(participantId, debateId);
    this.emit('participant:removed', { debate, participant, removedBy });
  }

  updateParticipantRole(debateId: string, participantId: string, newRole: ParticipantRole, updatedBy: string): Participant {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    const updater = debate.participants.get(updatedBy);
    if (!updater || updater.role !== 'owner') {
      throw new Error('Only owner can change roles');
    }

    const participant = debate.participants.get(participantId);
    if (!participant) throw new Error('Participant not found');

    if (participant.role === 'owner' || newRole === 'owner') {
      throw new Error('Cannot change owner role');
    }

    participant.role = newRole;
    this.emit('participant:role_changed', { debate, participant, newRole, updatedBy });
    return participant;
  }

  private getJoinedParticipantCount(debate: CollaborativeDebate): number {
    return Array.from(debate.participants.values())
      .filter(p => p.status === 'joined' || p.status === 'voted')
      .length;
  }

  private trackParticipantDebate(participantId: string, debateId: string): void {
    const debates = this.participantDebates.get(participantId) ?? new Set();
    debates.add(debateId);
    this.participantDebates.set(participantId, debates);
  }

  private untrackParticipantDebate(participantId: string, debateId: string): void {
    const debates = this.participantDebates.get(participantId);
    if (debates) {
      debates.delete(debateId);
    }
  }

  getParticipantDebates(participantId: string): CollaborativeDebate[] {
    const debateIds = this.participantDebates.get(participantId) ?? new Set();
    return Array.from(debateIds)
      .map(id => this.debates.get(id))
      .filter((d): d is CollaborativeDebate => d !== undefined);
  }

  // ============ Messaging ============

  addMessage(params: {
    debateId: string;
    participantId: string;
    content: string;
    replyTo?: string;
  }): DebateMessage {
    const debate = this.debates.get(params.debateId);
    if (!debate) throw new Error('Debate not found');

    const participant = debate.participants.get(params.participantId);
    if (!participant) throw new Error('Not a participant');
    if (participant.role === 'observer') throw new Error('Observers cannot post messages');
    if (!['in_progress', 'voting'].includes(debate.status)) {
      throw new Error('Debate not accepting messages');
    }

    const message: DebateMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      participantId: params.participantId,
      participantName: participant.name,
      content: params.content,
      timestamp: Date.now(),
      replyTo: params.replyTo,
      reactions: new Map(),
    };

    debate.messages.push(message);
    this.emit('message:added', { debate, message });
    return message;
  }

  editMessage(debateId: string, messageId: string, participantId: string, newContent: string): DebateMessage {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    const message = debate.messages.find(m => m.id === messageId);
    if (!message) throw new Error('Message not found');
    if (message.participantId !== participantId) throw new Error('Can only edit own messages');

    message.content = newContent;
    message.edited = true;
    message.editedAt = Date.now();

    this.emit('message:edited', { debate, message });
    return message;
  }

  addReaction(debateId: string, messageId: string, participantId: string, reaction: string): void {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    const participant = debate.participants.get(participantId);
    if (!participant) throw new Error('Not a participant');

    const message = debate.messages.find(m => m.id === messageId);
    if (!message) throw new Error('Message not found');

    if (!message.reactions) {
      message.reactions = new Map();
    }

    const existing = message.reactions.get(reaction) ?? [];
    if (!existing.includes(participantId)) {
      existing.push(participantId);
      message.reactions.set(reaction, existing);
      this.emit('reaction:added', { debate, message, reaction, participantId });
    }
  }

  getMessages(debateId: string, since?: number): DebateMessage[] {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    if (since) {
      return debate.messages.filter(m => m.timestamp > since);
    }
    return [...debate.messages];
  }

  // ============ Voting ============

  submitVote(params: {
    debateId: string;
    participantId: string;
    approved: boolean;
    confidence: number;
    reasoning: string;
    concerns?: string[];
    suggestions?: string[];
  }): HumanVote {
    const debate = this.debates.get(params.debateId);
    if (!debate) throw new Error('Debate not found');
    if (debate.status !== 'voting') throw new Error('Debate not in voting phase');

    const participant = debate.participants.get(params.participantId);
    if (!participant) throw new Error('Not a participant');
    if (participant.role === 'observer') throw new Error('Observers cannot vote');

    const existingVoteIndex = debate.humanVotes.findIndex(v => v.participantId === params.participantId);
    if (existingVoteIndex >= 0 && !debate.config.allowRevoting) {
      throw new Error('Already voted and revoting not allowed');
    }

    const vote: HumanVote = {
      participantId: params.participantId,
      approved: params.approved,
      confidence: Math.max(0, Math.min(1, params.confidence)),
      reasoning: params.reasoning,
      concerns: params.concerns,
      suggestions: params.suggestions,
      timestamp: Date.now(),
    };

    if (existingVoteIndex >= 0) {
      debate.humanVotes[existingVoteIndex] = vote;
    } else {
      debate.humanVotes.push(vote);
    }

    participant.status = 'voted';
    participant.votedAt = Date.now();
    participant.vote = vote;

    this.emit('vote:submitted', { debate, vote, isRevote: existingVoteIndex >= 0 });

    const eligibleVoters = this.getEligibleVoterCount(debate);
    if (debate.config.requireAllVotes && debate.humanVotes.length >= eligibleVoters) {
      this.finalizeDebate(params.debateId);
    }

    return vote;
  }

  getVotes(debateId: string, requesterId: string): HumanVote[] {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    const requester = debate.participants.get(requesterId);
    if (!requester) throw new Error('Not a participant');

    if (debate.config.anonymousVoting && debate.status === 'voting') {
      return debate.humanVotes.map(v => ({
        ...v,
        participantId: 'anonymous',
        reasoning: requester.role === 'owner' ? v.reasoning : '[hidden during voting]',
      }));
    }

    return [...debate.humanVotes];
  }

  private getEligibleVoterCount(debate: CollaborativeDebate): number {
    return Array.from(debate.participants.values())
      .filter(p => p.role !== 'observer' && ['joined', 'voted'].includes(p.status))
      .length;
  }

  // ============ AI Integration ============

  setAISupervisorVotes(debateId: string, votes: Vote[]): void {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    debate.aiSupervisorVotes = votes;
    this.emit('ai_votes:set', { debate, votes });
  }

  // ============ Decision Calculation ============

  private calculateDecision(debate: CollaborativeDebate): CollaborativeDecision {
    const humanVotes = debate.humanVotes;
    const eligibleVoters = this.getEligibleVoterCount(debate);
    const voteCount = humanVotes.length;
    const abstainedCount = eligibleVoters - voteCount;

    const humanApprovalRate = voteCount > 0
      ? humanVotes.filter(v => v.approved).length / voteCount
      : 0;

    let aiApprovalRate: number | undefined;
    if (debate.aiSupervisorVotes && debate.aiSupervisorVotes.length > 0) {
      aiApprovalRate = debate.aiSupervisorVotes.filter(v => v.approved).length / debate.aiSupervisorVotes.length;
    }

    const combinedApprovalRate = aiApprovalRate !== undefined
      ? (humanApprovalRate + aiApprovalRate) / 2
      : humanApprovalRate;

    const approved = combinedApprovalRate >= debate.config.consensusThreshold;

    let consensus: 'strong' | 'moderate' | 'weak' | 'none';
    if (combinedApprovalRate >= 0.9 || combinedApprovalRate <= 0.1) {
      consensus = 'strong';
    } else if (combinedApprovalRate >= 0.7 || combinedApprovalRate <= 0.3) {
      consensus = 'moderate';
    } else if (combinedApprovalRate >= 0.5 || combinedApprovalRate <= 0.5) {
      consensus = 'weak';
    } else {
      consensus = 'none';
    }

    const summary = this.generateDecisionSummary(debate, approved, humanApprovalRate, aiApprovalRate);

    return {
      approved,
      humanApprovalRate,
      aiApprovalRate,
      combinedApprovalRate,
      consensus,
      summary,
      participantCount: eligibleVoters,
      voteCount,
      abstainedCount,
      decidedAt: Date.now(),
    };
  }

  private generateDecisionSummary(
    debate: CollaborativeDebate,
    approved: boolean,
    humanRate: number,
    aiRate?: number
  ): string {
    const humanPct = Math.round(humanRate * 100);
    const aiPct = aiRate !== undefined ? Math.round(aiRate * 100) : null;

    let summary = `The proposal "${debate.title}" was ${approved ? 'APPROVED' : 'REJECTED'}. `;
    summary += `Human approval: ${humanPct}% (${debate.humanVotes.length} votes). `;
    
    if (aiPct !== null) {
      summary += `AI supervisor approval: ${aiPct}% (${debate.aiSupervisorVotes?.length} votes). `;
    }

    const concerns = debate.humanVotes
      .flatMap(v => v.concerns ?? [])
      .filter((c, i, arr) => arr.indexOf(c) === i);
    
    if (concerns.length > 0) {
      summary += `Key concerns: ${concerns.slice(0, 3).join(', ')}.`;
    }

    return summary;
  }

  // ============ Statistics ============

  getStatistics(): CollaborativeDebateStats {
    const debates = Array.from(this.debates.values());
    const completed = debates.filter(d => d.status === 'completed');
    const active = debates.filter(d => ['in_progress', 'voting', 'waiting_for_participants'].includes(d.status));

    const totalParticipants = completed.reduce((sum, d) => sum + d.participants.size, 0);
    const avgParticipants = completed.length > 0 ? totalParticipants / completed.length : 0;

    const votingTimes = completed
      .filter(d => d.votingStartedAt && d.completedAt)
      .map(d => d.completedAt! - d.votingStartedAt!);
    const avgVotingTime = votingTimes.length > 0
      ? votingTimes.reduce((a, b) => a + b, 0) / votingTimes.length
      : 0;

    const consensusDebates = completed.filter(d => 
      d.finalDecision && ['strong', 'moderate'].includes(d.finalDecision.consensus)
    );
    const consensusRate = completed.length > 0 ? consensusDebates.length / completed.length : 0;

    const byTaskType: Record<string, { count: number; consensusRate: number }> = {};
    for (const debate of completed) {
      const taskType = debate.taskType;
      if (!byTaskType[taskType]) {
        byTaskType[taskType] = { count: 0, consensusRate: 0 };
      }
      byTaskType[taskType].count++;
      if (debate.finalDecision && ['strong', 'moderate'].includes(debate.finalDecision.consensus)) {
        byTaskType[taskType].consensusRate++;
      }
    }
    for (const taskType of Object.keys(byTaskType)) {
      byTaskType[taskType].consensusRate = byTaskType[taskType].consensusRate / byTaskType[taskType].count;
    }

    return {
      totalDebates: debates.length,
      activeDebates: active.length,
      completedDebates: completed.length,
      averageParticipants: avgParticipants,
      averageVotingTime: avgVotingTime,
      consensusRate,
      byTaskType,
    };
  }

  // ============ Export/Import ============

  exportDebate(debateId: string): {
    debate: Omit<CollaborativeDebate, 'participants'> & { participants: [string, Participant][] };
  } {
    const debate = this.debates.get(debateId);
    if (!debate) throw new Error('Debate not found');

    return {
      debate: {
        ...debate,
        participants: Array.from(debate.participants.entries()),
        // @ts-expect-error -- type mismatch from inferred schema types
        messages: debate.messages.map(m => ({
          ...m,
          reactions: m.reactions ? Array.from(m.reactions.entries()) : undefined,
        })) as DebateMessage[],
      },
    };
  }

  importDebate(data: {
    debate: Omit<CollaborativeDebate, 'participants'> & { participants: [string, Participant][] };
  }): CollaborativeDebate {
    const id = `collab_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    const debate: CollaborativeDebate = {
      ...data.debate,
      id,
      status: 'completed',
      participants: new Map(data.debate.participants),
      messages: data.debate.messages.map(m => ({
        ...m,
        reactions: m.reactions ? new Map(m.reactions as unknown as [string, string[]][]) : undefined,
      })) as DebateMessage[],
    };

    this.debates.set(id, debate);
    this.emit('debate:imported', debate);
    return debate;
  }
}

export const collaborativeDebateManager = new CollaborativeDebateManagerService();
