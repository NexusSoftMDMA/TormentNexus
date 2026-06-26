"use client";

import { AlertTriangle, CheckCircle2 } from "lucide-react";
import type { TemplateValidationIssue } from "@/lib/template-validation";

export function YamlPanel({
  yaml,
  filename,
  issues,
  onChange,
}: {
  yaml: string;
  filename: string;
  issues: TemplateValidationIssue[];
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex h-full min-h-[400px] flex-col">
      <div className="flex items-center justify-between border-b border-border bg-card/40 px-6 py-2.5">
        <span className="font-mono text-[11px] text-muted-foreground">
          {filename}
        </span>
        <ValidationBadge issues={issues} />
      </div>
      <textarea
        value={yaml}
        onChange={(e) => onChange(e.target.value)}
        spellCheck={false}
        className="flex-1 resize-none border-0 bg-background px-6 py-4 font-mono text-[12px] leading-relaxed text-foreground focus:outline-none"
      />
      {issues.length > 0 && (
        <ul className="border-t border-border bg-destructive/5 px-6 py-2.5 text-[11px] text-destructive">
          {issues.map((i, idx) => (
            <li key={idx} className="flex items-start gap-1.5 leading-snug">
              <AlertTriangle className="mt-[1px] h-3 w-3 shrink-0" />
              <span>
                <span className="font-mono">{i.path}</span>
                {typeof i.line === "number" && ` (line ${i.line})`}: {i.message}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function ValidationBadge({ issues }: { issues: TemplateValidationIssue[] }) {
  if (issues.length === 0) {
    return (
      <span className="flex items-center gap-1 rounded-md bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-medium text-emerald-400">
        <CheckCircle2 className="h-3 w-3" />
        Valid
      </span>
    );
  }
  return (
    <span className="flex items-center gap-1 rounded-md bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
      <AlertTriangle className="h-3 w-3" />
      {issues.length} {issues.length === 1 ? "issue" : "issues"}
    </span>
  );
}
