import { MvmtConfig } from '../config/schema.js';
import { Connector } from '../connectors/types.js';
import { Logger } from '../utils/logger.js';

export type LoadedConnector = {
  connector: Connector;
  sourceId: string;
  toolCount: number;
};

export async function initializeConnectors(
  config: MvmtConfig,
  stdioMode: boolean,
  logger: Logger,
): Promise<LoadedConnector[]> {
  if (config.proxy.some((proxyConfig) => proxyConfig.enabled !== false)) {
    emit('Legacy proxy connectors are ignored by the mount-only runtime.', stdioMode, logger, 'warn');
  }
  return [];
}

export function formatConnectorError(err: unknown): string {
  const message = err instanceof Error ? err.message : 'Unknown error';

  try {
    const parsed = JSON.parse(message) as unknown;
    if (Array.isArray(parsed)) {
      const paths = parsed
        .map((issue) => {
          if (!issue || typeof issue !== 'object' || !Array.isArray((issue as { path?: unknown }).path)) {
            return undefined;
          }
          return (issue as { path: Array<string | number> }).path.join('.');
        })
        .filter((path): path is string => Boolean(path));

      if (paths.length > 0) {
        return `upstream server returned invalid MCP schema at ${paths.join(', ')}`;
      }
    }
  } catch {
    // Not a JSON-formatted validation error.
  }

  return message;
}

function emit(
  message: string,
  stdioMode: boolean,
  logger: Logger,
  level: 'info' | 'warn' | 'error' = 'info',
): void {
  if (stdioMode) {
    process.stderr.write(`${message}\n`);
    return;
  }

  if (level === 'warn') logger.warn(message);
  else if (level === 'error') logger.error(message);
  else logger.info(message);
}
