import { ClientConfig } from '../config/schema.js';
import { LeaseRecord, leaseAllows, leaseResources } from '../core/leases/store.js';

// A Grant is the unified read-model for a scoped capability, independent
// of where it is stored (clients[] in config, or the lease store) and of
// which surface consumes it (/mcp, the lease browser, raw HTTP).
//
// It is a pure projection: it owns no storage and mutates nothing. The
// adapters below turn the two existing record shapes into one Grant shape
// so auth resolution, exposure checks, and audit can work against a
// single type. Stores stay split; only the read-model is unified.

export type GrantAction = 'search' | 'read' | 'write';

export interface GrantScopeEntry {
  // Absolute virtual path or glob, in the same form the permission
  // matcher uses for API-token grants, e.g. /workspace/** or /notes/x.md.
  path: string;
  actions: GrantAction[];
}

// Where the underlying record lives. Drives nothing in the model itself;
// callers use it for audit labelling and surface-specific behaviour.
export type GrantKind = 'token' | 'oauth' | 'lease';

export interface Grant {
  // Stable id of the underlying record: ClientConfig.id or LeaseRecord.id.
  id: string;
  // Human label: ClientConfig.name or LeaseRecord.label.
  label: string;
  kind: GrantKind;
  scope: GrantScopeEntry[];
  // Whether the grant is reachable over the relay. See isGrantPublished()
  // for the grandfather rule applied to records minted before the
  // publish concept existed.
  published: boolean;
  expiresAt?: string;
  revokedAt?: string;
}

// Grandfather rule: a record without an explicit `published` value
// predates the exposure boundary and is treated as published, so every
// pre-existing API token and lease keeps working over the relay. Only a
// record explicitly minted with `published: false` is capability-only.
export function isGrantPublished(value: boolean | undefined): boolean {
  return value !== false;
}

export function clientConfigToGrant(client: ClientConfig): Grant {
  return {
    id: client.id,
    label: client.name,
    kind: client.auth.type === 'oauth' ? 'oauth' : 'token',
    scope: client.permissions.map((permission) => ({
      path: permission.path,
      actions: [...permission.actions],
    })),
    published: isGrantPublished(client.published),
    ...(client.expiresAt ? { expiresAt: client.expiresAt } : {}),
  };
}

export function leaseRecordToGrant(lease: LeaseRecord): Grant {
  return {
    id: lease.id,
    label: lease.label,
    kind: 'lease',
    scope: leaseGrantScope(lease),
    published: isGrantPublished(lease.published),
    ...(lease.expiresAt ? { expiresAt: lease.expiresAt } : {}),
    ...(lease.revokedAt ? { revokedAt: lease.revokedAt } : {}),
  };
}

// A lease's capability scope, expressed in the same path-glob + actions
// form the permission matcher uses for API-token grants. A folder
// resource expands to a /** subtree; a file resource maps to its exact
// source path. Read access also grants `search` (the index-backed read
// tool). An upload-only lease (write without read) carries no
// MCP-resolvable scope — its writes go through the lease browser's
// upload path, not the permission matcher — so it projects to an empty
// scope here, matching the long-standing /mcp behaviour for leases.
export function leaseGrantScope(lease: LeaseRecord): GrantScopeEntry[] {
  if (!leaseAllows(lease, 'read')) return [];
  const actions: GrantAction[] = leaseAllows(lease, 'write')
    ? ['search', 'read', 'write']
    : ['search', 'read'];
  return leaseResources(lease).map((resource) => ({
    path: resource.type === 'folder'
      ? `${stripTrailingSlashes(resource.sourcePath)}/**`
      : resource.sourcePath,
    actions: [...actions],
  }));
}

function stripTrailingSlashes(value: string): string {
  let end = value.length;
  while (end > 0 && value[end - 1] === '/') end -= 1;
  return value.slice(0, end);
}
