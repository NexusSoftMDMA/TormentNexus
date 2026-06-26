import type { ReactNode } from "react";

/**
 * Canonical page-level heading used across the cockpit.
 *
 * Replaces six hand-rolled variants where the eyebrow / title / subtitle
 * shapes drifted (text-xl vs text-2xl vs sm:text-3xl, h1 vs h2, sometimes
 * no eyebrow). One place to tune, one consistent rhythm everywhere.
 *
 * Layout: eyebrow on top in muted uppercase, title underneath at text-2xl
 * semibold, optional subtitle as text-sm muted, optional action slot
 * (typically a primary button) right-aligned on sm+ screens.
 */
export function PageHeader({
  eyebrow,
  title,
  subtitle,
  action,
}: {
  eyebrow: string;
  title: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <header className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">
          {eyebrow}
        </p>
        <h1 className="mt-1 text-2xl font-semibold tracking-tight">{title}</h1>
        {subtitle && (
          <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">
            {subtitle}
          </p>
        )}
      </div>
      {action && <div className="shrink-0 self-start sm:self-auto">{action}</div>}
    </header>
  );
}
