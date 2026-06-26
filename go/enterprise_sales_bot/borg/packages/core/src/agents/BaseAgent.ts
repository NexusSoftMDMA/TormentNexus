
import { v4 as uuidv4 } from 'uuid';

export enum AgentStatus {
    IDLE = 'idle',
    RUNNING = 'running',
    COMPLETED = 'completed',
    FAILED = 'failed'
}

export interface AgentResult {
    success: boolean;
    output: string;
    error?: string;
}

export abstract class BaseAgent {
    public id: string;
    public type: string;
    public status: AgentStatus = AgentStatus.IDLE;
    public task: string;
    public result: AgentResult | null = null;
    public createdAt: Date;
    public logs: string[] = [];

    constructor(type: string, task: string) {
        this.id = uuidv4();
        this.type = type;
        this.task = task;
        this.createdAt = new Date();
    }

    abstract run(): Promise<void>;

    protected log(message: string) {
        const timestamp = new Date().toISOString();
        const logEntry = `[${timestamp}] [${this.type}:${this.id.substring(0, 8)}] ${message}`;
        this.logs.push(logEntry);
        console.log(logEntry);
    }

    protected complete(output: string) {
        this.status = AgentStatus.COMPLETED;
        this.result = { success: true, output };
        this.log(`Task completed: ${output.substring(0, 100)}...`);
    }

    protected fail(error: string) {
        this.status = AgentStatus.FAILED;
        this.result = { success: false, output: '', error };
        this.log(`Task failed: ${error}`);
    }
}
