/**
 * SwarmOrchestrator.ts
 * 
 * Manages horizontal parallel execution of sub-tasks by multiple AI agents.
 * 
 * Usage: Provide a massive prompt/task and let the Swarm split it, 
 * delegate to worker nodes, and aggregate the results.
 * 
 * v2.7.35: Wired decomposeGoal to the Autopilot Council for real LLM-backed
 * task decomposition, and executeTask to the Autopilot session runner for
 * live agent execution. Falls back to basic local decomposition if the
 * autopilot server is unavailable.
 */

import { v4 as uuidv4 } from 'uuid';
import { EventEmitter } from 'events';
import { DEFAULT_OPENROUTER_FREE_MODEL } from '@tormentnexus/ai';
import { MeshService, SwarmMessageType, SwarmMessage } from '../../mesh/MeshService.js';
import { MissionService, SwarmMission } from '../../services/MissionService.js';
import { RateLimiter } from './RateLimiter.js';
import { GitWorktreeManager } from '../../orchestrator/GitWorktreeManager.js';
import { resolveSwarmOrchestratorBase } from './orchestrator-base.js';

export interface SwarmToolPolicy {
    allow?: string[];
    deny?: string[];
}

export interface NormalizedSwarmToolPolicy {
    effectivePolicy?: SwarmToolPolicy;
    warnings: string[];
}

export function normalizeSwarmToolPolicy(policy?: SwarmToolPolicy): NormalizedSwarmToolPolicy {
    if (!policy) {
        return { effectivePolicy: undefined, warnings: [] };
    }

    const allow = Array.from(new Set((policy.allow || []).map(v => String(v).trim()).filter(Boolean)));
    const deny = Array.from(new Set((policy.deny || []).map(v => String(v).trim()).filter(Boolean)));
    const warnings: string[] = [];

    if (allow.length === 0 && deny.length === 0) {
        return { effectivePolicy: undefined, warnings };
    }

    const denySet = new Set(deny);
    const overlapping = allow.filter(tool => denySet.has(tool));
    const effectiveAllow = allow.filter(tool => !denySet.has(tool));

    if (overlapping.length > 0) {
        warnings.push(`toolPolicy overlap detected; deny wins for: ${overlapping.join(', ')}`);
    }

    const effectivePolicy: SwarmToolPolicy = {};
    if (effectiveAllow.length > 0) effectivePolicy.allow = effectiveAllow;
    if (deny.length > 0) effectivePolicy.deny = deny;

    if (!effectivePolicy.allow && !effectivePolicy.deny) {
        return { effectivePolicy: undefined, warnings };
    }

    return { effectivePolicy, warnings };
}

export interface SwarmToolDenyEvent {
    tool: string;
    reason: string;
    timestamp: number;
}

export interface SwarmTask {
    id: string;
    description: string;
    assignedModel?: string;
    result?: string;
    status: 'pending' | 'running' | 'completed' | 'failed' | 'pending_approval' | 'awaiting_subtask' | 'healing' | 'throttled' | 'verifying';
    error?: string;
    /** Phase 83: Number of times this task has been retried */
    retryCount?: number;
    /** Phase 84: 1-5, higher is better */
    priority: number;
    /** Autopilot session ID if delegated to the Council */
    sessionId?: string;
    /** Phase 82: ID of a mission spawned to solve this task */
    subMissionId?: string;
    /** Phase 85: Resource usage tracking */
    usage?: { tokens: number; estimatedMemory: number };

    // Phase 88: Verification
    verifiedBy?: string;
    slashed?: boolean;

    // Phase 91: MCP Tool integration
    tools?: string[];

    // Phase 96: Mission-level tool permission boundaries
    toolPolicy?: SwarmToolPolicy;
    deniedToolEvents?: SwarmToolDenyEvent[];

    // Phase 94: Sub-Agent Task Routing
    requirements?: string[];

    // Phase 95: Git Worktree Isolation
    worktreePath?: string;

    // Phase 125: Red Team Debate Agent
    isRedTeam?: boolean;
}

export interface SwarmConfig {
    maxConcurrency?: number;
    defaultModel?: string;
    timeoutMs?: number;
    opencodeUrl?: string;
    /** Phase 83: Max retries for failed tasks before calling the Healer */
    maxRetries?: number;
    /** Phase 85: Token/Resource limits */
    maxTokensPerMission?: number;
    maxTokensPerTask?: number;
    /** Phase 86: Adaptive Rate Limiting */
    rpmLimit?: number;
    tpmLimit?: number;
    /** Phase 95: Optional GitWorktreeManager for context isolation */
    gitWorktreeManager?: GitWorktreeManager;
}

export class SwarmOrchestrator extends EventEmitter {
    private tasks: Map<string, SwarmTask> = new Map();
    private config: Required<Pick<SwarmConfig, 'maxConcurrency' | 'defaultModel' | 'timeoutMs' | 'maxRetries' | 'maxTokensPerMission' | 'maxTokensPerTask' | 'rpmLimit' | 'tpmLimit'>>;
    private opencodeUrl: string | null;
    private mesh: MeshService;
    private rateLimiter: RateLimiter;
    private missionService?: MissionService;
    private healerService?: any; // Phase 83
    private currentMissionId?: string;
    private gitWorktreeManager?: GitWorktreeManager; // Phase 95

    // Track tasks waiting for manual approval
    private approvalResolvers: Map<string, (approved: boolean) => void> = new Map();

    constructor(config: SwarmConfig = {}, missionService?: MissionService, healerService?: any) {
        super();
        this.config = {
            maxConcurrency: config.maxConcurrency || 5,
            defaultModel: config.defaultModel || DEFAULT_OPENROUTER_FREE_MODEL,
            timeoutMs: config.timeoutMs || 300000,
            maxRetries: config.maxRetries || 3,
            maxTokensPerMission: config.maxTokensPerMission || 1000000,
            maxTokensPerTask: config.maxTokensPerTask || 100000,
            rpmLimit: config.rpmLimit || 50,
            tpmLimit: config.tpmLimit || 40000
        };
        this.rateLimiter = new RateLimiter(this.config.rpmLimit, this.config.tpmLimit);
        this.opencodeUrl = resolveSwarmOrchestratorBase(config.opencodeUrl);
        this.mesh = new MeshService();
        this.missionService = missionService;
        this.healerService = healerService;
        this.gitWorktreeManager = config.gitWorktreeManager; // Phase 95
    }

    /**
     * Resumes execution for an existing mission.
     */
    public async resumeMission(missionId: string): Promise<void> {
        if (!this.missionService) throw new Error('MissionService not available');
        const mission = this.missionService.getMission(missionId);
        if (!mission) throw new Error(`Mission ${missionId} not found`);

        this.currentMissionId = missionId;
        mission.tasks.forEach(t => this.tasks.set(t.id, t));

        console.log(`[Swarm] Resuming mission ${missionId} with ${this.tasks.size} tasks.`);
        await this.executeSwarm();
    }

    /**
     * Manually approve or reject a task.
     */
    public approveTask(taskId: string, approved: boolean): boolean {
        const resolver = this.approvalResolvers.get(taskId);
        if (resolver) {
            resolver(approved);
            this.approvalResolvers.delete(taskId);
            return true;
        }
        return false;
    }

    /**
     * Explodes a single task into its own sub-mission. (Phase 82)
     */
    public async decomposeTask(taskId: string): Promise<SwarmMission | null> {
        if (!this.missionService) return null;
        const task = this.tasks.get(taskId);
        if (!task || !this.currentMissionId) return null;

        console.log(`[Swarm] Recursively decomposing task: ${task.description}`);

        // Use the council to decompose this specific task description
        const subTasks = await this.decomposeGoal(`SUB-TASK: ${task.description}`, task.tools, task.toolPolicy);

        // Link the sub-mission to the parent
        const subMission = this.missionService.createMission(
            task.description,
            subTasks,
            this.currentMissionId
        );

        // Update the parent task state
        task.status = 'awaiting_subtask';
        task.subMissionId = subMission.id;
        this.missionService.updateMissionTask(this.currentMissionId, taskId, {
            status: 'awaiting_subtask',
            subMissionId: subMission.id
        });

        this.emit('task:decomposed', { taskId, subMissionId: subMission.id });
        return subMission;
    }

    /**
     * Decompose a master prompt into actionable sub-tasks.
     * 
     * Uses the Autopilot Council's debate endpoint to generate a real
     * task breakdown via multi-model consensus. Falls back to a simple
     * three-phase split if the council is unreachable.
     */
    public async decomposeGoal(masterPrompt: string, tools?: string[], toolPolicy?: SwarmToolPolicy): Promise<SwarmTask[]> {
        console.log(`[Swarm] Decomposing: ${masterPrompt}`);
        const { effectivePolicy, warnings } = normalizeSwarmToolPolicy(toolPolicy);

        if (warnings.length > 0) {
            for (const warning of warnings) {
                console.warn(`[Swarm] Tool policy normalization warning: ${warning}`);
            }
        }

        let subTasks: SwarmTask[] = [];

        try {
            if (!this.opencodeUrl) {
                throw new Error('No TormentNexus Orchestrator base configured.');
            }
            // Use the Autopilot Council to decompose the goal via multi-model debate.
            const res = await fetch(`${this.opencodeUrl}/api/council/debate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    id: uuidv4(),
                    description: `Decompose this goal into 3-7 independent, parallel-executable sub-tasks. Return ONLY a JSON array of objects with {description: string}. Goal: ${masterPrompt}`,
                    context: `You are a task decomposition architect. Break the goal into concrete, actionable sub-tasks that can be executed independently and in parallel by different AI models. Each task should be self-contained.`,
                    files: []
                }),
                signal: AbortSignal.timeout(this.config.timeoutMs)
            });

            if (res.ok) {
                const debateResult = await res.json();
                const consensusText = debateResult?.consensus?.finalAnswer
                    || debateResult?.consensus?.text
                    || debateResult?.result
                    || '';

                const parsed = this.extractTasksFromLLMResponse(consensusText, masterPrompt, tools, effectivePolicy);
                if (parsed.length > 0) {
                    subTasks = parsed;
                    console.log(`[Swarm] Council decomposed goal into ${subTasks.length} real tasks`);
                }
            }
        } catch (err: any) {
            console.warn(`[Swarm] Council unavailable for decomposition (${err.message}), using local fallback`);
        }

        // Fallback: basic three-phase decomposition if council didn't return tasks
        if (subTasks.length === 0) {
            subTasks = [
                { id: uuidv4(), description: `Analyze and plan: ${masterPrompt}`, status: 'pending', priority: 5, tools: tools || [], toolPolicy: effectivePolicy },
                { id: uuidv4(), description: `Implement core logic for: ${masterPrompt}`, status: 'pending', priority: 3, tools: tools || [], toolPolicy: effectivePolicy },
                { id: uuidv4(), description: `Verify and test: ${masterPrompt}`, status: 'pending', priority: 4, tools: tools || [], toolPolicy: effectivePolicy }
            ];
            console.log(`[Swarm] Using fallback decomposition (${subTasks.length} tasks)`);
        }

        // Phase 125: Red Team Debate Agent
        // Inject a high-priority adversarial critique task to stress-test the generated plan
        if (subTasks.length > 0) {
            const planSummary = subTasks.map((t, i) => `${i + 1}. ${t.description}`).join('\n');
            subTasks.push({
                id: uuidv4(),
                description: `Critique and stress-test this proposed mission plan for logical gaps, edge cases, and security flaws:\n\n${planSummary}`,
                status: 'pending',
                priority: 5, // Top priority to ensure early critique availability
                tools: [],   // Critique doesn't need execution tools
                toolPolicy: effectivePolicy,
                requirements: ['researcher'],
                isRedTeam: true
            });
            console.log(`[Swarm] Injected Red Team critique task against the proposed plan.`);
        }

        subTasks.forEach(t => this.tasks.set(t.id, t));

        // Create a persistent mission if service is available
        if (this.missionService) {
            const mission = this.missionService.createMission(masterPrompt, subTasks);
            this.currentMissionId = mission.id;

            if (effectivePolicy || warnings.length > 0) {
                this.missionService.updateMissionContext(mission.id, {
                    _swarmPolicy: {
                        effectiveToolPolicy: effectivePolicy,
                        policyWarnings: warnings,
                        capturedAt: Date.now()
                    }
                });
            }

            console.log(`[Swarm] Persistent mission created: ${this.currentMissionId}`);
        }

        return subTasks;
    }

    /**
     * Attempts to extract a JSON array of tasks from an LLM response string.
     */
    private extractTasksFromLLMResponse(text: string, fallbackPrompt: string, tools?: string[], toolPolicy?: SwarmToolPolicy): SwarmTask[] {
        if (!text || typeof text !== 'string') return [];
        const cleaned = text.replace(/```json?\n?/g, '').replace(/```/g, '').trim();
        const arrayMatch = cleaned.match(/\[[\s\S]*\]/);
        if (!arrayMatch) return [];

        try {
            const parsed = JSON.parse(arrayMatch[0]);
            if (!Array.isArray(parsed)) return [];

            return parsed
                .filter((item: any) => item && (item.description || item.task || item.name))
                .map((item: any) => {
                    const desc = (item.description || item.task || item.name).toLowerCase();
                    const reqs: string[] = [];

                    // Phase 94: Sub-Agent Task Classification
                    if (desc.includes('research') || desc.includes('search') || desc.includes('find') || desc.includes('analyze')) {
                        reqs.push('researcher');
                    }
                    if (desc.includes('code') || desc.includes('implement') || desc.includes('refactor') || desc.includes('fix')) {
                        reqs.push('coder');
                    }

                    return {
                        id: uuidv4(),
                        description: item.description || item.task || item.name,
                        status: 'pending' as const,
                        priority: item.priority || 3,
                        tools: tools || [],
                        toolPolicy,
                        requirements: reqs // Attach to task object
                    };
                });
        } catch {
            return [];
        }
    }

    /**
     * Phase 89: Dynamic Resource Allocation
     * Scans global mission state to throttle maxConcurrency if higher-priority
     * tasks are currently saturating the network.
     */
    private calculatePriorityQuota(taskPriority: number): number {
        if (!this.missionService) return this.config.maxConcurrency;

        const allMissions = this.missionService.getAllMissions().filter(m => m.status === 'active');
        let higherPriorityRunning = 0;

        for (const m of allMissions) {
            if (m.id === this.currentMissionId) continue;

            for (const t of m.tasks) {
                if (t.status === 'running' || t.status === 'verifying') {
                    if (t.priority > taskPriority) {
                        higherPriorityRunning++;
                    }
                }
            }
        }

        // Throttling logic: If there are higher priority tasks running globally,
        // we heavily throttle this batch to leave mesh capacity open.
        if (higherPriorityRunning > 0) {
            console.log(`[SwarmOrchestrator] 🚦 Dynamic Quota: Throttling P${taskPriority} tasks because ${higherPriorityRunning} higher-priority tasks are active globally.`);
            // Throttle to 20% of max, but at least 1
            return Math.max(1, Math.floor(this.config.maxConcurrency * 0.2));
        }

        return this.config.maxConcurrency;
    }

    /**
     * Executes the swarm loop until all tasks are complete.
     */
    public async executeSwarm(): Promise<Map<string, SwarmTask>> {
        this.emit('swarm:started', { totalTasks: this.tasks.size, missionId: this.currentMissionId });

        // Phase 84: Sort tasks by priority descending before execution
        const pending = Array.from(this.tasks.values())
            .filter(t => t.status === 'pending')
            .sort((a, b) => b.priority - a.priority);

        // Phase 89: Dynamic Resource Allocation
        // Evaluate concurrency on a per-batch basis depending on the pending queue
        while (pending.length > 0) {
            const currentTask = pending[0];
            const dynamicConcurrency = this.calculatePriorityQuota(currentTask.priority);

            const batch = pending.splice(0, dynamicConcurrency);
            const promises = batch.map(task => this.executeTask(task));
            await Promise.allSettled(promises);
        }

        this.emit('swarm:completed', { results: Array.from(this.tasks.values()), missionId: this.currentMissionId });
        return this.tasks;
    }

    /**
     * Execute a single task and persist its status.
     * Includes HITL approval for sensitive tasks.
     */
    private async executeTask(task: SwarmTask): Promise<void> {
        // HITL Check: if task description contains sensitive keywords, wait for approval
        const sensitiveKeywords = ['delete', 'remove', 'format', 'overwrite', 'deploy', 'publish'];
        const needsApproval = sensitiveKeywords.some(kw => task.description.toLowerCase().includes(kw));

        if (needsApproval) {
            task.status = 'pending_approval';
            if (this.missionService && this.currentMissionId) {
                this.missionService.updateMissionTask(this.currentMissionId, task.id, { status: 'pending_approval' });
            }
            this.emit('task:awaiting_approval', task);

            const approved = await new Promise<boolean>((resolve) => {
                this.approvalResolvers.set(task.id, resolve);
            });

            if (!approved) {
                task.status = 'failed';
                task.result = 'Task rejected by user.';
                if (this.missionService && this.currentMissionId) {
                    this.missionService.updateMissionTask(this.currentMissionId, task.id, { status: 'failed', result: task.result });
                }
                this.emit('task:rejected', task);
                return;
            }
        }

        // Phase 82: If the task has a sub-mission, execution is handled by the sub-mission lifecycle.
        // For now, we simulate completion once the sub-mission is created.
        if (task.status === 'awaiting_subtask') {
            console.log(`[Swarm] Task ${task.id} is awaiting sub-mission ${task.subMissionId}. Skipping direct execution.`);
            return;
        }

        // Phase 86: Adaptive Rate Limiting
        const estimatedTokens = 1000;
        while (!this.rateLimiter.tryAcquire(estimatedTokens)) {
            if (task.status !== 'throttled') {
                task.status = 'throttled';
                if (this.missionService && this.currentMissionId) {
                    this.missionService.updateMissionTask(this.currentMissionId, task.id, { status: 'throttled' });
                }
                this.emit('task:throttled', task);
            }
            // Wait before checking capacity again
            await new Promise(resolve => setTimeout(resolve, Math.max(1000, this.rateLimiter.getBackoffRemainingMs())));
        }

        task.status = 'running';
        task.assignedModel = this.config.defaultModel;

        // Update persistence
        if (this.missionService && this.currentMissionId) {
            this.missionService.updateMissionTask(this.currentMissionId, task.id, { status: 'running', assignedModel: task.assignedModel });
        }

        this.emit('task:started', task);

        try {
            // Phase 95: Git Worktree Isolation for coding tasks
            // Spin up an isolated worktree so parallel agents don't conflict on files
            const isCodingTask = (task.requirements || []).includes('coder');
            if (isCodingTask && this.gitWorktreeManager) {
                try {
                    const wtPath = await this.gitWorktreeManager.createTaskEnvironment(task.id);
                    task.worktreePath = wtPath;
                    console.log(`[SwarmOrchestrator] 🌳 Worktree created for task ${task.id.slice(0, 8)} at ${wtPath}`);
                } catch (wtErr: any) {
                    console.warn(`[SwarmOrchestrator] Worktree creation failed (${wtErr.message}), proceeding without isolation.`);
                }
            }

            // Phase 90: Read global mission context
            let globalContext = {};
            if (this.missionService && this.currentMissionId) {
                const mission = this.missionService.getMission(this.currentMissionId);
                if (mission && mission.context) {
                    globalContext = mission.context;
                }
            }

            const offerPayload = {
                task: task.description,
                requirements: task.requirements || [], // Phase 94: Pass requirements to the Mesh
                tools: task.tools, // Phase 91/92: Add tools array to payload
                toolPolicy: task.toolPolicy, // Phase 96: Permission boundaries for delegated task tools
                missionId: this.currentMissionId,
                originalTaskId: task.id,
                context: Object.keys(globalContext).length > 0 ? globalContext : undefined, // Phase 90
                worktreePath: task.worktreePath // Phase 95: Pass isolated worktree path to agent
            };

            console.log(`[SwarmOrchestrator] 🌐 Broadcasting TASK_OFFER to Mesh Network (Attempt ${task.retryCount || 1}/${this.config.maxRetries}): "${task.description.slice(0, 30)}..."`);
            this.mesh.broadcast(SwarmMessageType.TASK_OFFER, offerPayload);

            // Phase 92: Bidding Window (Collect TASK_BIDs)
            const BID_WINDOW_MS = 1000;
            const bids: { sender: string; load: number }[] = [];

            const bidPromise = new Promise<string | null>((resolve) => {
                const bidTimeout = setTimeout(() => {
                    this.mesh.off('message', bidHandler);
                    if (bids.length > 0) {
                        // Sort by lowest load
                        bids.sort((a, b) => a.load - b.load);
                        resolve(bids[0].sender);
                    } else {
                        resolve(null);
                    }
                }, BID_WINDOW_MS);

                const bidHandler = (msg: SwarmMessage) => {
                    if (msg.type === SwarmMessageType.TASK_BID) {
                        const bidPayload = msg.payload as any;
                        if (bidPayload.originalTaskId === task.id && msg.target === this.mesh.nodeId) {
                            bids.push({ sender: msg.sender, load: bidPayload.load || 0 });
                        }
                    }
                };
                this.mesh.on('message', bidHandler);
            });

            const winnerNodeId = await bidPromise;
            if (!winnerNodeId) {
                throw new Error('No agents bid on the task offer.');
            }

            console.log(`[SwarmOrchestrator] 🎯 Assigning task ${task.id.slice(0, 8)} to node ${winnerNodeId.slice(0, 8)}...`);
            this.mesh.sendDirect(winnerNodeId, SwarmMessageType.TASK_ASSIGN, offerPayload);

            // Wait for a TASK_RESULT from the winning agent
            const resultPromise = new Promise<{ result?: any, error?: string }>((resolve, reject) => {
                const timeout = setTimeout(() => {
                    this.mesh.off('message', handler);
                    reject(new Error('Mesh execution timed out'));
                }, this.config.timeoutMs);

                const handler = (msg: SwarmMessage) => {
                    if (msg.type === SwarmMessageType.TASK_RESULT) {
                        const payload = msg.payload as any;
                        if (payload.originalTaskId === task.id && msg.target === this.mesh.nodeId) {
                            clearTimeout(timeout);
                            this.mesh.off('message', handler);
                            resolve({ result: payload.result, error: payload.error });
                        }
                    }
                };
                this.mesh.on('message', handler);
            });

            const meshResult = await resultPromise;
            if (meshResult.error) throw new Error(meshResult.error);

            // Phase 90: Extract and sync context updates
            if (meshResult.result && typeof meshResult.result === 'object') {
                if (meshResult.result._contextUpdate && this.missionService && this.currentMissionId) {
                    console.log(`[SwarmOrchestrator] 🧠 Received Context Update from task ${task.id}:`, meshResult.result._contextUpdate);
                    this.missionService.updateMissionContext(this.currentMissionId, meshResult.result._contextUpdate);
                    // Optionally strip it so it doesn't clutter the task result
                    delete meshResult.result._contextUpdate;
                }

                // Phase 96: Capture denied-tool telemetry from worker and persist on task/mission history
                const telemetry = (meshResult.result as any)._swarmTelemetry;
                const deniedTools = telemetry?.deniedTools;
                if (Array.isArray(deniedTools) && deniedTools.length > 0) {
                    const normalizedDeniedTools = deniedTools
                        .filter((item: any) => item && typeof item.tool === 'string' && typeof item.reason === 'string')
                        .map((item: any) => ({
                            tool: item.tool,
                            reason: item.reason,
                            timestamp: typeof item.timestamp === 'number' ? item.timestamp : Date.now()
                        }));

                    if (normalizedDeniedTools.length > 0) {
                        task.deniedToolEvents = [...(task.deniedToolEvents || []), ...normalizedDeniedTools];

                        this.emit('task:tool_denied', {
                            taskId: task.id,
                            missionId: this.currentMissionId,
                            deniedTools: normalizedDeniedTools
                        });

                        if (global.mcpServerInstance?.eventBus) {
                            global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'SwarmOrchestrator', {
                                id: uuidv4(),
                                type: 'SWARM_TOOL_DENIED',
                                sender: 'SwarmOrchestrator',
                                timestamp: Date.now(),
                                payload: {
                                    missionId: this.currentMissionId,
                                    taskId: task.id,
                                    deniedTools: normalizedDeniedTools
                                }
                            });
                        }
                    }

                    delete (meshResult.result as any)._swarmTelemetry;
                }
            }

            task.result = typeof meshResult.result === 'string'
                ? meshResult.result
                : JSON.stringify(meshResult.result, null, 2);

            // Phase 88: Verification Flow
            task.status = 'verifying';
            if (this.missionService && this.currentMissionId) {
                this.missionService.updateMissionTask(this.currentMissionId, task.id, {
                    status: task.status,
                    result: task.result
                });
            }
            this.emit('task:verifying', task);

            console.log(`[SwarmOrchestrator] 🔍 Broadcasting VERIFY_OFFER for task: "${task.description.slice(0, 30)}..."`);
            this.mesh.broadcast(SwarmMessageType.VERIFY_OFFER, {
                task: task.description,
                result: task.result,
                originalTaskId: task.id,
                missionId: this.currentMissionId
            });

            // Wait for a VERIFY_RESULT
            const verifyPromise = new Promise<{ approved: boolean, reason: string, verifierId: string }>((resolve, reject) => {
                const timeout = setTimeout(() => {
                    this.mesh.off('message', handler);
                    // If no one verifies, accept it to avoid deadlock
                    resolve({ approved: true, reason: 'Verification timed out, auto-approving.', verifierId: 'system' });
                }, this.config.timeoutMs);

                const handler = (msg: SwarmMessage) => {
                    if (msg.type === SwarmMessageType.VERIFY_RESULT) {
                        const payload = msg.payload as any;
                        if (payload.originalTaskId === task.id && msg.target === this.mesh.nodeId) {
                            clearTimeout(timeout);
                            this.mesh.off('message', handler);
                            resolve({ approved: payload.approved, reason: payload.reason, verifierId: msg.sender });
                        }
                    }
                };
                this.mesh.on('message', handler);
            });

            const verifyResult = await verifyPromise;

            if (!verifyResult.approved) {
                console.warn(`[SwarmOrchestrator] ❌ Verification failed for task ${task.id} by ${verifyResult.verifierId}: ${verifyResult.reason}`);
                task.slashed = true;
                // Treat verification failure as a task execution error to trigger retries
                throw new Error(`Verification Rejected: ${verifyResult.reason}`);
            }

            console.log(`[SwarmOrchestrator] ✅ Task ${task.id} verified by ${verifyResult.verifierId}`);
            task.status = 'completed';
            task.verifiedBy = verifyResult.verifierId;

            // Phase 85: Resource Tracking
            const resultPayload = meshResult.result as any;
            const tokensUsed = resultPayload?.usage?.total_tokens || 1000;
            const memoryUsed = resultPayload?.usage?.memory_bytes || 1024 * 1024;
            task.usage = { tokens: tokensUsed, estimatedMemory: memoryUsed };

            // Update persistence
            if (this.missionService && this.currentMissionId) {
                const mission = this.missionService.getMission(this.currentMissionId);
                if (mission) {
                    mission.usage.tokens += tokensUsed;
                    mission.usage.estimatedMemory = Math.max(mission.usage.estimatedMemory, memoryUsed);

                    // Check limits
                    if (this.config.maxTokensPerMission && mission.usage.tokens > this.config.maxTokensPerMission) {
                        task.status = 'failed';
                        task.error = `RESOURCE_LIMIT_EXCEEDED: Mission token limit (${this.config.maxTokensPerMission}) reached.`;
                    }
                }

                this.missionService.updateMissionTask(this.currentMissionId, task.id, {
                    status: task.status,
                    result: task.result,
                    error: task.error,
                    usage: task.usage,
                    deniedToolEvents: task.deniedToolEvents,
                    slashed: task.slashed,
                    verifiedBy: task.verifiedBy,
                    isRedTeam: task.isRedTeam
                });
            }

            this.emit('task:completed', task);
            return;

        } catch (err: any) {
            console.warn(`[SwarmOrchestrator] Mesh delegation failed for "${task.description}": ${err.message}.`);

            // Phase 86: Rate Limit detection
            if (err.message && (err.message.includes('429') || err.message.toLowerCase().includes('rate limit'))) {
                console.warn(`[Swarm] Provider Rate Limit hit! Triggering backoff...`);
                this.rateLimiter.triggerBackoff(30); // 30 second global backoff
            }

            // Phase 83: Retry logic
            const currentRetry = task.retryCount || 0;
            if (currentRetry < this.config.maxRetries) {
                task.retryCount = currentRetry + 1;
                console.log(`[Swarm] Retrying task ${task.id} (${task.retryCount}/${this.config.maxRetries})...`);
                if (this.missionService && this.currentMissionId) {
                    this.missionService.updateMissionTask(this.currentMissionId, task.id, { retryCount: task.retryCount });
                }
                // Wait a bit before retrying (exponential backoff simulated)
                await new Promise(r => setTimeout(r, 2000 * task.retryCount!));
                return this.executeTask(task);
            }

            console.log(`[Swarm] Max retries reached for task ${task.id}. Triggering Healer...`);
            task.status = 'healing';
            if (this.missionService && this.currentMissionId) {
                this.missionService.updateMissionTask(this.currentMissionId, task.id, { status: 'healing' });
            }
            this.emit('task:healing', task);

            // Phase 83: Healer Integration
            if (this.healerService) {
                const healed = await this.healerService.heal(err, `Swarm Task: ${task.description}`);
                if (healed) {
                    console.log(`[Swarm] Healer successfully repaired the environment for task ${task.id}. Retrying one last time...`);
                    return this.executeTask(task);
                }
            }

            // If healer fails or is unavailable, mark as failed
            console.log(`[Worker - local] Completing with failure after exhausted retries: ${task.description}`);
            task.result = `[Failure] Task "${task.description}" failed after ${this.config.maxRetries} retries. Error: ${err.message}`;
            task.status = 'failed';
        }

        // Phase 95: Cleanup worktree after task execution (success or failure)
        if (task.worktreePath && this.gitWorktreeManager) {
            try {
                await this.gitWorktreeManager.cleanupTaskEnvironment(task.id);
                console.log(`[SwarmOrchestrator] 🧹 Worktree cleaned up for task ${task.id.slice(0, 8)}`);
            } catch (cleanupErr: any) {
                console.warn(`[SwarmOrchestrator] Worktree cleanup failed for task ${task.id}: ${cleanupErr.message}`);
            }
        }

        // Final persistence update
        if (this.missionService && this.currentMissionId) {
            this.missionService.updateMissionTask(this.currentMissionId, task.id, {
                status: task.status,
                result: task.result,
                deniedToolEvents: task.deniedToolEvents
            });
        }

        this.emit('task:completed', task);
    }

    public destroy() {
        this.mesh.destroy();
    }
}
