"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";
import { track } from "@/lib/track";

function isLocalhost(): boolean {
  if (typeof window === "undefined") return false;
  const h = window.location.hostname;
  return h === "localhost" || h === "127.0.0.1" || h === "0.0.0.0";
}

/**
 * Sends a GA4 `page_view` on every client-side route change. Next's App Router
 * does SPA navigation, so without this only the first load is ever counted —
 * the entire docs site and /resources would be invisible in GA.
 *
 * In production GA4 already auto-sends the initial page_view (via the
 * `gtag('config', …)` in Analytics.tsx), so we skip the very first render to
 * avoid double counting. On localhost there is no GA, so we log the landing
 * view too for dev visibility (see track()).
 */
export default function PageViewTracker() {
  const pathname = usePathname();
  const firstRender = useRef(true);

  useEffect(() => {
    const isFirst = firstRender.current;
    firstRender.current = false;
    if (isFirst && !isLocalhost()) return;

    track("page_view", {
      page_path: pathname || "/",
      page_title: typeof document !== "undefined" ? document.title : "",
    });
  }, [pathname]);

  return null;
}
