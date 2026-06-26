/**
 * Workflow Engine - Graph-Based Multi-Agent Orchestration
 * 
 * Implements graph-based workflow engine for multi-agent orchestration
 * with shared state, durable execution, and flexible control flow.
 * 
 * Features:
 * - Graph definition (nodes + edges)
 * - Agent nodes with input/output
 * - Conditional edges (LLM-based routing)
 * - Shared centralized state
 * - Durable execution (resume from failure)
 * - Human-in-the-loop checkpoints
 * 
 * Inspired by LangGraph architecture.
 */

import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';
import { EventEmitter } from 'events';

// State types
export type WorkflowState = Record<string, unknown>;

// Node types
export type NodeFunction = (state: WorkflowState) => Promise<WorkflowState> | WorkflowState;

export interface NodeDefinition {
    id: string;
    name: string;
    description?: string;
    fn: NodeFunction;
    requiresApproval?: boolean;  // HITL checkpoint
}

// Edge types
export type EdgeCondition = (state: WorkflowState) => boolean | Promise<boolean>;
export type EdgeRouter = (state: WorkflowState) => string | Promise<string>;  // Returns next node ID

export interface EdgeDefinition {
    id: string;
    from: string;
    to: string | EdgeRouter;  // Direct target or router function
    condition?: EdgeCondition;  // Optional condition
}

export interface WorkflowGraphNode {
    id: string;
    label: string;
    data: { description?: string };
    type: 'checkpoint' | 'default';
}

export interface WorkflowGraphEdge {
    id: string;
    source: string;
    target: string;
    animated: boolean;
    label?: string;
}

export interface WorkflowGraph {
    nodes: WorkflowGraphNode[];
    edges: WorkflowGraphEdge[];
}

function getWorkflowGraphNodeType(requiresApproval?: boolean): WorkflowGraphNode['type'] {
    return requiresApproval ? 'checkpoint' : 'default';
}

// Workflow execution types
export interface WorkflowExecution {
    id: string;
    workflowId: string;
    state: WorkflowState;
    currentNode: string;
    history: ExecutionStep[];
    status: 'running' | 'paused' | 'completed' | 'failed' | 'awaiting_approval';
    createdAt: Date;
    updatedAt: Date;
    error?: string;
}

export interface ExecutionStep {
    nodeId: string;
    nodeName: string;
    inputState: WorkflowState;
    outputState: WorkflowState;
    startTime: Date;
    endTime: Date;
    status: 'success' | 'failed' | 'skipped';
    error?: string;
}

// Workflow definition
export interface WorkflowDefinition {
    id: string;
    name: string;
    description?: string;
    entryPoint: string;  // Starting node ID
    nodes: Map<string, NodeDefinition>;
    edges: EdgeDefinition[];
}

// Workflow options
export interface WorkflowEngineOptions {
    persistDir?: string;
    maxHistory?: number;
    timeoutMs?: number;
}

/**
 * State Store - Persistent state management with snapshots
 * Refactored to use individual files per execution for scalability.
 */
export class StateStore {
    private states = new Map<string, WorkflowState>();
    private snapshots = new Map<string, WorkflowState[]>();
    private readonly maxLoadedEntries = 200;

    constructor(private persistDir?: string) {
        if (persistDir) {
            this.ensurePersistDir();
            // We don't load all states into memory at startup anymore to save RAM.
            // Entries are loaded on demand and bounded via LRU eviction.
        }
    }

    private markRecentlyUsed(executionId: string): void {
        const state = this.states.get(executionId);
        if (!state) {
            return;
        }

        this.states.delete(executionId);
        this.states.set(executionId, state);

        const snapshots = this.snapshots.get(executionId);
        if (snapshots) {
            this.snapshots.delete(executionId);
            this.snapshots.set(executionId, snapshots);
        }
    }

    private enforceMemoryLimit(): void {
        while (this.states.size > this.maxLoadedEntries) {
            const oldestExecutionId = this.states.keys().next().value as string | undefined;
            if (!oldestExecutionId) {
                break;
            }

            this.states.delete(oldestExecutionId);
            this.snapshots.delete(oldestExecutionId);
        }
    }

    private ensurePersistDir(): void {
        if (this.persistDir && !fs.existsSync(this.persistDir)) {
            fs.mkdirSync(this.persistDir, { recursive: true });
        }
    }

    private getFilePath(id: string): string {
        if (!this.persistDir) throw new Error("Persistence disabled");
        return path.join(this.persistDir, `${id}.json`);
    }

    private persistState(id: string): void {
        if (!this.persistDir) return;
        try {
            const filePath = this.getFilePath(id);
            const data = {
                state: this.states.get(id),
                snapshots: this.snapshots.get(id)
            };
            fs.writeFileSync(filePath, JSON.stringify(data, null, 2));
        } catch (e) {
            console.error(`[StateStore] Failed to persist state ${id}:`, e);
        }
    }

    private loadState(id: string): void {
        if (!this.persistDir) return;
        if (this.states.has(id)) {
            this.markRecentlyUsed(id);
            return;
        }

        const filePath = this.getFilePath(id);
        if (fs.existsSync(filePath)) {
            try {
                const data = JSON.parse(fs.readFileSync(filePath, 'utf-8'));
                if (data.state) this.states.set(id, data.state);
                if (data.snapshots) this.snapshots.set(id, data.snapshots);
                this.markRecentlyUsed(id);
                this.enforceMemoryLimit();
            } catch (e) {
                console.error(`[StateStore] Failed to load state ${id}:`, e);
            }
        }
    }

    /**
     * Get state by execution ID
     */
    get(executionId: string): WorkflowState | undefined {
        this.loadState(executionId);
        this.markRecentlyUsed(executionId);
        return this.states.get(executionId);
    }

    /**
     * Set state for an execution
     */
    set(executionId: string, state: WorkflowState): void {
        this.states.set(executionId, { ...state });
        this.markRecentlyUsed(executionId);
        this.enforceMemoryLimit();
        this.persistState(executionId);
    }

    /**
     * Update state (merge with existing)
     */
    update(executionId: string, updates: Partial<WorkflowState>): WorkflowState {
        const current = this.get(executionId) || {};
        const updated = { ...current, ...updates };
        this.set(executionId, updated);
        return updated;
    }

    /**
     * Create a snapshot of current state
     */
    snapshot(executionId: string): void {
        const state = this.get(executionId);
        if (!state) return;

        if (!this.snapshots.has(executionId)) {
            this.snapshots.set(executionId, []);
        }
        this.snapshots.get(executionId)!.push({ ...state });
        this.markRecentlyUsed(executionId);
        this.persistState(executionId);
    }

    /**
     * Restore from a snapshot
     */
    restore(executionId: string, snapshotIndex?: number): WorkflowState | undefined {
        this.loadState(executionId);
        const history = this.snapshots.get(executionId);
        if (!history || history.length === 0) return undefined;

        const index = snapshotIndex ?? history.length - 1;
        const state = history[index];
        if (state) {
            this.set(executionId, state);
        }
        return state;
    }

    /**
     * Get snapshot history
     */
    getSnapshots(executionId: string): WorkflowState[] {
        this.loadState(executionId);
        return this.snapshots.get(executionId) || [];
    }

    /**
     * Delete state
     */
    delete(executionId: string): void {
        this.states.delete(executionId);
        this.snapshots.delete(executionId);
        if (this.persistDir) {
            const filePath = this.getFilePath(executionId);
            if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
        }
    }
}

/**
 * Graph Builder - Build workflow graphs fluently
 */
export class GraphBuilder {
    private nodes = new Map<string, NodeDefinition>();
    private edges: EdgeDefinition[] = [];
    private entryPoint: string | null = null;

    /**
     * Add a node to the graph
     */
    addNode(
        id: string,
        fn: NodeFunction,
        options: { name?: string; description?: string; requiresApproval?: boolean } = {}
    ): GraphBuilder {
        this.nodes.set(id, {
            id,
            name: options.name || id,
            description: options.description,
            fn,
            requiresApproval: options.requiresApproval,
        });

        // First node becomes entry point by default
        if (!this.entryPoint) {
            this.entryPoint = id;
        }

        return this;
    }

    /**
     * Set the entry point node
     */
    setEntryPoint(nodeId: string): GraphBuilder {
        if (!this.nodes.has(nodeId)) {
            throw new Error(`Node ${nodeId} not found`);
        }
        this.entryPoint = nodeId;
        return this;
    }

    /**
     * Add a direct edge between nodes
     */
    addEdge(from: string, to: string): GraphBuilder {
        this.edges.push({
            id: `edge-${from}-${to}`,
            from,
            to,
        });
        return this;
    }

    /**
     * Add a conditional edge
     */
    addConditionalEdge(from: string, to: string, condition: EdgeCondition): GraphBuilder {
        this.edges.push({
            id: `edge-${from}-${to}-conditional`,
            from,
            to,
            condition,
        });
        return this;
    }

    /**
     * Add a routing edge (dynamic next node selection)
     */
    addRouterEdge(from: string, router: EdgeRouter): GraphBuilder {
        this.edges.push({
            id: `edge-${from}-router`,
            from,
            to: router,
        });
        return this;
    }

    /**
     * Get the graph structure for UI visualization (React Flow compatible)
     */
    getGraph(): WorkflowGraph {
        const nodes = Array.from(this.nodes.values()).map(n => ({
            id: n.id,
            label: n.name,
            data: { description: n.description },
            type: getWorkflowGraphNodeType(n.requiresApproval)
        }));

        const edges = this.edges.map(e => ({
            id: e.id,
            source: e.from,
            target: typeof e.to === 'string' ? e.to : 'dynamic',
            animated: !!e.condition,
            label: e.condition ? 'Conditional' : undefined
        }));

        return { nodes, edges };
    }

    /**
     * Build the workflow definition
     */
    build(id: string, name?: string, description?: string): WorkflowDefinition {
        if (!this.entryPoint) {
            throw new Error('Workflow must have an entry point');
        }
        if (this.nodes.size === 0) {
            throw new Error('Workflow must have at least one node');
        }

        return {
            id,
            name: name || id,
            description,
            entryPoint: this.entryPoint,
            nodes: new Map(this.nodes),
            edges: [...this.edges],
        };
    }
}

/**
 * Workflow Engine - Execute workflow graphs
 */
export class WorkflowEngine extends EventEmitter {
    private workflows = new Map<string, WorkflowDefinition>();
    private executions = new Map<string, WorkflowExecution>();
    private stateStore: StateStore;
    private options: Required<WorkflowEngineOptions>;

    constructor(options: WorkflowEngineOptions = {}) {
        super();
        this.options = {
            persistDir: options.persistDir || '',
            maxHistory: options.maxHistory ?? 100,
            timeoutMs: options.timeoutMs ?? 300000,  // 5 minutes default
        };

        this.stateStore = new StateStore(
            this.options.persistDir ? path.join(this.options.persistDir, 'states') : undefined
        );
    }

    /**
     * Get graph definition for UI
     */
    getGraph(workflowId: string): WorkflowGraph | undefined {
        const workflow = this.workflows.get(workflowId);
        if (!workflow) return undefined;

        const nodes = Array.from(workflow.nodes.values()).map(n => ({
            id: n.id,
            label: n.name,
            data: { description: n.description },
            type: getWorkflowGraphNodeType(n.requiresApproval)
        }));

        const edges = workflow.edges.map(e => ({
            id: e.id,
            source: e.from,
            target: typeof e.to === 'string' ? e.to : 'dynamic',
            animated: !!e.condition,
            label: e.condition ? 'Conditional' : undefined
        }));

        return { nodes, edges };
    }

    /**
     * Register a workflow definition
     */
    registerWorkflow(workflow: WorkflowDefinition): void {
        this.workflows.set(workflow.id, workflow);
    }

    /**
     * Create a new graph builder
     */
    static createGraph(): GraphBuilder {
        return new GraphBuilder();
    }

    /**
     * Generate unique execution ID
     */
    private generateExecutionId(): string {
        return crypto.randomBytes(8).toString('hex');
    }

    /**
     * Start a new workflow execution
     */
    async start(workflowId: string, initialState: WorkflowState = {}): Promise<WorkflowExecution> {
        const workflow = this.workflows.get(workflowId);
        if (!workflow) {
            throw new Error(`Workflow ${workflowId} not found`);
        }

        const executionId = this.generateExecutionId();
        const execution: WorkflowExecution = {
            id: executionId,
            workflowId,
            state: initialState,
            currentNode: workflow.entryPoint,
            history: [],
            status: 'running',
            createdAt: new Date(),
            updatedAt: new Date(),
        };

        this.executions.set(executionId, execution);
        this.stateStore.set(executionId, initialState);

        this.emit('execution:start', execution);

        // Run the workflow
        await this.run(executionId);

        return this.executions.get(executionId)!;
    }

    /**
     * Run execution until completion or pause
     */
    private async run(executionId: string): Promise<void> {
        const execution = this.executions.get(executionId);
        if (!execution || execution.status !== 'running') return;

        const workflow = this.workflows.get(execution.workflowId);
        if (!workflow) return;

        while (execution.status === 'running') {
            const node = workflow.nodes.get(execution.currentNode);
            if (!node) {
                execution.status = 'failed';
                execution.error = `Node ${execution.currentNode} not found`;
                break;
            }

            // Check for HITL checkpoint
            if (node.requiresApproval) {
                execution.status = 'awaiting_approval';
                execution.updatedAt = new Date();
                this.stateStore.snapshot(executionId);
                this.emit('execution:awaiting_approval', execution, node);
                return;
            }

            // Execute the node
            const step = await this.executeNode(executionId, node);
            execution.history.push(step);

            // Trim history if needed
            if (execution.history.length > this.options.maxHistory) {
                execution.history = execution.history.slice(-this.options.maxHistory);
            }

            if (step.status === 'failed') {
                execution.status = 'failed';
                execution.error = step.error;
                break;
            }

            // Update state
            execution.state = step.outputState;
            this.stateStore.set(executionId, execution.state);

            // Find next node
            const nextNode = await this.getNextNode(workflow, execution.currentNode, execution.state);

            if (!nextNode) {
                // No next node means workflow is complete
                execution.status = 'completed';
                break;
            }

            execution.currentNode = nextNode;
            execution.updatedAt = new Date();
        }

        this.emit('execution:end', execution);
    }

    /**
     * Execute a single node
     */
    private async executeNode(executionId: string, node: NodeDefinition): Promise<ExecutionStep> {
        const execution = this.executions.get(executionId)!;
        const inputState = { ...execution.state };
        const startTime = new Date();

        try {
            const result = await Promise.race([
                node.fn(inputState),
                new Promise<never>((_, reject) =>
                    setTimeout(() => reject(new Error('Node execution timeout')), this.options.timeoutMs)
                ),
            ]);

            const outputState = result as WorkflowState;

            this.emit('node:complete', executionId, node, outputState);

            return {
                nodeId: node.id,
                nodeName: node.name,
                inputState,
                outputState,
                startTime,
                endTime: new Date(),
                status: 'success',
            };
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : String(error);

            this.emit('node:error', executionId, node, errorMessage);

            return {
                nodeId: node.id,
                nodeName: node.name,
                inputState,
                outputState: inputState,
                startTime,
                endTime: new Date(),
                status: 'failed',
                error: errorMessage,
            };
        }
    }

    /**
     * Get the next node based on edges
     */
    private async getNextNode(
        workflow: WorkflowDefinition,
        currentNodeId: string,
        state: WorkflowState
    ): Promise<string | null> {
        const edges = workflow.edges.filter(e => e.from === currentNodeId);

        if (edges.length === 0) {
            return null;  // No outgoing edges
        }

        for (const edge of edges) {
            // Check condition if present
            if (edge.condition) {
                const shouldFollow = await edge.condition(state);
                if (!shouldFollow) continue;
            }

            // Resolve target
            if (typeof edge.to === 'function') {
                const nextId = await edge.to(state);
                if (nextId && workflow.nodes.has(nextId)) {
                    return nextId;
                }
            } else if (workflow.nodes.has(edge.to)) {
                return edge.to;
            }
        }

        return null;
    }

    /**
     * Approve a HITL checkpoint and continue execution
     */
    async approve(executionId: string): Promise<void> {
        const execution = this.executions.get(executionId);
        if (!execution || execution.status !== 'awaiting_approval') {
            throw new Error('Execution not awaiting approval');
        }

        execution.status = 'running';
        execution.updatedAt = new Date();

        const workflow = this.workflows.get(execution.workflowId)!;
        const node = workflow.nodes.get(execution.currentNode);

        if (node) {
            // Execute the approved node
            const step = await this.executeNode(executionId, node);
            execution.history.push(step);

            // Trim history
            if (execution.history.length > this.options.maxHistory) {
                execution.history = execution.history.slice(-this.options.maxHistory);
            }

            if (step.status === 'failed') {
                execution.status = 'failed';
                execution.error = step.error;
                this.emit('execution:end', execution); // Notify failure
                return;
            }

            // Update state
            execution.state = step.outputState;
            this.stateStore.set(executionId, execution.state);

            // Advance to next node
            const nextNode = await this.getNextNode(workflow, execution.currentNode, execution.state);
            if (nextNode) {
                execution.currentNode = nextNode;
                // Continue run loop
                await this.run(executionId);
            } else {
                execution.status = 'completed';
                this.emit('execution:end', execution);
            }
        } else {
            // Should not happen if valid graph
            execution.status = 'failed';
            execution.error = `Node ${execution.currentNode} not found during approval`;
        }
    }

    /**
     * Reject a HITL checkpoint and stop execution
     */
    reject(executionId: string, reason?: string): void {
        const execution = this.executions.get(executionId);
        if (!execution || execution.status !== 'awaiting_approval') {
            throw new Error('Execution not awaiting approval');
        }

        execution.status = 'failed';
        execution.error = reason || 'Rejected at checkpoint';
        execution.updatedAt = new Date();
    }

    /**
     * Pause a running execution
     */
    pause(executionId: string): void {
        const execution = this.executions.get(executionId);
        if (!execution || execution.status !== 'running') {
            throw new Error('Execution not running');
        }

        execution.status = 'paused';
        execution.updatedAt = new Date();
        this.stateStore.snapshot(executionId);
    }

    /**
     * Resume a paused execution
     */
    async resume(executionId: string): Promise<void> {
        const execution = this.executions.get(executionId);
        if (!execution || execution.status !== 'paused') {
            throw new Error('Execution not paused');
        }

        execution.status = 'running';
        execution.updatedAt = new Date();
        await this.run(executionId);
    }

    /**
     * Rollback to a previous state
     */
    rollback(executionId: string, snapshotIndex?: number): WorkflowState | undefined {
        const execution = this.executions.get(executionId);
        if (!execution) {
            throw new Error('Execution not found');
        }

        const state = this.stateStore.restore(executionId, snapshotIndex);
        if (state) {
            execution.state = state;
            execution.status = 'paused';
            execution.updatedAt = new Date();
        }
        return state;
    }

    /**
     * Get execution by ID
     */
    getExecution(executionId: string): WorkflowExecution | undefined {
        return this.executions.get(executionId);
    }

    /**
     * Get state for an execution
     */
    getState(executionId: string): WorkflowState | undefined {
        return this.stateStore.get(executionId);
    }

    /**
     * List all executions
     */
    listExecutions(): WorkflowExecution[] {
        return Array.from(this.executions.values());
    }

    /**
     * Get execution history
     */
    getHistory(executionId: string): ExecutionStep[] {
        const execution = this.executions.get(executionId);
        return execution?.history || [];
    }

    /**
     * Get state snapshots for an execution
     */
    getSnapshots(executionId: string): WorkflowState[] {
        return this.stateStore.getSnapshots(executionId);
    }
}

export default WorkflowEngine;
