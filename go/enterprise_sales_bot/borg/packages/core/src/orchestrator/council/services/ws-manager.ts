import type { WebSocketMessage, Session, CouncilDecision, LogEntry } from './types.js';

type WSClient = {
  ws: WebSocket;
  sessionId?: string;
};

class WebSocketManager {
  private clients = new Set<WSClient>();

  add(ws: WebSocket, sessionId?: string) {
    const client: WSClient = { ws, sessionId };
    this.clients.add(client);
    return client;
  }

  remove(client: WSClient) {
    this.clients.delete(client);
  }

  broadcast(message: WebSocketMessage) {
    const payload = JSON.stringify(message);
    for (const client of this.clients) {
      if (client.ws.readyState === WebSocket.OPEN) {
        client.ws.send(payload);
      }
    }
  }

  broadcastToSession(sessionId: string, message: WebSocketMessage) {
    const payload = JSON.stringify(message);
    for (const client of this.clients) {
      if (client.ws.readyState === WebSocket.OPEN && client.sessionId === sessionId) {
        client.ws.send(payload);
      }
    }
  }

  notifySessionUpdate(session: Session) {
    this.broadcast({
      type: 'session_update',
      payload: session,
      timestamp: Date.now(),
    });
  }

  notifySessionStarted(session: Session) {
    this.notifySessionUpdate(session);
  }

  notifySessionStopped(sessionId: string) {
    this.broadcast({
      type: 'session_update',
      payload: { id: sessionId, status: 'stopped' },
      timestamp: Date.now(),
    });
  }

  notifyCouncilDecision(sessionId: string, decision: CouncilDecision) {
    this.broadcastToSession(sessionId, {
      type: 'council_decision',
      payload: decision,
      timestamp: Date.now(),
    });
  }

  notifyLog(sessionId: string, log: LogEntry) {
    this.broadcastToSession(sessionId, {
      type: 'log',
      payload: log,
      timestamp: Date.now(),
    });
  }

  notifyError(error: string) {
    this.broadcast({
      type: 'error',
      payload: { message: error },
      timestamp: Date.now(),
    });
  }
}

export const wsManager = new WebSocketManager();
