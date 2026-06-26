// Thin GA4 event wrapper. `window.gtag` only exists after the visitor accepts
// analytics (see Analytics.tsx) and never on localhost, so the gtag call is a
// no-op until consent is granted — no extra guarding needed at call sites.
//
// On localhost we instead print the event to the console (`[analytics] <name>`
// with its params) so you can click around in local dev and SEE exactly what
// would be sent, without polluting real GA data. Set `window.__ttAnalyticsDebug
// = false` in the console to silence it.
//
// Mark `copy_install_command` and `click_github` (and `click_install`) as key
// events in the GA4 UI to turn them into conversions.

type Params = Record<string, string | number | boolean>;

declare global {
  interface Window {
    gtag?: (command: string, eventName: string, params?: Params) => void;
    __ttAnalyticsDebug?: boolean;
  }
}

function isLocalhost(): boolean {
  if (typeof window === "undefined") return false;
  const h = window.location.hostname;
  return h === "localhost" || h === "127.0.0.1" || h === "0.0.0.0";
}

export function track(event: string, params?: Params): void {
  if (typeof window === "undefined") return;
  if (isLocalhost()) {
    // Local dev: log instead of sending (GA never loads on localhost anyway).
    if (window.__ttAnalyticsDebug !== false) {
      // eslint-disable-next-line no-console
      console.log(`%c[analytics]%c ${event}`, "color:#60a5fa;font-weight:600", "color:inherit", params ?? {});
    }
    return;
  }
  window.gtag?.("event", event, params);
}
