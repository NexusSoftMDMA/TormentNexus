/**
 * Process-local event bus for chat list mutations.
 *
 * The cockpit sidebar (and any other surface that renders a list of
 * chats) needs to react when a chat is created/updated/deleted from
 * ANY source — REST POST, MCP tool call, runner status change. The
 * daemon polled every 12s before, which is fine for status drift but
 * looks broken when a freshly-fired MCP chat takes 12s to appear.
 *
 * We chose a process-local EventEmitter over a DB-NOTIFY broadcast
 * because chorus is single-daemon today; if we ever shard the daemon,
 * this becomes a libsql LISTEN. Subscribers are SSE handlers in
 * `routes/chats-events.ts`.
 */

import { EventEmitter } from 'node:events';

export type ChatChangeKind = 'created' | 'updated' | 'deleted';

export interface ChatChangeEvent {
  chatId: string;
  kind: ChatChangeKind;
  ts: number;
}

class ChatEventsBus extends EventEmitter {
  emitChange(chatId: string, kind: ChatChangeKind): void {
    const ev: ChatChangeEvent = { chatId, kind, ts: Date.now() };
    this.emit('change', ev);
  }
}

export const chatEventsBus = new ChatEventsBus();
chatEventsBus.setMaxListeners(50);
