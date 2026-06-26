/**
 * Presentation metadata for optional, Store-installable frontend apps. The
 * registry (app-registry.ts) owns the runtime manifest (component, sizing);
 * this owns how they look in the Store's "taOS Apps" section. ids must match
 * the registry entries marked `optional`.
 *
 * The platform social apps (Reddit / YouTube / GitHub / X) were DE-SEEDED from
 * the default Store: they are unfinished and now live as the operator's private
 * App Studio drafts, to be finished + published on stream rather than offered to
 * every user. Their registry manifests + components remain for that reseed.
 */

export interface OptionalAppMeta {
  /** Registry app id — must match an `optional: true` manifest. */
  id: string;
  /** Display name. */
  name: string;
  /** lucide-react icon name (rendered via the shared Icon component). */
  icon: string;
  /** One-line description for the Store card. */
  tagline: string;
  /** Brand-tinted CSS background for the card cover. */
  cover: string;
}

// Empty after the social-app de-seed. New optional Store apps get added here.
export const OPTIONAL_APPS: OptionalAppMeta[] = [];
