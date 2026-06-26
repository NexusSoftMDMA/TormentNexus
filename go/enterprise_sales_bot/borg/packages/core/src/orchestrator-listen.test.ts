import net from 'node:net';

import express from 'express';
import { afterEach, describe, expect, it } from 'vitest';

import { listenExpress } from './orchestrator-listen.js';

const serversToClose: net.Server[] = [];

afterEach(async () => {
    while (serversToClose.length > 0) {
        const server = serversToClose.pop();
        if (!server) {
            continue;
        }

        await new Promise<void>((resolveClose) => {
            server.close(() => resolveClose());
        });
    }
});

async function reservePort(): Promise<{ port: number; server: net.Server }> {
    return await new Promise((resolvePort, rejectPort) => {
        const server = net.createServer();
        server.once('error', rejectPort);
        server.listen(0, '127.0.0.1', () => {
            const address = server.address();
            if (!address || typeof address === 'string') {
                rejectPort(new Error('Unable to reserve port'));
                return;
            }

            resolvePort({ port: address.port, server });
        });
    });
}

describe('listenExpress', () => {
    it('resolves when express binds successfully', async () => {
        const app = express();
        const server = await listenExpress(app, 0, '127.0.0.1');
        serversToClose.push(server);

        const address = server.address();
        expect(address).toBeTruthy();
        expect(typeof address === 'string' ? null : address?.port).toBeGreaterThan(0);
    });

    it('rejects with EADDRINUSE when the target port is occupied', async () => {
        const { port, server: occupiedServer } = await reservePort();
        serversToClose.push(occupiedServer);

        const app = express();

        await expect(listenExpress(app, port, '127.0.0.1')).rejects.toMatchObject({
            code: 'EADDRINUSE',
        });
    });
});
