export function sqliteErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export function isSqliteUnavailableError(error: unknown): boolean {
    const message = sqliteErrorMessage(error);
    return message.includes('SQLite runtime is unavailable')
        || message.includes('SQLite runtime is unavailable for TormentNexus DB-backed features')
        || (
            message.includes('Could not locate the bindings file')
            && message.includes('better-sqlite3')
        )
        || message.includes('ERR_DLOPEN_FAILED');
}

export function formatOptionalSqliteFailure(action: string, error: unknown): string {
    if (isSqliteUnavailableError(error)) {
        return `${action}: SQLite runtime is unavailable for this run.`;
    }

    return `${action}: ${sqliteErrorMessage(error)}`;
}
