
import fs from 'fs';
import path from 'path';

export interface SessionState {
    isAutoDriveActive: boolean;
    activeGoal: string | null;
    lastObjective: string | null;
    lastHeartbeat: number;
    threadId?: string;
}

export class SessionManager {
    private persistencePath: string;
    private state: SessionState;
    private saveInterval: NodeJS.Timeout | null = null;

    constructor(baseDir: string = process.cwd()) {
        this.persistencePath = path.join(baseDir, '.tormentnexus-session.json');
        this.state = {
            isAutoDriveActive: false,
            activeGoal: null,
            lastObjective: null,
            lastHeartbeat: Date.now()
        };
        this.load();
        this.startAutoSave();
    }

    private load() {
        try {
            if (fs.existsSync(this.persistencePath)) {
                const data = fs.readFileSync(this.persistencePath, 'utf-8');
                this.state = JSON.parse(data);
                console.log(`[SessionManager] Loaded session state from ${this.persistencePath}`);
            }
        } catch (e) {
            console.error(`[SessionManager] Failed to load session:`, e);
        }
    }

    public save() {
        try {
            fs.writeFileSync(this.persistencePath, JSON.stringify(this.state, null, 2));
        } catch (e) {
            console.error(`[SessionManager] Failed to save session:`, e);
        }
    }

    private startAutoSave() {
        this.saveInterval = setInterval(() => {
            this.save();
        }, 5000); // Save every 5s
    }

    public getState(): SessionState {
        return { ...this.state };
    }

    public updateState(updates: Partial<SessionState>) {
        this.state = { ...this.state, ...updates, lastHeartbeat: Date.now() };
    }

    public clearSession() {
        this.state = {
            isAutoDriveActive: false,
            activeGoal: null,
            lastObjective: null,
            lastHeartbeat: Date.now()
        };
        this.save();
    }

    public disableAutoDriveRestore() {
        this.state = {
            ...this.state,
            isAutoDriveActive: false,
            activeGoal: null,
            lastHeartbeat: Date.now(),
        };
        this.save();
    }

    public touch() {
        this.state.lastHeartbeat = Date.now();
    }

    public dispose() {
        if (this.saveInterval) {
            clearInterval(this.saveInterval);
            this.saveInterval = null;
        }
    }
}
