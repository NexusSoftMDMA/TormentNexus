/**
 * MissionService.ts
 *
 * Handles persistent storage and recovery of Swarm Missions.
 * Missions represent high-level goals decomposed into parallel tasks.
 *
 * v2.7.40: Initial implementation for Phase 80.
 */

import fs from 'fs';
import path from 'path';
import crypto from 'crypto';
import { EventEmitter } from 'events';
import { SwarmTask } from '../agents/swarm/SwarmOrchestrator.js';

export interface SwarmUsage {
    tokens: number;
    estimatedMemory: number; // in bytes
}

export interface SwarmMission {
    id: string;
    goal: string;
    status: 'active' | 'completed' | 'failed' | 'paused';
    tasks: SwarmTask[];
    parentId?: string; // Phase 82: Parent mission ID
    priority: number;  // Phase 84: 1-5, higher is better
    usage: SwarmUsage; // Phase 85: Resource Tracking
    context: Record<string, any>; // Phase 90: Shared Context
    createdAt: string;
    updatedAt: string;
}

export class MissionService extends EventEmitter {
    private historyPath: string;
    private missions: Map<string, SwarmMission> = new Map();

    constructor(rootDir: string) {
        super();
        this.historyPath = path.join(rootDir, '.tormentnexus', 'mission_history.json');
        this.loadMissions();
    }

    private loadMissions() {
        try {
            if (fs.existsSync(this.historyPath)) {
                const raw = fs.readFileSync(this.historyPath, 'utf-8');
                const data = JSON.parse(raw) as SwarmMission[];
                data.forEach(m => this.missions.set(m.id, m));
                console.log(`[MissionService] 📂 Loaded ${this.missions.size} missions from history.`);
            }
        } catch (e) {
            console.warn('[MissionService] Failed to load mission history:', e);
        }
    }

    private saveMissions() {
        try {
            const dir = path.dirname(this.historyPath);
            if (!fs.existsSync(dir)) {
                fs.mkdirSync(dir, { recursive: true });
            }
            const data = Array.from(this.missions.values());
            fs.writeFileSync(this.historyPath, JSON.stringify(data, null, 2));
        } catch (e) {
            console.warn('[MissionService] Failed to save mission history:', e);
        }
    }

    /**
     * Creates a new mission and persists it.
     */
    public createMission(goal: string, tasks: SwarmTask[], parentId?: string, priority: number = 3): SwarmMission {
        const mission: SwarmMission = {
            id: crypto.randomUUID(),
            goal,
            status: 'active',
            tasks,
            parentId,
            priority,
            usage: { tokens: 0, estimatedMemory: 0 },
            context: {}, // Phase 90
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
        };
        this.missions.set(mission.id, mission);
        this.saveMissions();
        this.emit('mission:created', mission);

        // Bubble up to global event bus for Telemetry (Phase 80)
        if (global.mcpServerInstance?.eventBus) {
            global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'MissionService', {
                id: crypto.randomUUID(),
                type: 'MISSION_CREATED',
                sender: 'MissionService',
                timestamp: Date.now(),
                payload: { missionId: mission.id, goal: mission.goal }
            });
        }
        return mission;
    }

    /**
     * Updates a specific task within a mission.
     * Automatically updates mission status if all tasks are complete/failed.
     */
    public updateMissionTask(missionId: string, taskId: string, updates: Partial<SwarmTask>): SwarmMission | null {
        const mission = this.missions.get(missionId);
        if (!mission) return null;

        const taskIndex = mission.tasks.findIndex(t => t.id === taskId);
        if (taskIndex === -1) return null;

        mission.tasks[taskIndex] = { ...mission.tasks[taskIndex], ...updates };
        mission.updatedAt = new Date().toISOString();

        if (updates.status === 'running') mission.status = 'active';
        if (mission.tasks.some(t => t.status === 'pending_approval')) mission.status = 'paused';
        if (mission.tasks.every(t => t.status === 'completed')) mission.status = 'completed';
        if (mission.tasks.some(t => t.status === 'failed')) mission.status = 'failed';

        this.saveMissions();
        this.emit('mission:updated', mission);

        // Bubble up to global event bus for Telemetry (Phase 80)
        if (global.mcpServerInstance?.eventBus) {
            global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'MissionService', {
                id: crypto.randomUUID(),
                type: 'MISSION_UPDATED',
                sender: 'MissionService',
                timestamp: Date.now(),
                payload: { missionId: mission.id, status: mission.status, taskId, taskStatus: updates.status }
            });
        }

        return mission;
    }

    /**
     * Phase 90: Updates the global shared context of a mission.
     */
    public updateMissionContext(missionId: string, diff: Record<string, any>): SwarmMission | null {
        const mission = this.missions.get(missionId);
        if (!mission) return null;

        mission.context = { ...mission.context, ...diff };
        mission.updatedAt = new Date().toISOString();

        this.saveMissions();
        this.emit('mission:context_updated', mission);

        if (global.mcpServerInstance?.eventBus) {
            global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'MissionService', {
                id: crypto.randomUUID(),
                type: 'MISSION_CONTEXT_UPDATED',
                sender: 'MissionService',
                timestamp: Date.now(),
                payload: { missionId: mission.id, context: mission.context }
            });
        }

        return mission;
    }

    /**
     * Resumes an existing mission by resetting failed/pending tasks and setting status to active.
     */
    public resumeMission(missionId: string): SwarmMission | null {
        const mission = this.missions.get(missionId);
        if (!mission) return null;

        mission.status = 'active';
        mission.updatedAt = new Date().toISOString();

        // Reset non-completed tasks to pending
        mission.tasks = mission.tasks.map(t =>
            t.status !== 'completed' ? { ...t, status: 'pending' } : t
        );

        this.saveMissions();
        this.emit('mission:resumed', mission);

        if (global.mcpServerInstance?.eventBus) {
            global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'MissionService', {
                id: crypto.randomUUID(),
                type: 'MISSION_RESUMED',
                sender: 'MissionService',
                timestamp: Date.now(),
                payload: { missionId: mission.id }
            });
        }

        return mission;
    }

    public getMission(id: string): SwarmMission | undefined {
        return this.missions.get(id);
    }

    public getAllMissions(): SwarmMission[] {
        return Array.from(this.missions.values()).sort((a, b) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
    }

    /**
     * Returns active missions that might need recovery.
     */
    public getActiveMissions(): SwarmMission[] {
        return Array.from(this.missions.values()).filter(m => m.status === 'active' || m.status === 'paused');
    }
}
