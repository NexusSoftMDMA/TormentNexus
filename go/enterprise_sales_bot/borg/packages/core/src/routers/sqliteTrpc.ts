import { TRPCError } from '@trpc/server';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../db/sqliteAvailability.js';

export function rethrowSqliteUnavailableAsTrpc(action: string, error: unknown): never {
    if (isSqliteUnavailableError(error)) {
        throw new TRPCError({
            code: 'INTERNAL_SERVER_ERROR',
            message: formatOptionalSqliteFailure(action, error),
        });
    }

    throw error;
}
