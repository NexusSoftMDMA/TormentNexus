"use client";

import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/cn";
import { useTheme } from "./ThemeProvider";

export function ThemeToggle({ collapsed }: { collapsed?: boolean }) {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  return (
    <button
      onClick={toggleTheme}
      role="switch"
      suppressHydrationWarning
      aria-checked={!isDark}
      aria-label={`Switch to ${isDark ? "light" : "dark"} mode`}
      title={`Switch to ${isDark ? "light" : "dark"} mode`}
      className={cn(
        "w-full flex items-center rounded-[var(--tt-radius)] border border-transparent transition-colors",
        "text-[var(--tt-fg-dim)] hover:text-[var(--tt-fg)] hover:border-[var(--tt-border)] hover:tt-tint-1",
        collapsed ? "justify-center h-9" : "justify-between gap-2 px-2 h-9",
      )}
    >
      {collapsed ? (
        isDark ? <Moon size={15} /> : <Sun size={15} />
      ) : (
        <>
          <span className="flex items-center gap-2 text-[10px] uppercase tracking-[0.18em]">
            {isDark ? <Moon size={13} /> : <Sun size={13} />}
            {isDark ? "Dark" : "Light"}
          </span>
          {/* Mini track */}
          <span className="relative inline-flex h-4 w-7 items-center rounded-full border border-[var(--tt-border)] tt-tint-1">
            <span
              className={cn(
                "absolute h-3 w-3 rounded-full bg-[var(--tt-fg-muted)] transition-transform",
                isDark ? "translate-x-0.5" : "translate-x-3.5 bg-[var(--tt-brand)]",
              )}
            />
          </span>
        </>
      )}
    </button>
  );
}
