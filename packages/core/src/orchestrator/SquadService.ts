import path from "path";
import { Director } from "@tormentnexus/agents";
import type { IMCPServer } from "@tormentnexus/adk";

interface ServerWithCouncil {
    council: unknown;
}

interface ServerWithMemoryManager {
    memoryManager: unknown;
}

interface IndexerJobController {
    start(): void;
    stop(): void;
    getStatus(): { running: boolean; indexing: boolean };
}

interface IndexerJobModule {
    IndexerJob: new (memoryManager: unknown, rootPath: string) => IndexerJobController;
}

/**
 * Proxy that intercepts tool calls and injects CWD/Path context
 * to strictly bind an Agent to a Git Worktree.
 */
class WorktreeServerProxy implements IMCPServer {
    constructor(private server: IMCPServer, private worktreeRoot: string) { }

    /**
     * Reason: IMCPServer does not guarantee a `council` property, but some runtime servers expose it.
     * What: Reflection-safe accessor that returns council when available.
     * Why: Preserves compatibility with richer server runtimes without broad casts.
     */
    private getCouncilRuntime(): unknown {
        if (!this.server || typeof this.server !== 'object') {
            return undefined;
        }
        return Reflect.get(this.server as object, 'council');
    }

    // Passthrough properties
    get modelSelector() { return this.server.modelSelector; }
    get permissionManager() { return this.server.permissionManager; }
    get directorConfig() { return this.server.directorConfig; }
    get council() { return this.getCouncilRuntime(); }

  async handleRequest(request: any) { return this.server.handleRequest(request); }
  get name() { return this.server.name; }
  get version() { return this.server.version; }
  async getStatus() { return this.server.getStatus(); }
  getTools() { return this.server.getTools(); }
  async start() { return this.server.start(); }
  async stop() { return this.server.stop(); }
    async executeTool(name: string, args: any): Promise<any> {
        // Intercept Path-based tools
        const pathTools = ['read_file', 'write_file', 'list_files', 'delete_file', 'replace_in_file', 'create_directory', 'move_file', 'copy_file', 'git_worktree_add', 'git_worktree_remove'];

        let newArgs = { ...args };

        // 1. Resolve 'path' argument against worktreeRoot
        if (args && args.path && typeof args.path === 'string') {
            // If relative, resolve. If absolute, ensure it starts with worktreeRoot (Security?)
            // For now, assume cooperative agent (relative paths preferred)
            if (!path.isAbsolute(args.path)) {
                newArgs.path = path.join(this.worktreeRoot, args.path);
                console.log(`[WorktreeProxy] Resolved path: ${args.path} -> ${newArgs.path}`);
            }
        }

        // 2. Inject 'cwd' for command tools
        const cwdTools = ['execute_command', 'git_worktree_list', 'git_worktree_add', 'git_worktree_remove', 'start_autotest', 'list_files']; // list_files doesn't take cwd, handles path above
        if (cwdTools.includes(name)) {
            newArgs.cwd = this.worktreeRoot;
            console.log(`[WorktreeProxy] Injected CWD: ${newArgs.cwd}`);
        }

        return this.server.executeTool(name, newArgs);
    }
}

interface SquadMember {
    id: string;
    branch: string;
    worktreePath: string;
    director: Director;
    status: 'idle' | 'busy' | 'finished';
}

export class SquadService {
    private members: Map<string, SquadMember> = new Map();
    private worktreesDir: string;

    constructor(private server: IMCPServer & { modelSelector?: any; permissionManager?: any; directorConfig?: any; }) {
        this.worktreesDir = path.join(process.cwd(), '.tormentnexus', 'worktrees');
    }

    /**
     * Spawn a new Agent Squad Member in an isolated Git Worktree
     */
    async spawnMember(branchName: string, goal: string): Promise<string> {
        console.log(`[SquadService] Spawning member for branch '${branchName}'...`);

        // 1. Prepare Worktree
        const worktreePath = path.join(this.worktreesDir, branchName);

        // Create dir if needed (parent)
        // Actually git worktree add handles dir creation, but parent must exist
        // Use server tool to create worktree
        try {
            await this.server.executeTool('git_worktree_add', {
                path: worktreePath,
                branch: branchName,
                cwd: process.cwd() // Run from Main Repo Root
            });
        } catch (e: any) {
            console.error(`[SquadService] Worktree creation failed (might exist): ${e.message}`);
        }

        // 2. Create Proxy Server
        const proxy = new WorktreeServerProxy(this.server, worktreePath);

        // 3. Create Director
        const director = new Director(proxy);

        // 4. Register
        const member: SquadMember = {
            id: `squad-${branchName}`,
            branch: branchName,
            worktreePath,
            director,
            status: 'busy'
        };
        this.members.set(member.id, member);

        // 5. Start Task (Async / Fire & Forget)
        console.log(`[SquadService] 🚀 Starting Director in ${worktreePath}`);

        // We don't await the whole task, we just kick it off
        // But Director.executeTask is async.
        // We'll wrap it to track status
        director.executeTask(goal, 20).then(result => {
            console.log(`[SquadService] Member ${member.id} finished: ${result}`);
            member.status = 'finished';
        }).catch(err => {
            console.error(`[SquadService] Member ${member.id} crashed: ${err.message}`);
            member.status = 'idle'; // Or error state
        });

        return `Spawned Squad Member ${member.id} in ${worktreePath}`;
    }

    listMembers() {
        return Array.from(this.members.values()).map(m => ({
            id: m.id,
            branch: m.branch,
            status: m.status,
            active: m.director.getIsActive(),
            brain: m.director.getStatus() // Expose full brain state (Goal, Step, History)
        }));
    }

    async killMember(branchName: string) {
        const id = `squad-${branchName}`;
        const member = this.members.get(id);
        if (member) {
            // Stop Director? Director doesn't have an abort controller yet for executeTask loop.
            // But we can stop auto drive.
            member.director.stopAutoDrive();

            // Cleanup Worktree
            await this.server.executeTool('git_worktree_remove', {
                path: member.worktreePath,
                force: true,
                cwd: process.cwd()
            });

            this.members.delete(id);
            return `Killed member ${id} and removed worktree.`;
        }
        return `Member ${id} not found.`;
    }

    async messageMember(branchName: string, message: string) {
        const id = `squad-${branchName}`;
        const member = this.members.get(id);
        if (member) {
            return await member.director.handleUserMessage(message);
        }
        return `Member ${id} not found.`;
    }

    // --- Job Management (Indexer, etc) ---

    private indexerJob: IndexerJobController | null = null;

    /**
     * Reason: indexer wiring depends on optional runtime server capabilities not present in the base interface.
     * What: Narrow server to memory-manager capable runtime via property existence check.
     * Why: Avoids broad casts while keeping lazy indexer boot behavior unchanged.
     */
    private getMemoryManagerRuntime(): unknown {
        if (!this.server || typeof this.server !== 'object') {
            return undefined;
        }
        if (!Reflect.has(this.server as object, 'memoryManager')) {
            return undefined;
        }
        return Reflect.get(this.server as object, 'memoryManager');
    }

    public async toggleIndexer(enabled: boolean) {
        if (enabled) {
            if (!this.indexerJob) {
                const { IndexerJob } = await import('../jobs/IndexerJob.js') as unknown as IndexerJobModule;
                const memoryManager = this.getMemoryManagerRuntime();
                if (memoryManager) {
                    this.indexerJob = new IndexerJob(memoryManager, process.cwd());
                }
            }
            if (this.indexerJob) {
                this.indexerJob.start();
                return true;
            }
        } else {
            if (this.indexerJob) {
                this.indexerJob.stop();
                return false;
            }
        }
        return false;
    }

    public getIndexerStatus() {
        if (!this.indexerJob) return { running: false, indexing: false };
        return this.indexerJob.getStatus();
    }
}
