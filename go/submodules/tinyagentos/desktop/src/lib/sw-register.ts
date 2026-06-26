/**
 * Register the SPA's service worker once on boot.
 * Safe to call from a useEffect — does nothing on browsers without
 * service worker support and never throws.
 */
export async function registerServiceWorker(): Promise<void> {
  if (typeof navigator === "undefined" || !navigator.serviceWorker) return;
  try {
    await navigator.serviceWorker.register("/sw.js");
  } catch (err) {
    console.warn("[taos] service worker registration failed:", err);
  }
}
