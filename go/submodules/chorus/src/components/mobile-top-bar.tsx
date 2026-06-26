"use client";

import { MobileNav } from "./mobile-nav";

/**
 * Mobile-only top bar — hamburger that opens the sidebar overlay.
 * Hidden on md+ where the desktop sidebar is always visible.
 *
 * Replaces the previous TopBar which also had Search + New Chat. Those
 * controls now live inside the sidebar; on mobile they're reachable through
 * the same MobileNav sheet.
 */
export function MobileTopBar() {
  return (
    <header className="flex h-12 items-center border-b border-border bg-background/60 px-3 backdrop-blur md:hidden">
      <MobileNav />
    </header>
  );
}
