import type { Server } from 'node:http';

interface AppLike {
    listen: (
        port: number,
        host: string,
        callback?: () => void,
    ) => Server;
}

export async function listenExpress(app: AppLike, port: number, host: string): Promise<Server> {
    return await new Promise<Server>((resolve, reject) => {
        const server = app.listen(port, host);

        const onError = (error: Error) => {
            cleanup();
            reject(error);
        };

        const onListening = () => {
            cleanup();
            resolve(server);
        };

        const cleanup = () => {
            server.off('error', onError);
            server.off('listening', onListening);
        };

        server.once('error', onError);
        server.once('listening', onListening);
    });
}
