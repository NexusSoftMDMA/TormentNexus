import { Hono } from 'hono';
import { collaborativeDebateManager } from '../services/collaborative-debate-manager.js';
import type { CollaborativeDebateStatus, ParticipantRole } from '../services/collaborative-debate-manager.js';
import type { TaskType } from './types.js';

const app = new Hono();

// ============ Debate CRUD ============

app.post('/', async (c) => {
  const body = await c.req.json();
  const debate = collaborativeDebateManager.createDebate({
    title: body.title,
    description: body.description,
    taskType: body.taskType ?? 'general',
    ownerId: body.ownerId,
    ownerName: body.ownerName,
    config: body.config,
    tags: body.tags,
    metadata: body.metadata,
  });
  return c.json({ success: true, data: debate });
});

app.get('/', async (c) => {
  const status = c.req.query('status') as CollaborativeDebateStatus | undefined;
  const participantId = c.req.query('participantId');
  const taskType = c.req.query('taskType') as TaskType | undefined;
  const tag = c.req.query('tag');
  
  const debates = collaborativeDebateManager.listDebates({ status, participantId, taskType, tag });
  return c.json({ success: true, data: debates });
});

app.get('/stats', async (c) => {
  const stats = collaborativeDebateManager.getStatistics();
  return c.json({ success: true, data: stats });
});

app.get('/participant/:participantId', async (c) => {
  const debates = collaborativeDebateManager.getParticipantDebates(c.req.param('participantId'));
  return c.json({ success: true, data: debates });
});

app.get('/:id', async (c) => {
  const debate = collaborativeDebateManager.getDebate(c.req.param('id'));
  if (!debate) {
    return c.json({ success: false, error: 'Debate not found' }, 404);
  }
  return c.json({ success: true, data: debate });
});

app.post('/:id/start-inviting', async (c) => {
  try {
    const debate = collaborativeDebateManager.startInviting(c.req.param('id'));
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/start', async (c) => {
  try {
    const debate = collaborativeDebateManager.startDebate(c.req.param('id'));
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/start-voting', async (c) => {
  try {
    const debate = collaborativeDebateManager.startVoting(c.req.param('id'));
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/finalize', async (c) => {
  try {
    const debate = collaborativeDebateManager.finalizeDebate(c.req.param('id'));
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/cancel', async (c) => {
  try {
    const body = await c.req.json().catch(() => ({}));
    const debate = collaborativeDebateManager.cancelDebate(c.req.param('id'), body.reason);
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/:id/export', async (c) => {
  try {
    const data = collaborativeDebateManager.exportDebate(c.req.param('id'));
    return c.json({ success: true, data });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/import', async (c) => {
  try {
    const body = await c.req.json();
    const debate = collaborativeDebateManager.importDebate(body);
    return c.json({ success: true, data: debate });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Invites ============

app.post('/:id/invites', async (c) => {
  try {
    const body = await c.req.json();
    const token = collaborativeDebateManager.createInvite({
      debateId: c.req.param('id'),
      inviterId: body.inviterId,
      role: body.role ?? 'voter',
      email: body.email,
      expiresInMs: body.expiresInMs,
    });
    return c.json({ success: true, data: token });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/join', async (c) => {
  try {
    const body = await c.req.json();
    const result = collaborativeDebateManager.joinWithToken({
      token: body.token,
      participantId: body.participantId,
      participantName: body.participantName,
      email: body.email,
    });
    return c.json({ success: true, data: result });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Participants ============

app.post('/:id/participants', async (c) => {
  try {
    const body = await c.req.json();
    const participant = collaborativeDebateManager.addParticipantDirectly({
      debateId: c.req.param('id'),
      addedBy: body.addedBy,
      participantId: body.participantId,
      participantName: body.participantName,
      email: body.email,
      role: body.role ?? 'voter',
    });
    return c.json({ success: true, data: participant });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.delete('/:id/participants/:participantId', async (c) => {
  try {
    const body = await c.req.json();
    collaborativeDebateManager.removeParticipant(
      c.req.param('id'),
      c.req.param('participantId'),
      body.removedBy
    );
    return c.json({ success: true });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.patch('/:id/participants/:participantId/role', async (c) => {
  try {
    const body = await c.req.json();
    const participant = collaborativeDebateManager.updateParticipantRole(
      c.req.param('id'),
      c.req.param('participantId'),
      body.role as ParticipantRole,
      body.updatedBy
    );
    return c.json({ success: true, data: participant });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Messages ============

app.post('/:id/messages', async (c) => {
  try {
    const body = await c.req.json();
    const message = collaborativeDebateManager.addMessage({
      debateId: c.req.param('id'),
      participantId: body.participantId,
      content: body.content,
      replyTo: body.replyTo,
    });
    return c.json({ success: true, data: message });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/:id/messages', async (c) => {
  try {
    const since = c.req.query('since') ? parseInt(c.req.query('since')!) : undefined;
    const messages = collaborativeDebateManager.getMessages(c.req.param('id'), since);
    return c.json({ success: true, data: messages });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.patch('/:id/messages/:messageId', async (c) => {
  try {
    const body = await c.req.json();
    const message = collaborativeDebateManager.editMessage(
      c.req.param('id'),
      c.req.param('messageId'),
      body.participantId,
      body.content
    );
    return c.json({ success: true, data: message });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/messages/:messageId/reactions', async (c) => {
  try {
    const body = await c.req.json();
    collaborativeDebateManager.addReaction(
      c.req.param('id'),
      c.req.param('messageId'),
      body.participantId,
      body.reaction
    );
    return c.json({ success: true });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Voting ============

app.post('/:id/votes', async (c) => {
  try {
    const body = await c.req.json();
    const vote = collaborativeDebateManager.submitVote({
      debateId: c.req.param('id'),
      participantId: body.participantId,
      approved: body.approved,
      confidence: body.confidence,
      reasoning: body.reasoning,
      concerns: body.concerns,
      suggestions: body.suggestions,
    });
    return c.json({ success: true, data: vote });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/:id/votes', async (c) => {
  try {
    const requesterId = c.req.query('requesterId');
    if (!requesterId) {
      return c.json({ success: false, error: 'requesterId required' }, 400);
    }
    const votes = collaborativeDebateManager.getVotes(c.req.param('id'), requesterId);
    return c.json({ success: true, data: votes });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/:id/ai-votes', async (c) => {
  try {
    const body = await c.req.json();
    collaborativeDebateManager.setAISupervisorVotes(c.req.param('id'), body.votes);
    return c.json({ success: true });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

export default app;
