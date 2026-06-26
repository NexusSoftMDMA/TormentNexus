import { z } from 'zod';
import { isApiTokenVerifier, normalizeApiTokenVerifierForDuplicateCheck } from '../utils/api-token-hash.js';

export const GLOBAL_SECRET_PATH_PATTERNS = [
  '.mvmt/**',
  '.ssh/**',
  '.gnupg/**',
  '.aws/**',
  '.config/gh/**',
  '.config/gcloud/**',
  '.config/azure/**',
  '.kube/**',
  '.docker/**',
  '.docker/config.json',
  '.npmrc',
  '.netrc',
  '.pypirc',
  '.git-credentials',
] as const;

export const DEFAULT_MOUNT_EXCLUDE_PATTERNS = [
  '.git/**',
  '**/.git/**',
  'node_modules/**',
  '**/node_modules/**',
  '.claude/**',
  '**/.claude/**',
  ...GLOBAL_SECRET_PATH_PATTERNS,
] as const;

export const DEFAULT_MOUNT_PROTECT_PATTERNS = [
  '.env',
  '.env.*',
  '.claude/**',
  ...GLOBAL_SECRET_PATH_PATTERNS,
] as const;

export const CommandTunnelSchema = z.object({
  provider: z.enum(['cloudflare-quick', 'pinggy', 'localhost-run', 'custom']),
  command: z.string().min(1),
  url: z.string().url().optional(),
});

export const RelayTunnelSchema = z.object({
  provider: z.literal('relay'),
  relayUrl: z.string().url(),
  workspaceSlug: z.string().min(1).regex(/^[a-z0-9][a-z0-9-]{0,62}$/, 'workspace slug must be lowercase alphanum/dash'),
  agentToken: z.string().min(1),
  url: z.string().url().optional(),
});

export const TunnelSchema = z.discriminatedUnion('provider', [
  CommandTunnelSchema,
  RelayTunnelSchema,
]);

export const ProxySchema = z
  .object({
    // id is the stable raw-connector identifier. When omitted, name is
    // used. Set id explicitly when you need to rename a proxy without
    // changing its virtual policy path or semantic tool source list.
    id: z.string().min(1).optional(),
    name: z.string().min(1),
    source: z.string().optional(), // legacy setup-provenance metadata; runtime ignores it
    transport: z.enum(['stdio', 'http']).default('stdio'),
    command: z.string().min(1).optional(),
    args: z.array(z.string()).default([]),
    url: z.string().min(1).optional(),
    env: z.record(z.string()).default({}),
    writeAccess: z.boolean().optional(),
    enabled: z.boolean().default(true),
  })
  .superRefine((data, ctx) => {
    if (data.transport === 'stdio' && !data.command) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'stdio transport requires "command"',
        path: ['command'],
      });
    }

    if (data.transport === 'http' && !data.url) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'http transport requires "url"',
        path: ['url'],
      });
    }
  });

export const LocalFolderMountSchema = z.object({
  name: z.string().min(1).regex(/^[a-z0-9][a-z0-9_-]*$/, 'mount name must be lowercase alphanum/dash/underscore'),
  type: z.literal('local_folder').default('local_folder'),
  path: z.string().min(1).regex(/^\/(?!$)/, 'mount path must be absolute and cannot be /'),
  root: z.string().min(1),
  description: z.string().default(''),
  guidance: z.string().default(''),
  exclude: z.array(z.string().min(1)).default(() => [...DEFAULT_MOUNT_EXCLUDE_PATTERNS]),
  protect: z.array(z.string().min(1)).default(() => [...DEFAULT_MOUNT_PROTECT_PATTERNS]),
  writeAccess: z.boolean().default(false),
  enabled: z.boolean().default(true),
});

export const PermissionAction = z.enum(['search', 'read', 'write']);

export const PermissionSchema = z.object({
  path: z.string().min(1).regex(/^\/(?!$)/, 'permission path must be absolute and cannot be /'),
  actions: z.array(PermissionAction).min(1, 'permission actions must include at least one action').default([]),
});

export const ClientAuthTokenSchema = z.object({
  type: z.literal('token'),
  // Verifier for an issued bearer token. Plaintext is shown once at issuance
  // and never persisted.
  tokenHash: z.string().refine(isApiTokenVerifier, 'tokenHash must be a scrypt verifier'),
});

export const ClientAuthOAuthSchema = z.object({
  type: z.literal('oauth'),
  oauthClientIds: z.array(z.string().min(1)).default([]),
});

export const ClientAuthSchema = z.discriminatedUnion('type', [
  ClientAuthTokenSchema,
  ClientAuthOAuthSchema,
]);

export const ClientSchema = z.object({
  id: z.string().min(1).regex(/^[a-z0-9][a-z0-9_-]*$/, 'client id must be lowercase alphanum/dash/underscore'),
  name: z.string().min(1),
  description: z.string().default(''),
  createdAt: z.string().datetime().optional(),
  lastUsedAt: z.string().datetime().optional(),
  expiresAt: z.string().datetime().optional(),
  credentialVersion: z.number().int().min(1).optional(),
  clientBinding: z.string().min(1).optional(),
  auth: ClientAuthSchema,
  rawToolsEnabled: z.boolean().default(false),
  permissions: z.array(PermissionSchema).default([]),
  // Exposure boundary for the capability/grant model. A grant reachable
  // over the relay must be explicitly published. Absent means "minted
  // before the publish concept existed" and is grandfathered as
  // published so existing tokens keep working — see isGrantPublished().
  published: z.boolean().optional(),
});

export const SemanticToolEntrySchema = z.object({
  enabled: z.boolean().default(true),
  sourceIds: z.array(z.string().min(1)).default([]),
});

// SemanticToolsSchema is an object keyed by well-known tool names rather
// than an array, so adding a new semantic tool is a schema addition, not
// a structural change. Keys are camelCase; the runtime tool name is the
// snake_case form (search_personal_context, read_context_item).
export const SemanticToolsSchema = z.object({
  searchPersonalContext: SemanticToolEntrySchema.optional(),
  readContextItem: SemanticToolEntrySchema.optional(),
}).default({});

export const ConfigSchema = z
  .object({
    version: z.literal(1),
    server: z
      .object({
        port: z.number().int().min(1).max(65535).default(4141),
        allowedOrigins: z.array(z.string().min(1)).default([]),
        access: z.enum(['local', 'tunnel']).default('local'),
        tunnel: TunnelSchema.optional(),
      })
      .superRefine((data, ctx) => {
        if (data.access === 'tunnel' && !data.tunnel) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: 'tunnel access requires "tunnel" config',
            path: ['tunnel'],
          });
        }
      })
      .default({}),
    proxy: z.array(ProxySchema).default([]),
    mounts: z.array(LocalFolderMountSchema).default([]),
    // clients and semanticTools are additive to the v1 schema. When
    // absent, the runtime synthesizes a single default client that maps
    // to the existing session token (preserves pre-PR behavior).
    clients: z.array(ClientSchema).optional(),
    semanticTools: SemanticToolsSchema.optional(),
  })
  .superRefine((data, ctx) => {
    const knownSourceIds = collectKnownSourceIds(data);
    const knownPolicyRoots = collectKnownPolicyRoots(data);
    validateUniqueSourceIds(data, ctx);
    const seenClientIds = new Set<string>();
    // Track auth bindings across clients so config order does not silently
    // become an authorization decision when two clients share the same
    // OAuth client_id or token hash.
    const seenTokenHashes = new Map<string, number>();
    const seenOauthClientIds = new Map<string, number>();

    for (const [clientIndex, client] of (data.clients ?? []).entries()) {
      if (seenClientIds.has(client.id)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `duplicate client id "${client.id}"`,
          path: ['clients', clientIndex, 'id'],
        });
      } else {
        seenClientIds.add(client.id);
      }

      if (client.auth.type === 'token') {
        const tokenHash = normalizeApiTokenVerifierForDuplicateCheck(client.auth.tokenHash);
        const firstSeen = seenTokenHashes.get(tokenHash);
        if (firstSeen !== undefined) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: `duplicate tokenHash; first seen on clients[${firstSeen}]`,
            path: ['clients', clientIndex, 'auth', 'tokenHash'],
          });
        } else {
          seenTokenHashes.set(tokenHash, clientIndex);
        }
      } else if (client.auth.type === 'oauth') {
        for (const [oauthIndex, oauthClientId] of client.auth.oauthClientIds.entries()) {
          const firstSeen = seenOauthClientIds.get(oauthClientId);
          if (firstSeen !== undefined) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: `oauth client_id "${oauthClientId}" is already mapped on clients[${firstSeen}]`,
              path: ['clients', clientIndex, 'auth', 'oauthClientIds', oauthIndex],
            });
          } else {
            seenOauthClientIds.set(oauthClientId, clientIndex);
          }
        }
      }

      for (const [permIndex, permission] of client.permissions.entries()) {
        validatePermissionPath(
          permission.path,
          knownPolicyRoots,
          ctx,
          ['clients', clientIndex, 'permissions', permIndex, 'path'],
        );
      }
    }

    if (data.semanticTools) {
      for (const [toolKey, tool] of Object.entries(data.semanticTools)) {
        if (!tool) continue;
        for (const [sourceIndex, sourceId] of tool.sourceIds.entries()) {
          if (!knownSourceIds.has(sourceId)) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: `unknown sourceId "${sourceId}"; configure a proxy or mount first`,
              path: ['semanticTools', toolKey, 'sourceIds', sourceIndex],
            });
          }
        }
      }
    }
  });

function validateUniqueSourceIds(data: { proxy: ProxyConfig[]; mounts: LocalFolderMountConfig[] }, ctx: z.RefinementCtx): void {
  const seen = new Map<string, { path: (string | number)[] }>();
  const track = (sourceId: string, issuePath: (string | number)[]) => {
    const previous = seen.get(sourceId);
    if (previous) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: `duplicate sourceId "${sourceId}"; first seen at ${previous.path.join('.')}`,
        path: issuePath,
      });
      return;
    }
    seen.set(sourceId, { path: issuePath });
  };

  for (const [index, proxy] of data.proxy.entries()) {
    track(resolveProxySourceId(proxy), ['proxy', index, proxy.id ? 'id' : 'name']);
  }
  const seenMountPaths = new Map<string, number>();
  for (const [index, mount] of data.mounts.entries()) {
    track(mount.name, ['mounts', index, 'name']);
    validateMountPath(mount.path, ctx, ['mounts', index, 'path']);
    const normalizedMountPath = normalizePolicyPath(mount.path);
    const firstMountPath = seenMountPaths.get(normalizedMountPath);
    if (firstMountPath !== undefined) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: `duplicate mount path "${normalizedMountPath}"; first seen at mounts.${firstMountPath}.path`,
        path: ['mounts', index, 'path'],
      });
    } else {
      seenMountPaths.set(normalizedMountPath, index);
    }
  }
}

function validateMountPath(pathPattern: string, ctx: z.RefinementCtx, issuePath: (string | number)[]): void {
  if (pathPattern.includes('*')) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'mount path must be a literal virtual path and cannot include wildcards',
      path: issuePath,
    });
  }
}

function validatePermissionPath(
  pathPattern: string,
  knownPolicyRoots: string[],
  ctx: z.RefinementCtx,
  issuePath: (string | number)[],
): void {
  if (!hasSupportedPermissionWildcard(pathPattern)) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'permission path may only use /** as a trailing subtree wildcard',
      path: issuePath,
    });
    return;
  }

  if (!permissionTargetsKnownRoot(pathPattern, knownPolicyRoots)) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: `permission path "${pathPattern}" does not target any configured mount or proxy source`,
      path: issuePath,
    });
  }
}

// resolveProxySourceId returns the stable source id for a raw proxy entry
// (id when set, otherwise name). Exported so runtime lookup and schema
// validation use the same identifier.
export function resolveProxySourceId(proxy: ProxyConfig): string {
  return proxy.id ?? proxy.name;
}

function collectKnownSourceIds(data: { proxy: ProxyConfig[]; mounts: LocalFolderMountConfig[] }): Set<string> {
  const ids = new Set<string>();
  for (const proxy of data.proxy) {
    ids.add(resolveProxySourceId(proxy));
  }
  for (const mount of data.mounts) {
    ids.add(mount.name);
  }
  return ids;
}

function collectKnownPolicyRoots(data: { proxy: ProxyConfig[]; mounts: LocalFolderMountConfig[] }): string[] {
  const roots = new Set<string>();
  for (const proxy of data.proxy) {
    roots.add(normalizePolicyPath(`/${resolveProxySourceId(proxy)}`));
  }
  for (const mount of data.mounts) {
    roots.add(normalizePolicyPath(mount.path));
  }
  return [...roots];
}

function hasSupportedPermissionWildcard(pathPattern: string): boolean {
  if (pathPattern === '/**') return true;
  if (!pathPattern.includes('*')) return true;
  return pathPattern.endsWith('/**') && !pathPattern.slice(0, -3).includes('*');
}

function permissionTargetsKnownRoot(pathPattern: string, knownPolicyRoots: string[]): boolean {
  if (pathPattern === '/**') return true;
  const base = pathPattern.endsWith('/**')
    ? normalizePolicyPath(pathPattern.slice(0, -3))
    : normalizePolicyPath(pathPattern);
  return knownPolicyRoots.some((root) => base === root || base.startsWith(`${root}/`));
}

function normalizePolicyPath(value: string): string {
  const withForwardSlashes = value.replaceAll('\\', '/');
  let end = withForwardSlashes.length;
  while (end > 0 && withForwardSlashes[end - 1] === '/') end -= 1;
  const normalized = withForwardSlashes.slice(0, end);
  return normalized || '/';
}

export type TunnelConfig = z.infer<typeof TunnelSchema>;
export type ProxyConfig = z.infer<typeof ProxySchema>;
export type LocalFolderMountConfig = z.infer<typeof LocalFolderMountSchema>;
export type PermissionConfig = z.infer<typeof PermissionSchema>;
export type ClientAuthConfig = z.infer<typeof ClientAuthSchema>;
export type ClientConfig = z.infer<typeof ClientSchema>;
export type SemanticToolEntryConfig = z.infer<typeof SemanticToolEntrySchema>;
export type SemanticToolsConfig = z.infer<typeof SemanticToolsSchema>;
export type MvmtConfig = z.infer<typeof ConfigSchema>;
