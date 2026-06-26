import { Hono } from 'hono';
import { upgradeWebSocket } from 'hono/bun';
import { wsManager } from '../services/ws-manager.js';

const ws = new Hono();

ws.get(
  '/',
  upgradeWebSocket((c) => {
    let client: ReturnType<typeof wsManager.add> | null = null;

    return {
      onOpen(wsConn) {
        const sessionId = c.req.query('sessionId');
        client = wsManager.add(wsConn as unknown as WebSocket, sessionId);
        console.log(`[WS] Client connected${sessionId ? ` (session: ${sessionId})` : ''}`);
      },

      onMessage(event) {
        try {
          const data = JSON.parse(event.data.toString());
          
          if (data.type === 'subscribe' && data.sessionId && client) {
            client.sessionId = data.sessionId;
            console.log(`[WS] Client subscribed to session: ${data.sessionId}`);
          }
        } catch {
          console.error('[WS] Invalid message format');
        }
      },

      onClose() {
        if (client) {
          wsManager.remove(client);
          console.log('[WS] Client disconnected');
        }
      },

      onError(error) {
        console.error('[WS] Error:', error);
      },
    };
  })
);

export { ws as wsRoutes };
