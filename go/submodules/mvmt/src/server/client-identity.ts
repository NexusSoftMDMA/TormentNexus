import type { Request } from 'express';
import { ClientIdentity, isClientIdentity } from '../core/auth/client-identity.js';

// The pure auth-state types and resolution logic live in core/auth.
// Express-bound helpers stay here in server/ because they reach into
// `req.res.locals` — an HTTP-shaped concern that has no place in core.
export * from '../core/auth/client-identity.js';

export function readClientIdentity(req: Request): ClientIdentity | undefined {
  const value = req.res?.locals?.mvmtClient;
  return isClientIdentity(value) ? value : undefined;
}

export function attachClientIdentity(req: Request, identity: ClientIdentity): void {
  if (req.res) req.res.locals.mvmtClient = identity;
}
