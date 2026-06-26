export interface EndpointOptions {
  envUrl?: string | null;
  defaultPort: number;
  defaultPath?: string;
}

function normalizeEnvUrl(url: string, suffix?: string): string {
  const trimmed = url.trim().replace(/\/$/, '');
  if (!suffix) {
    return trimmed;
  }

  if (trimmed.endsWith(suffix)) {
    return trimmed;
  }

  return `${trimmed}${suffix}`;
}

export function resolveHttpBaseUrl({ envUrl, defaultPort, defaultPath }: EndpointOptions): string {
  const fromEnv = envUrl?.trim();
  if (fromEnv) {
    return normalizeEnvUrl(fromEnv, defaultPath);
  }

  if (typeof window !== 'undefined') {
    const base = `${window.location.protocol}//${window.location.hostname}:${defaultPort}`;
    return defaultPath ? `${base}${defaultPath}` : base;
  }

  const base = `http://localhost:${defaultPort}`;
  return defaultPath ? `${base}${defaultPath}` : base;
}

export function resolveWsUrl({ envUrl, defaultPort, defaultPath }: EndpointOptions): string {
  const fromEnv = envUrl?.trim();
  if (fromEnv) {
    return normalizeEnvUrl(fromEnv, defaultPath);
  }

  if (typeof window !== 'undefined') {
    const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const base = `${scheme}://${window.location.hostname}:${defaultPort}`;
    return defaultPath ? `${base}${defaultPath}` : base;
  }

  const base = `ws://localhost:${defaultPort}`;
  return defaultPath ? `${base}${defaultPath}` : base;
}

export function resolveTrpcHttpUrl(envUrl?: string | null): string {
  const fromEnv = envUrl?.trim();
  if (fromEnv) {
    return normalizeEnvUrl(fromEnv, '/trpc');
  }

  if (typeof window !== 'undefined') {
    return `${window.location.origin}/api/trpc`;
  }

  return 'http://localhost:3000/api/trpc';
}

export function resolveCoreWsUrl(envUrl?: string | null): string {
  return resolveWsUrl({ envUrl, defaultPort: 3001 });
}

export function resolveCouncilWsUrl(envUrl?: string | null): string {
  return resolveWsUrl({ envUrl, defaultPort: 3000 });
}

export function resolveTerminalWsUrl(envUrl?: string | null): string {
  return resolveWsUrl({ envUrl, defaultPort: 8081 });
}

export function resolveCliApiBaseUrl(envUrl?: string | null): string {
  return resolveHttpBaseUrl({ envUrl, defaultPort: 3001 });
}
