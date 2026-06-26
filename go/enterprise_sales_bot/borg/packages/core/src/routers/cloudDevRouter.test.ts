import { beforeEach, describe, expect, it } from 'vitest';

import { cloudDevRouter } from './cloudDevRouter.js';

function createCaller() {
    return cloudDevRouter.createCaller({} as never);
}

describe('cloudDevRouter force send and broadcast behavior', () => {
    beforeEach(async () => {
        const caller = createCaller();
        const sessions = await caller.listSessions();
        await Promise.all(sessions.map((session) => caller.deleteSession({ sessionId: session.id })));
    });

    it('blocks sending to terminal sessions unless force=true', async () => {
        const caller = createCaller();

        const created = await caller.createSession({
            provider: 'jules',
            projectName: 'force-send-regression',
            task: 'validate terminal send gate',
            autoAcceptPlan: false,
        });

        await caller.updateSessionStatus({ sessionId: created.id, status: 'completed' });

        await expect(
            caller.sendMessage({
                sessionId: created.id,
                content: 'should fail without force',
                force: false,
            })
        ).rejects.toThrow(/Use force:true/i);

        const forced = await caller.sendMessage({
            sessionId: created.id,
            content: 'forced terminal follow-up',
            force: true,
        });

        expect(forced.message.forceSent).toBe(true);

        const messages = await caller.getMessages({ sessionId: created.id, limit: 10 });
        expect(messages[messages.length - 1]).toMatchObject({
            role: 'user',
            content: 'forced terminal follow-up',
            forceSent: true,
        });
    });

    it('broadcast skips terminal sessions by default and includes them with force', async () => {
        const caller = createCaller();

        const active = await caller.createSession({
            provider: 'jules',
            projectName: 'broadcast-active',
            task: 'active target',
            autoAcceptPlan: false,
        });
        const completed = await caller.createSession({
            provider: 'codex',
            projectName: 'broadcast-completed',
            task: 'terminal target',
            autoAcceptPlan: false,
        });

        await caller.updateSessionStatus({ sessionId: active.id, status: 'active' });
        await caller.updateSessionStatus({ sessionId: completed.id, status: 'completed' });

        const noForce = await caller.broadcastMessage({
            content: 'default broadcast',
            force: false,
        });

        expect(noForce.delivered).toBe(1);
        expect(noForce.results.map((entry) => entry.sessionId)).toEqual([active.id]);
        expect(noForce.skippedByReason.terminal_requires_force).toBe(1);

        const forced = await caller.broadcastMessage({
            content: 'forced broadcast',
            force: true,
        });

        expect(forced.delivered).toBe(2);
        expect(new Set(forced.results.map((entry) => entry.sessionId))).toEqual(new Set([active.id, completed.id]));

        const completedMessages = await caller.getMessages({ sessionId: completed.id, limit: 20 });
        const terminalForced = completedMessages.filter((message) => message.content === 'forced broadcast');
        expect(terminalForced.length).toBe(1);
        expect(terminalForced[0]?.forceSent).toBe(true);
    });
});
