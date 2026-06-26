import { fileInspectorApp } from './file-inspector/manifest.js';
import { photosApp } from './photos/manifest.js';

// AppManifest is the minimal first-party-app contract. Each manifest carries
// a self-contained HTML payload (same inline pattern as DASHBOARD_PAGE_HTML)
// that the server returns from a single launch route at GET /apps/<id>.
// This is a registry + launch route, not full static asset hosting — there
// is deliberately no GET /apps/<id>/* surface for CSS/JS/image assets. When
// an app outgrows inline HTML, that surface is the next step. Third-party
// install, dynamic discovery, per-app permissions, and disk-served bundles
// are out of scope here — first-party only.
export interface AppManifest {
  id: string;
  label: string;
  description: string;
  // Self-contained HTML page; inlined CSS+JS allowed. The page may call
  // /api/fs/* and other cookie-authenticated routes.
  html: string;
}

export const INSTALLED_APPS: AppManifest[] = [fileInspectorApp, photosApp];

export function getApp(id: string): AppManifest | undefined {
  return INSTALLED_APPS.find((app) => app.id === id);
}
