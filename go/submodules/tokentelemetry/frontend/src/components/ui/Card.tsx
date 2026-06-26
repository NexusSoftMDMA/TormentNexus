import * as React from "react";
import { cn } from "@/lib/cn";

type Tone = "default" | "raised" | "sunken" | "glow";

const TONE: Record<Tone, string> = {
  default: "bg-[var(--tt-panel)] border-[var(--tt-border)]",
  raised:  "bg-[var(--tt-raised)] border-[var(--tt-border-strong)]",
  sunken:  "bg-[var(--tt-sunken)] border-[var(--tt-border)]",
  glow:    "bg-gradient-to-br from-[var(--tt-panel)] to-[var(--tt-raised)] border-[var(--tt-border-strong)] shadow-[0_0_0_1px_rgba(96,165,250,0.05),0_24px_60px_-30px_rgba(96,165,250,0.25)]",
};

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  tone?: Tone;
  interactive?: boolean;
  padding?: "none" | "sm" | "md" | "lg";
}

const PAD = { none: "", sm: "p-4", md: "p-5", lg: "p-6" } as const;

export const Card = React.forwardRef<HTMLDivElement, CardProps>(function Card(
  { tone = "default", interactive, padding = "md", className, ...rest }, ref,
) {
  return (
    <div
      ref={ref}
      className={cn(
        "rounded-[var(--tt-radius-lg)] border backdrop-blur-[1px] transition-colors",
        TONE[tone],
        PAD[padding],
        interactive && "hover:border-[var(--tt-border-strong)] hover:bg-[var(--tt-raised)] cursor-pointer",
        className,
      )}
      {...rest}
    />
  );
});

export function CardHeader({ className, ...rest }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex items-center justify-between gap-3 mb-4", className)} {...rest} />;
}

export function CardTitle({ className, ...rest }: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn("text-[13px] font-semibold text-[var(--tt-fg)] tracking-tight flex items-center gap-2", className)}
      {...rest}
    />
  );
}

export function CardEyebrow({ className, ...rest }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "text-[10px] font-semibold text-[var(--tt-fg-dim)] uppercase tracking-[0.18em]",
        className,
      )}
      {...rest}
    />
  );
}
