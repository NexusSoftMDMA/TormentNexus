export type UserPromptRole = 'goal' | 'objective' | 'prompt';

export interface UserPromptInput {
    content: string;
    role?: UserPromptRole;
    promptNumber?: number;
    sessionId?: string;
    activeGoal?: string | null;
    lastObjective?: string | null;
}

export interface StructuredUserPrompt {
    role: UserPromptRole;
    content: string;
    promptNumber: number;
    sessionId?: string;
    activeGoal?: string | null;
    lastObjective?: string | null;
    contentHash: string;
    recordedAt: number;
}

function sanitizeLine(value: string, maxLength: number): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function safeContent(value: string): string {
    return sanitizeLine(value, 400);
}

export function buildStructuredUserPrompt(input: UserPromptInput & { recordedAt: number; contentHash: string }): StructuredUserPrompt {
    return {
        role: input.role ?? 'prompt',
        content: safeContent(input.content),
        promptNumber: typeof input.promptNumber === 'number' && input.promptNumber > 0 ? Math.floor(input.promptNumber) : 1,
        sessionId: typeof input.sessionId === 'string' && input.sessionId.trim() ? input.sessionId.trim() : undefined,
        activeGoal: typeof input.activeGoal === 'string' ? sanitizeLine(input.activeGoal, 180) : input.activeGoal,
        lastObjective: typeof input.lastObjective === 'string' ? sanitizeLine(input.lastObjective, 180) : input.lastObjective,
        contentHash: input.contentHash,
        recordedAt: input.recordedAt,
    };
}

export function buildUserPromptContent(prompt: StructuredUserPrompt): string {
    const heading = prompt.role === 'goal'
        ? 'Captured goal prompt'
        : prompt.role === 'objective'
            ? 'Captured objective prompt'
            : 'Captured user prompt';

    const contextLines = [
        prompt.activeGoal ? `Active goal: ${prompt.activeGoal}` : null,
        prompt.lastObjective ? `Last objective: ${prompt.lastObjective}` : null,
    ].filter((line): line is string => Boolean(line));

    return [
        `${heading} #${prompt.promptNumber}`,
        prompt.content,
        ...(contextLines.length > 0 ? ['', ...contextLines] : []),
    ].join('\n').trim();
}

export function getStructuredUserPrompt(metadata: Record<string, unknown>): StructuredUserPrompt | null {
    const candidate = metadata.structuredUserPrompt;
    if (!candidate || typeof candidate !== 'object') {
        return null;
    }

    const prompt = candidate as Partial<StructuredUserPrompt>;
    if (typeof prompt.content !== 'string' || typeof prompt.contentHash !== 'string' || typeof prompt.recordedAt !== 'number') {
        return null;
    }

    const role = prompt.role === 'goal' || prompt.role === 'objective' || prompt.role === 'prompt'
        ? prompt.role
        : 'prompt';

    return {
        role,
        content: prompt.content,
        promptNumber: typeof prompt.promptNumber === 'number' ? prompt.promptNumber : 1,
        sessionId: typeof prompt.sessionId === 'string' ? prompt.sessionId : undefined,
        activeGoal: typeof prompt.activeGoal === 'string' || prompt.activeGoal === null ? prompt.activeGoal : undefined,
        lastObjective: typeof prompt.lastObjective === 'string' || prompt.lastObjective === null ? prompt.lastObjective : undefined,
        contentHash: prompt.contentHash,
        recordedAt: prompt.recordedAt,
    };
}