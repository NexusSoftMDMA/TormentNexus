/**
 * localStorage-backed sidebar-collapsed state.
 *
 * Pure helpers for read/write so we can unit-test the boolean
 * round-trip without standing up a DOM/jsdom. The component-side hook
 * (`useSidebarCollapsed`) wraps these with `useSyncExternalStore` so
 * the React Compiler's set-state-in-effect rule passes — instead of a
 * `useEffect → setState` hydration we let React subscribe to the
 * storage source directly.
 */

export const SIDEBAR_COLLAPSED_KEY = 'chorus.sidebar.collapsed';

/**
 * Read the persisted collapsed flag. Returns `false` if storage is
 * unavailable (SSR, private mode, quota errors) so the default UX is
 * "expanded" rather than collapsed-by-mistake.
 */
export function readCollapsed(
  storage: Pick<Storage, 'getItem'> | null | undefined,
): boolean {
  if (!storage) return false;
  try {
    return storage.getItem(SIDEBAR_COLLAPSED_KEY) === '1';
  } catch {
    return false;
  }
}

/**
 * Persist the collapsed flag. Swallows storage errors silently —
 * users in private mode or with quota issues just lose the preference,
 * the toggle still works in-session.
 */
export function writeCollapsed(
  storage: Pick<Storage, 'setItem'> | null | undefined,
  collapsed: boolean,
): void {
  if (!storage) return;
  try {
    storage.setItem(SIDEBAR_COLLAPSED_KEY, collapsed ? '1' : '0');
  } catch {
    /* private mode / quota / disabled storage — ignore */
  }
}
