/**
 * SSE stream — `/chats/events`. Pushes chat-list mutations
 * (created/updated/deleted) so the cockpit sidebar shows freshly-fired
 * chats — including those created via MCP or external curl — without
 * waiting for a poll cycle.
 *
 * The payload is intentionally minimal: `{ chatId, kind, ts }`. The
 * client refetches the list on any event; we don't try to patch state
 * incrementally because the list is short (12 most recent) and the
 * round-trip is cheap.
 */

import type { FastifyInstance } from 'fastify';
import { chatEventsBus, type ChatChangeEvent } from '../../lib/chat-events-bus.js';

const HEARTBEAT_INTERVAL_MS = 25_000;

export function registerChatEventsRoute(fastify: FastifyInstance): void {
  fastify.get('/chats/events', async (_request, reply) => {
    reply.hijack();

    reply.raw.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    });

    // Initial hello so the client knows the channel is open. Some
    // browsers (and proxies) buffer the first chunk until enough bytes
    // arrive — sending a comment line forces the headers through.
    reply.raw.write(': connected\n\n');

    const onChange = (ev: ChatChangeEvent): void => {
      try {
        reply.raw.write(`data: ${JSON.stringify(ev)}\n\n`);
      } catch {
        /* connection closed mid-write */
      }
    };

    chatEventsBus.on('change', onChange);

    // Heartbeat keeps idle connections alive across NAT timeouts and
    // makes EventSource's 'error' fire promptly when the daemon dies.
    const heartbeat = setInterval(() => {
      try {
        reply.raw.write(': hb\n\n');
      } catch {
        /* ignore */
      }
    }, HEARTBEAT_INTERVAL_MS);

    const cleanup = (): void => {
      clearInterval(heartbeat);
      chatEventsBus.off('change', onChange);
      try {
        reply.raw.end();
      } catch {
        /* already closed */
      }
    };

    reply.raw.on('close', cleanup);
    reply.raw.on('error', cleanup);
  });
}
