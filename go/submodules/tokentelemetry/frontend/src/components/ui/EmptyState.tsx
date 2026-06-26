import * as React from "react";
import { cn } from "@/lib/cn";

export interface EmptyStateProps {
  icon?: React.ReactNode;
  title: React.ReactNode;
  description?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn("flex flex-col items-center text-center gap-3 py-16 px-6", className)}>
      {icon && (
        <div className="h-12 w-12 rounded-[var(--tt-radius)] tt-tint-2 border border-[var(--tt-border)] grid place-items-center text-[var(--tt-fg-dim)]">
          {icon}
        </div>
      )}
      <div className="text-[var(--tt-fg)] text-[15px] font-semibold tracking-tight">{title}</div>
      {description && (
        <div className="max-w-md text-[13px] text-[var(--tt-fg-muted)] leading-relaxed">
          {description}
        </div>
      )}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}
