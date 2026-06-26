"use client";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Code2, Eye } from "lucide-react";
import { TOOLS, type AutoApprove } from "./primitives";

export interface YamlInputs {
  maxConcurrent: number;
  driverPolicies: Record<string, AutoApprove>;
  reviewerPolicies: Record<string, AutoApprove>;
  allowedDirs: string[];
  defaultDriver: string;
  defaultCostCap: string;
  defaultThreshold: string;
  mcpEnabled: boolean;
  webhookUrl: string;
}

export function buildYaml(s: YamlInputs): string {
  const dirs = s.allowedDirs.map((d) => `  - ${d}`).join("\n");
  const driverP = TOOLS.map(
    (t) => `    ${t.id.replace("-", "_")}: ${s.driverPolicies[t.id]}`,
  ).join("\n");
  const reviewerP = TOOLS.map(
    (t) => `    ${t.id.replace("-", "_")}: ${s.reviewerPolicies[t.id]}`,
  ).join("\n");

  return `# Chorus workspace settings
# Edit here, or via the Form tab, or via mm.update_settings from any MCP client.

concurrency:
  max_concurrent_chats: ${s.maxConcurrent}

permissions:
  driver:
${driverP}
  reviewer:
${reviewerP}

allowed_directories:
${dirs}

defaults:
  driver: ${s.defaultDriver}
  agreement_threshold: ${s.defaultThreshold}
  cost_cap_usd: ${s.defaultCostCap}
  on_error: fallback

notifications:
  mcp_enabled: ${s.mcpEnabled}
  webhook_url: ${s.webhookUrl ? `"${s.webhookUrl}"` : '""'}
`;
}

export function YamlEditor({ yaml }: { yaml: string }) {
  return (
    <Card className="overflow-hidden bg-card p-0">
      <div className="flex items-center justify-between border-b border-border px-4 py-2.5">
        <div className="flex items-center gap-2">
          <Code2 className="h-3.5 w-3.5 text-primary" />
          <span className="text-[11px] font-medium uppercase tracking-wider text-foreground">
            settings.yaml
          </span>
          <Badge
            variant="outline"
            className="border-border font-mono text-[10px]"
          >
            ~/.chorus/settings.yaml
          </Badge>
        </div>
        <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
          <Eye className="h-3 w-3" />
          live-rendered from form state
        </div>
      </div>
      <textarea
        value={yaml}
        readOnly
        spellCheck={false}
        className="block min-h-[520px] w-full resize-none border-0 bg-background px-5 py-4 font-mono text-[12px] leading-relaxed text-foreground/90 focus:outline-none"
      />
      <div className="flex items-center justify-between border-t border-border px-4 py-2.5 text-[11px] text-muted-foreground">
        <span>
          Same shape that{" "}
          <span className="font-mono text-foreground/80">
            mm.update_settings
          </span>{" "}
          accepts as a patch.
        </span>
        <button
          type="button"
          className="rounded-md border border-border bg-card px-2 py-1 text-[10px] text-muted-foreground transition hover:text-foreground"
        >
          Copy
        </button>
      </div>
    </Card>
  );
}
