export type DirectorPlanStatus = 'IN_PROGRESS' | 'IDLE';
export type DirectorStepStatus = 'RUNNING' | 'DONE';

export interface DirectorPlanStep {
    id: number;
    action: string;
    status: DirectorStepStatus;
    result: string;
}

export interface DirectorPlanView {
    goal: string;
    status: DirectorPlanStatus;
    steps: DirectorPlanStep[];
}

export type DirectorAutonomyLevel = 'high' | 'medium' | 'low' | 'unknown';

function isObject(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

export function normalizeDirectorPlan(configPayload: unknown, taskStatusPayload: unknown): DirectorPlanView {
    const defaultGoal = 'Defining Mission...';

    const goal = isObject(configPayload) && typeof configPayload.defaultTopic === 'string' && configPayload.defaultTopic.trim().length > 0
        ? configPayload.defaultTopic.trim()
        : defaultGoal;

    const normalizedTaskStatus = isObject(taskStatusPayload) ? taskStatusPayload : {};
    const rawStatus = normalizedTaskStatus.status;
    const isInProgress = rawStatus === 'processing' || rawStatus === 'busy';

    const taskId = typeof normalizedTaskStatus.taskId === 'string' && normalizedTaskStatus.taskId.trim().length > 0
        ? normalizedTaskStatus.taskId.trim()
        : null;

    const progress = typeof normalizedTaskStatus.progress === 'number' && Number.isFinite(normalizedTaskStatus.progress)
        ? normalizedTaskStatus.progress
        : 0;

    const steps: DirectorPlanStep[] = taskId
        ? [{ id: 1, action: taskId, status: 'RUNNING', result: `Progress: ${progress}%` }]
        : [];

    return {
        goal,
        status: isInProgress ? 'IN_PROGRESS' : 'IDLE',
        steps,
    };
}

export function normalizeDirectorAutonomyLevel(payload: unknown): DirectorAutonomyLevel {
    if (payload === 'high' || payload === 'medium' || payload === 'low') {
        return payload;
    }

    return 'unknown';
}
