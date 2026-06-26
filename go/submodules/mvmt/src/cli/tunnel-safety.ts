import { MvmtConfig } from '../config/schema.js';

export const LEGACY_TUNNEL_OVERRIDE_ENV = 'MVMT_ALLOW_LEGACY_TUNNEL';

export function tunnelLegacyAccessWarning(config: MvmtConfig, env: NodeJS.ProcessEnv = process.env): string | undefined {
  if (config.server.access !== 'tunnel') return undefined;
  if ((config.clients?.length ?? 0) > 0) return undefined;
  if (legacyTunnelOverrideEnabled(env)) return undefined;
  return [
    'No API tokens are configured. The tunnel can start, but MCP data access will reject the legacy session token.',
    'Run `mvmt token add` to grant scoped access.',
  ].join(' ');
}

export function legacyTunnelOverrideEnabled(env: NodeJS.ProcessEnv = process.env): boolean {
  const value = env[LEGACY_TUNNEL_OVERRIDE_ENV]?.trim().toLowerCase();
  return value === '1' || value === 'true' || value === 'yes';
}
