export interface ShellHistoryEntryRow {
    id: string;
    cwd: string;
    command: string;
    duration: number | null;
    exitCode: number | null;
    outputSnippet: string;
}

function isObject(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

export function normalizeShellHistory(payload: unknown): ShellHistoryEntryRow[] {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.reduce<ShellHistoryEntryRow[]>((acc, item, index) => {
        if (!isObject(item)) {
            return acc;
        }

        const rawId = typeof item.id === 'string' ? item.id.trim() : '';
        const rawCwd = typeof item.cwd === 'string' ? item.cwd.trim() : '';
        const rawCommand = typeof item.command === 'string' ? item.command.trim() : '';
        const rawOutputSnippet = typeof item.outputSnippet === 'string' ? item.outputSnippet : '';

        acc.push({
            id: rawId.length > 0 ? rawId : `history-${index}`,
            cwd: rawCwd.length > 0 ? rawCwd : '~',
            command: rawCommand.length > 0 ? rawCommand : '(no command)',
            duration: typeof item.duration === 'number' && Number.isFinite(item.duration) ? item.duration : null,
            exitCode: typeof item.exitCode === 'number' && Number.isFinite(item.exitCode) ? item.exitCode : null,
            outputSnippet: rawOutputSnippet,
        });

        return acc;
    }, []);
}