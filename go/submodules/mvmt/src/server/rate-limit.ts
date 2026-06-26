import { rateLimit as expressRateLimit, type Options } from 'express-rate-limit';
import type { RequestHandler } from 'express';

// Thin wrapper around express-rate-limit so CodeQL's js/missing-rate-limiting
// recognizes the middleware (its taint analysis is package-name-sensitive).
// The wrapper also pins our conventions:
//   - key by req.socket.remoteAddress, not req.ip. mvmt never calls
//     `app.set('trust proxy', ...)`, so the two are equivalent here, but
//     keying on the raw socket address makes the intent explicit: a
//     remote client cannot spoof its rate-limit bucket by sending an
//     X-Forwarded-For header through the tunnel.
//   - 429 body is `{ error: 'Too Many Requests' }` to match the rest
//     of our API.
//   - standardHeaders: 'draft-7' so clients see RateLimit-Remaining /
//     RateLimit-Reset. Retry-After is set automatically on 429.

export interface RateLimitOptions {
  windowMs: number;
  max: number;
}

export function rateLimit(options: RateLimitOptions): RequestHandler {
  const config: Partial<Options> = {
    windowMs: options.windowMs,
    limit: options.max,
    standardHeaders: 'draft-7',
    legacyHeaders: false,
    keyGenerator: (req) => req.socket.remoteAddress ?? 'unknown',
    handler: (_req, res) => {
      res.status(429).json({ error: 'Too Many Requests' });
    },
    // Tests create many server instances. Skip express-rate-limit's
    // proxy-trust validation so we don't spam stderr when req.ip and
    // req.socket.remoteAddress differ under test runners.
    validate: false,
  };
  return expressRateLimit(config);
}
