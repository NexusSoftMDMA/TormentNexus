
import { Transport } from "@modelcontextprotocol/sdk/shared/transport.js";
import { JSONRPCMessage } from "@modelcontextprotocol/sdk/types.js";
import { WebSocket, WebSocketServer } from 'ws';
import { IncomingMessage } from 'http';

export class WebSocketServerTransport implements Transport {
    private _ws: WebSocket | null = null;
    private _wss: WebSocketServer;

    onclose?: () => void;
    onerror?: (error: Error) => void;
    onmessage?: (message: JSONRPCMessage) => void;

    constructor(wss: WebSocketServer) {
        this._wss = wss;
        this._wss.on('connection', (ws: WebSocket, req: IncomingMessage) => {
            console.log("New WebSocket connection");
            this._ws = ws;

            ws.on('message', (data) => {
                try {
                    const str = data.toString();
                    const message = JSON.parse(str);
                    this.onmessage?.(message);
                } catch (e) {
                    console.error("Failed to parse WS message", e);
                }
            });

            ws.on('close', () => {
                console.log("WebSocket disconnected");
                this.onclose?.();
            });

            ws.on('error', (err) => {
                this.onerror?.(err);
            });
        });
    }

    async close(): Promise<void> {
        this._ws?.close();
        this._wss.close();
    }

    async send(message: JSONRPCMessage): Promise<void> {
        if (this._ws && this._ws.readyState === WebSocket.OPEN) {
            this._ws.send(JSON.stringify(message));
        }
    }

    async start(): Promise<void> {
        // Already started via constructor listener
    }
}
