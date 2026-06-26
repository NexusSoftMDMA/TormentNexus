
import { EventBus } from '../services/EventBus.js';
import chokidar from 'chokidar';
import path from 'path';

export class FileSensor {
    private watcher: any = null;
    private eventBus: EventBus;
    private rootDir: string;

    private watchPaths: string[] = [];

    constructor(eventBus: EventBus, rootDir: string, watchPaths?: string[]) {
        this.eventBus = eventBus;
        this.rootDir = rootDir;
        if (watchPaths) {
            this.watchPaths = watchPaths.map(p => p.split(path.sep).join('/'));
        } else {
            // Default: Watch src directories for code changes
            this.watchPaths = [
                path.join(this.rootDir, 'packages', '*', 'src', '**', '*.ts').split(path.sep).join('/'),
                path.join(this.rootDir, 'apps', '*', 'src', '**', '*.tsx').split(path.sep).join('/')
            ];
        }
    }

    public start() {
        if (this.watcher) return;

        console.log("[FileSensor] Starting file system monitoring...");
        console.log("[FileSensor] Watching:", this.watchPaths);

        this.watcher = chokidar.watch(this.watchPaths, {
            ignored: /(^|[\/\\])\../, // ignore dotfiles
            persistent: true,
            ignoreInitial: true, // Don't emit for existing files
            usePolling: process.platform === 'win32', // Force polling on Windows for reliability
            awaitWriteFinish: {
                stabilityThreshold: 500, // Reduced for responsiveness
                pollInterval: 100
            }
        });

        this.watcher
            .on('add', (path: string) => { console.log('FileSensor: add', path); this.emit('file:create', path); })
            .on('change', (path: string) => { console.log('FileSensor: change', path); this.emit('file:change', path); })
            .on('unlink', (path: string) => { console.log('FileSensor: unlink', path); this.emit('file:delete', path); })
            .on('error', (error: any) => console.error(`[FileSensor] Watcher error: ${error}`));
    }

    public stop() {
        if (this.watcher) {
            this.watcher.close();
            this.watcher = null;
        }
    }

    private emit(type: string, filePath: string) {
        // Compute relative path for cleaner logs
        const relativePath = path.relative(this.rootDir, filePath);
        this.eventBus.emitEvent(type, 'FileSensor', { path: relativePath, absolutePath: filePath });
    }
}
