import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rotationRoomService } from '../rotation-room.js';

const { mockBroadcast } = vi.hoisted(() => ({
  mockBroadcast: vi.fn(),
}));

vi.mock('../ws-manager.js', () => ({
  wsManager: {
    broadcast: mockBroadcast,
  },
}));

describe('rotationRoomService', () => {
  beforeEach(() => {
    mockBroadcast.mockClear();
    rotationRoomService.reset();
  });

  it('creates a room with initial rotating assignments', () => {
    const room = rotationRoomService.createRoom({
      title: 'Frontier trio',
      objective: 'Ship a feature together',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    expect(room.status).toBe('active');
    expect(room.roleAssignments.planner).toBe('claude');
    expect(room.roleAssignments.implementer).toBe('gpt');
    expect(room.roleAssignments.tester).toBe('gemini');
    expect(room.turns).toHaveLength(1);
    expect(room.turns[0]?.role).toBe('planner');
    expect(mockBroadcast).toHaveBeenCalled();
  });

  it('advances turns within a cycle and rotates roles on the next cycle', async () => {
    const room = rotationRoomService.createRoom({
      title: 'Rotation test',
      objective: 'Verify role handoff',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan drafted' });
    let updated = rotationRoomService.getRoom(room.id);
    expect(updated.turns.at(-1)?.role).toBe('implementer');
    expect(updated.currentCycleNumber).toBe(0);

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Implementation landed' });
    updated = rotationRoomService.getRoom(room.id);
    expect(updated.turns.at(-1)?.role).toBe('tester');

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Tests passed', checkpoint: 'Ready for merge' });
    updated = rotationRoomService.getRoom(room.id);
    expect(updated.currentCycleNumber).toBe(1);
    expect(updated.turns.at(-1)?.role).toBe('planner');
    expect(updated.roleAssignments.planner).toBe('gpt');
    expect(updated.roleAssignments.implementer).toBe('gemini');
    expect(updated.roleAssignments.tester).toBe('claude');
    expect(updated.lastCheckpoint).toBe('Ready for merge');
  });

  it('stores shared discussion messages against the active room', () => {
    const room = rotationRoomService.createRoom({
      title: 'Chatroom',
      objective: 'Share context',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
      ],
    });

    const message = rotationRoomService.postMessage({
      roomId: room.id,
      participantId: 'claude',
      content: 'I propose we refactor the router first.',
      kind: 'discussion',
    });

    const updated = rotationRoomService.getRoom(room.id);
    expect(message.participantName).toBe('Claude');
    expect(updated.transcript.some((entry) => entry.content.includes('refactor the router'))).toBe(true);
  });

  it('keeps plan rooms debating until everybody agrees', async () => {
    const room = rotationRoomService.createRoom({
      title: 'Plan room',
      objective: 'Reach planning consensus',
      mode: 'plan',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'claude', agrees: true });
    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: false, note: 'Need more detail' });

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan round 1' });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Implementation critique' });
    let updated = await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Testing critique' });

    expect(updated.status).toBe('active');
    expect(updated.currentCycleNumber).toBe(1);

    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: true });
    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'gemini', agrees: true });

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Consensus plan refined' });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Implementation checks aligned' });
    updated = await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Testing checks aligned' });

    expect(updated.status).toBe('paused');
    expect(updated.consensusReachedAt).toBeDefined();
    expect(updated.transcript.some((entry) => entry.content.includes('consensus reached'))).toBe(true);
  });

  it('switches into execution mode and rotates implementer turns between models', async () => {
    const room = rotationRoomService.createRoom({
      title: 'Execution room',
      objective: 'Implement in rounds',
      mode: 'plan',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'claude', agrees: true });
    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: true });
    rotationRoomService.setAgreement({ roomId: room.id, participantId: 'gemini', agrees: true });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan pass 1' });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan pass 2' });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan pass 3' });

    let updated = rotationRoomService.startExecution(room.id, 'Consensus plan locked');
    expect(updated.mode).toBe('execute');
    expect(updated.status).toBe('active');
    expect(updated.turns.at(-1)?.role).toBe('implementer');

    const firstImplementer = updated.turns.at(-1)?.participantId;
    updated = await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Implementation turn 1' });
    const secondImplementer = updated.turns.at(-1)?.participantId;

    expect(secondImplementer).not.toBe(firstImplementer);
    expect(updated.turns.at(-1)?.role).toBe('implementer');
  });

  it('adds a participant and rotates them into later cycles', async () => {
    const room = rotationRoomService.createRoom({
      title: 'Expandable room',
      objective: 'Let more models join',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    const updated = rotationRoomService.addParticipant({
      roomId: room.id,
      participant: { id: 'qwen', name: 'Qwen', provider: 'local', model: 'qwen2.5-coder' },
    });

    expect(updated.participants).toHaveLength(4);
    expect(updated.participantOrder).toContain('qwen');

    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Plan step' });
    await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Implementation step' });
    const nextCycle = await rotationRoomService.advanceTurn({ roomId: room.id, summary: 'Testing step' });

    expect(Object.values(nextCycle.roleAssignments)).toContain('qwen');
  });

  it('runs an autonomous supervisor after turns and completes the room when criteria are satisfied', async () => {
    rotationRoomService.setSupervisorRunner(async () => JSON.stringify({
      verdict: 'complete',
      summary: 'All requested work is implemented and verified.',
      confidence: 0.92,
      satisfiedCriteria: ['Implementation complete', 'Testing complete'],
      remainingCriteria: [],
      recommendedAction: 'Stop autonomous execution.',
    }));

    const room = rotationRoomService.createRoom({
      title: 'Supervised room',
      objective: 'Keep working until qwen says stop',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
      ],
      supervisor: {
        name: 'Local Qwen judge',
        provider: 'local',
        model: 'qwen2.5-coder',
        completionCriteria: 'The room has produced a complete, tested solution.',
        evaluationMode: 'after_turn',
        completionAction: 'complete',
      },
    });

    const updated = await rotationRoomService.advanceTurn({
      roomId: room.id,
      summary: 'Claude finished the initial implementation pass.',
    });

    expect(updated.status).toBe('completed');
    expect(updated.supervisor.evaluations).toHaveLength(1);
    expect(updated.supervisor.evaluations[0]?.verdict).toBe('complete');
    expect(updated.lastCheckpoint).toContain('implemented and verified');
  });
});
