export interface SessionBootstrapObservation {
    title?: string;
    narrative?: string;
    type?: string;
    toolName?: string;
    content?: string;
}

export interface SessionBootstrapSummary {
    content?: string;
}

export interface SessionBootstrapPayload {
    goal?: string | null;
    objective?: string | null;
    summaryCount: number;
    observationCount: number;
    toolAdvertisementCount: number;
    prompt: string;
}

function trimLine(value: string, maxLength: number): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function pickObservationLine(observation: SessionBootstrapObservation): string {
    const title = typeof observation.title === 'string' ? trimLine(observation.title, 140) : '';
    const narrative = typeof observation.narrative === 'string' ? trimLine(observation.narrative, 220) : '';
    const content = typeof observation.content === 'string' ? trimLine(observation.content, 220) : '';
    const type = typeof observation.type === 'string' ? trimLine(observation.type, 40) : '';
    const toolName = typeof observation.toolName === 'string' ? trimLine(observation.toolName, 60) : '';

    const body = title || narrative || content || 'Recent observation available';
    const tags = [type, toolName].filter(Boolean).join(' · ');
    return tags ? `${body} (${tags})` : body;
}

function pickSummaryLine(summary: SessionBootstrapSummary): string {
    return trimLine(summary.content ?? 'Recent session summary available', 240);
}

export function buildSessionBootstrapPrompt(input: {
    activeGoal?: string | null;
    lastObjective?: string | null;
    summaries: SessionBootstrapSummary[];
    observations: SessionBootstrapObservation[];
    toolAdvertisementLines?: string[];
}): SessionBootstrapPayload {
    const summaryLines = input.summaries.slice(0, 3).map((summary) => `- ${pickSummaryLine(summary)}`);
    const observationLines = input.observations.slice(0, 5).map((observation) => `- ${pickObservationLine(observation)}`);
    const toolAdvertisementLines = (input.toolAdvertisementLines ?? []).slice(0, 8).map((line) => `- ${trimLine(line, 220)}`);

    const sections = [
        'Memory bootstrap:',
        input.activeGoal ? `Current goal: ${trimLine(input.activeGoal, 180)}` : null,
        input.lastObjective ? `Last objective: ${trimLine(input.lastObjective, 180)}` : null,
        toolAdvertisementLines.length > 0 ? 'Suggested tools for the current topic:' : null,
        ...(toolAdvertisementLines.length > 0 ? toolAdvertisementLines : []),
        summaryLines.length > 0 ? 'Recent session summaries:' : null,
        ...(summaryLines.length > 0 ? summaryLines : []),
        observationLines.length > 0 ? 'Relevant observations:' : null,
        ...(observationLines.length > 0 ? observationLines : []),
    ].filter((line): line is string => Boolean(line));

    return {
        goal: input.activeGoal ?? null,
        objective: input.lastObjective ?? null,
        summaryCount: input.summaries.length,
        observationCount: input.observations.length,
        toolAdvertisementCount: toolAdvertisementLines.length,
        prompt: sections.join('\n'),
    };
}
