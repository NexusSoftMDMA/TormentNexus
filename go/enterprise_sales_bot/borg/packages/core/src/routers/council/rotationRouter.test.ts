import { beforeEach, describe, expect, it } from 'vitest';
import { rotationRoomService } from '../../orchestrator/council/services/rotation-room.js';
import { rotationRouter } from './rotationRouter.js';

function createCaller() {
  return rotationRouter.createCaller({} as never);
}

describe('rotationRouter', () => {
  beforeEach(() => {
    rotationRoomService.reset();
  });

  it('creates a rotation room and advances role ownership across cycles', async () => {
    const caller = createCaller();

    const room = await caller.create({
      title: 'Autopilot trio',
      objective: 'Rotate planning, implementation, and testing',
      sharedContext: 'The models should share the same task context.',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    expect(room.roleAssignments.planner).toBe('claude');
    expect(room.turns.at(-1)?.role).toBe('planner');

    await caller.postMessage({
      roomId: room.id,
      participantId: 'claude',
      content: 'Plan: add a scheduler service first.',
    });

    await caller.advanceTurn({ roomId: room.id, summary: 'Planning complete' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Implementation complete' });
    const advanced = await caller.advanceTurn({ roomId: room.id, summary: 'Testing complete' });

    expect(advanced.currentCycleNumber).toBe(1);
    expect(advanced.roleAssignments.planner).toBe('gpt');
    expect(advanced.turns.at(-1)?.role).toBe('planner');

    const fetched = await caller.get({ roomId: room.id });
    expect(fetched.transcript.some((entry) => entry.content.includes('scheduler service'))).toBe(true);
  });

  it('keeps plan-mode rooms debating until all models agree', async () => {
    const caller = createCaller();

    const room = await caller.create({
      title: 'Plan consensus room',
      objective: 'Debate until agreement',
      mode: 'plan',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    await caller.setAgreement({ roomId: room.id, participantId: 'claude', agrees: true });
    await caller.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: false, note: 'Need a tighter plan' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Round 1 planning' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Round 1 implementation debate' });
    let updated = await caller.advanceTurn({ roomId: room.id, summary: 'Round 1 testing debate' });

    expect(updated.status).toBe('active');
    expect(updated.currentCycleNumber).toBe(1);

    await caller.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: true });
    await caller.setAgreement({ roomId: room.id, participantId: 'gemini', agrees: true });
    await caller.advanceTurn({ roomId: room.id, summary: 'Round 2 planning' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Round 2 implementation debate' });
    updated = await caller.advanceTurn({ roomId: room.id, summary: 'Round 2 testing debate' });

    expect(updated.status).toBe('paused');
    expect(updated.consensusReachedAt).toBeDefined();
  });

  it('starts execution mode and rotates implementer turns between models', async () => {
    const caller = createCaller();

    const room = await caller.create({
      title: 'Implementation room',
      objective: 'Round-robin implementation',
      mode: 'plan',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    await caller.setAgreement({ roomId: room.id, participantId: 'claude', agrees: true });
    await caller.setAgreement({ roomId: room.id, participantId: 'gpt', agrees: true });
    await caller.setAgreement({ roomId: room.id, participantId: 'gemini', agrees: true });
    await caller.advanceTurn({ roomId: room.id, summary: 'Plan round 1' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Plan round 2' });
    await caller.advanceTurn({ roomId: room.id, summary: 'Plan round 3' });

    let updated = await caller.startExecution({ roomId: room.id, checkpoint: 'Ready to code' });
    expect(updated.mode).toBe('execute');
    expect(updated.turns.at(-1)?.role).toBe('implementer');

    const firstImplementer = updated.turns.at(-1)?.participantId;
    updated = await caller.advanceTurn({ roomId: room.id, summary: 'Implementation pass 1' });
    expect(updated.turns.at(-1)?.role).toBe('implementer');
    expect(updated.turns.at(-1)?.participantId).not.toBe(firstImplementer);
  });

  it('adds participants and runs a supervisor check through the router', async () => {
    const caller = createCaller();
    rotationRoomService.setSupervisorRunner(async () => JSON.stringify({
      verdict: 'continue',
      summary: 'Keep iterating; testing evidence is still incomplete.',
      confidence: 0.71,
      satisfiedCriteria: ['Implementation underway'],
      remainingCriteria: ['Testing evidence'],
      recommendedAction: 'Let the next implementer continue.',
    }));

    let room = await caller.create({
      title: 'Supervisor room',
      objective: 'Let multiple models work until judged complete',
      participants: [
        { id: 'claude', name: 'Claude', provider: 'anthropic', model: 'claude-sonnet-4.6' },
        { id: 'gpt', name: 'GPT', provider: 'openai', model: 'gpt-5.4' },
        { id: 'gemini', name: 'Gemini', provider: 'google', model: 'gemini-2.5-pro' },
      ],
    });

    room = await caller.addParticipant({
      roomId: room.id,
      participant: { id: 'qwen', name: 'Qwen', provider: 'local', model: 'qwen2.5-coder' },
    });
    expect(room.participants.map((participant) => participant.id)).toContain('qwen');

    room = await caller.configureSupervisor({
      roomId: room.id,
      supervisor: {
        name: 'Local Qwen supervisor',
        provider: 'local',
        model: 'qwen2.5-coder',
        completionCriteria: 'Stop only once implementation and tests both look complete.',
        evaluationMode: 'manual',
        completionAction: 'pause',
      },
    });
    expect(room.supervisor.config?.model).toBe('qwen2.5-coder');

    room = await caller.runSupervisorCheck({ roomId: room.id });
    expect(room.supervisor.evaluations).toHaveLength(1);
    expect(room.supervisor.evaluations[0]?.verdict).toBe('continue');
  });
});
